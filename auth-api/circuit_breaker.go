package main

import (
    "errors"
    "sync"
    "time"
)

type State int

const (
    StateClosed State = iota
    StateHalfOpen
    StateOpen
)

type CircuitBreaker[T any] struct {
    mutex           sync.RWMutex
    state           State
    failureCount    int
    failureThreshold int
    resetTimeout    time.Duration
    lastFailureTime time.Time
    halfOpenMaxCalls int
    halfOpenCalls    int
}

func NewCircuitBreaker[T any](failureThreshold int, resetTimeout time.Duration) *CircuitBreaker[T] {
    return &CircuitBreaker[T]{
        state:            StateClosed,
        failureThreshold: failureThreshold,
        resetTimeout:     resetTimeout,
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

func (cb *CircuitBreaker[T]) allowRequest() bool {
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

func (cb *CircuitBreaker[T]) recordResult(err error) {
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