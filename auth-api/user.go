package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
    "log"

    jwt "github.com/dgrijalva/jwt-go"
)

var (
    userServiceCB = NewCircuitBreaker(3, 10*time.Second)
)

var allowedUserHashes = map[string]interface{}{
	"admin_admin": nil,
	"johnd_foo":   nil,
	"janed_ddd":   nil,
}

type User struct {
	Username  string `json:"username"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Role      string `json:"role"`
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type UserService struct {
	Client            HTTPDoer
	UserAPIAddress    string
	AllowedUserHashes map[string]interface{}
}

func (h *UserService) Login(ctx context.Context, username, password string) (User, error) {
	user, err := h.getUser(ctx, username)
	if err != nil {
		return user, err
	}

	userKey := fmt.Sprintf("%s_%s", username, password)

	if _, ok := h.AllowedUserHashes[userKey]; !ok {
		return user, ErrWrongCredentials // this is BAD, business logic layer must not return HTTP-specific errors
	}

	return user, nil
}

func (h *UserService) getUser(ctx context.Context, username string) (User, error) {
    var user User

    retryConfig := RetryConfig{
        MaxAttempts: 3,
        WaitTime: 100 * time.Millisecond,
        MaxWaitTime: 2 * time.Second,
    }

    return Retry[User](retryConfig, func() (User, error) {
        err := userServiceCB.Execute(func() error {
            token, err := h.getUserAPIToken(username)
            if err != nil {
                log.Printf("Error generando token: %v", err)
                return user, err
            }
            log.Printf("Token generado para usuario %s", username)

            url := fmt.Sprintf("%s/users/%s", h.UserAPIAddress, username)
            log.Printf("Intentando acceder a: %s", url)

            req, _ := http.NewRequest("GET", url, nil)
            req.Header.Add("Authorization", "Bearer "+token)
            req = req.WithContext(ctx)

            resp, err := h.Client.Do(req)
            if err != nil {
                log.Printf("Error en la petición HTTP: %v", err)
                return user, err
            }
            defer resp.Body.Close()

            bodyBytes, err := ioutil.ReadAll(resp.Body)
            if err != nil {
                log.Printf("Error leyendo respuesta: %v", err)
                return user, err
            }

            log.Printf("Respuesta del servidor: %s", string(bodyBytes))

            if resp.StatusCode < 200 || resp.StatusCode >= 300 {
                return user, fmt.Errorf("could not get user data: %s", string(bodyBytes))
            }

            err = json.Unmarshal(bodyBytes, &user)
            return err
        })
        
        return user, err
    })
}

func (h *UserService) getUserAPIToken(username string) (string, error) {
    token := jwt.New(jwt.SigningMethodHS256)
    claims := token.Claims.(jwt.MapClaims)
    claims["username"] = username
    claims["scope"] = "read"
    // Agregar expiración
    claims["exp"] = time.Now().Add(time.Hour * 1).Unix()
    return token.SignedString([]byte(jwtSecret))
}
