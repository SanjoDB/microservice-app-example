package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"
    "errors"
    "sync"

    jwt "github.com/dgrijalva/jwt-go"
    "github.com/labstack/echo"
    "github.com/labstack/echo/middleware"
    gommonlog "github.com/labstack/gommon/log"
)

var (
    // ErrHttpGenericMessage that is returned in general case, details should be logged in such case
    ErrHttpGenericMessage = echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")

    // ErrWrongCredentials indicates that login attempt failed because of incorrect login or password
    ErrWrongCredentials = echo.NewHTTPError(http.StatusUnauthorized, "username or password is invalid")

    jwtSecret = "myfancysecret"
)

func main() {
    hostport := ":" + os.Getenv("AUTH_API_PORT")
    userAPIAddress := os.Getenv("USERS_API_ADDRESS")

    envJwtSecret := os.Getenv("JWT_SECRET")
    if len(envJwtSecret) != 0 {
        jwtSecret = envJwtSecret
    }

    userService := UserService{
        Client:         http.DefaultClient,
        UserAPIAddress: userAPIAddress,
        AllowedUserHashes: map[string]interface{}{
            "admin_admin": nil,
            "johnd_foo":   nil,
            "janed_ddd":   nil,
        },
    }

    e := echo.New()
    e.Logger.SetLevel(gommonlog.INFO)

    if zipkinURL := os.Getenv("ZIPKIN_URL"); len(zipkinURL) != 0 {
        e.Logger.Infof("init tracing to Zipkit at %s", zipkinURL)

        if tracedMiddleware, tracedClient, err := initTracing(zipkinURL); err == nil {
            e.Use(echo.WrapMiddleware(tracedMiddleware))
            userService.Client = tracedClient
        } else {
            e.Logger.Infof("Zipkin tracer init failed: %s", err.Error())
        }
    } else {
        e.Logger.Infof("Zipkin URL was not provided, tracing is not initialised")
    }

    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())

    // Route => handler
    e.GET("/version", func(c echo.Context) error {
        return c.String(http.StatusOK, "Auth API, written in Go\n")
    })

    e.POST("/login", getLoginHandler(userService))

    // Start server
    e.Logger.Fatal(e.Start(hostport))
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func getLoginHandler(userService UserService) echo.HandlerFunc {
    f := func(c echo.Context) error {
        requestData := LoginRequest{}
        decoder := json.NewDecoder(c.Request().Body)
        if err := decoder.Decode(&requestData); err != nil {
            log.Printf("could not read credentials from POST body: %s", err.Error())
            return ErrHttpGenericMessage
        }

        ctx := c.Request().Context()
        user, err := userService.Login(ctx, requestData.Username, requestData.Password)
        if err != nil {
            if err != ErrWrongCredentials {
                log.Printf("could not authorize user '%s': %s", requestData.Username, err.Error())
                return ErrHttpGenericMessage
            }

            return ErrWrongCredentials
        }
        token := jwt.New(jwt.SigningMethodHS256)

        // Set claims
        claims := token.Claims.(jwt.MapClaims)
        claims["username"] = user.Username
        claims["firstname"] = user.FirstName
        claims["lastname"] = user.LastName
        claims["role"] = user.Role
        claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

        // Generate encoded token and send it as response.
        t, err := token.SignedString([]byte(jwtSecret))
        if err != nil {
            log.Printf("could not generate a JWT token: %s", err.Error())
            return ErrHttpGenericMessage
        }

        return c.JSON(http.StatusOK, map[string]string{
            "accessToken": t,
        })
    }

    return echo.HandlerFunc(f)
}

type RetryConfig struct {
    MaxAttempts  int
    WaitTime     time.Duration
    MaxWaitTime  time.Duration
}

func Retry[T any](config RetryConfig, operation func() (T, error)) (T, error) {
    var result T
    var err error
    waitTime := config.WaitTime

    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        result, err = operation()
        if err == nil {
            return result, nil
        }

        if attempt == config.MaxAttempts {
            break
        }

        if waitTime < config.MaxWaitTime {
            waitTime = waitTime * 2
        }
        
        time.Sleep(waitTime)
    }

    return result, err
}

type State int

const (
    StateClosed State = iota
    StateHalfOpen
    StateOpen
)

type CircuitBreaker[T any] struct {
    mutex          sync.RWMutex
    state          State
    failureCount   int
    failureThreshold  int
    resetTimeout   time.Duration
    lastFailureTime time.Time
    halfOpenMaxCalls int
    halfOpenCalls    int
}

func NewCircuitBreaker[T any](failureThreshold int, resetTimeout time.Duration) *CircuitBreaker[T] {
    return &CircuitBreaker[T]{
        state:           StateClosed,
        failureThreshold: failureThreshold,
        resetTimeout:    resetTimeout,
        halfOpenMaxCalls: 1,
    }
}

func (cb *CircuitBreaker[T]) Execute(operation func() (T, error)) (T, error) {
    if !cb.allowRequest() {
        var zero T
        return zero, errors.New("circuit breaker is open")
    }

    result, err := operation()
    cb.recordResult(err)
    return result, err
}

func (cb *CircuitBreaker) allowRequest() bool {
    cb.mutex.RLock()
    defer cb.mutex.RUnlock()

    switch cb.state {
    case StateClosed:
        return true
    case StateOpen:
        if time.Since(cb.lastFailureTime) > cb.resetTimeout {
            cb.mutex.RUnlock()
            cb.mutex.Lock()
            cb.state = StateHalfOpen
            cb.halfOpenCalls = 0
            cb.mutex.Unlock()
            cb.mutex.RLock()
            return true
        }
        return false
    case StateHalfOpen:
        return cb.halfOpenCalls < cb.halfOpenMaxCalls
    default:
        return false
    }
}

func (cb *CircuitBreaker) recordResult(err error) {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()

    switch cb.state {
    case StateClosed:
        if err != nil {
            cb.failureCount++
            if cb.failureCount >= cb.failureThreshold {
                cb.state = StateOpen
                cb.lastFailureTime = time.Now()
            }
        } else {
            cb.failureCount = 0
        }
    case StateHalfOpen:
        cb.halfOpenCalls++
        if err != nil {
            cb.state = StateOpen
            cb.lastFailureTime = time.Now()
        } else if cb.halfOpenCalls >= cb.halfOpenMaxCalls {
            cb.state = StateClosed
            cb.failureCount = 0
        }
    }
}