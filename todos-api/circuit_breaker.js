'use strict';

const States = {
    CLOSED: 'CLOSED',
    HALF_OPEN: 'HALF_OPEN',
    OPEN: 'OPEN'
};

class CircuitBreaker {
    constructor(failureThreshold = 3, resetTimeout = 10000) {
        this.state = States.CLOSED;
        this.failureCount = 0;
        this.failureThreshold = failureThreshold;
        this.resetTimeout = resetTimeout;
        this.lastFailureTime = null;
        this.halfOpenMaxCalls = 1;
        this.halfOpenCalls = 0;
    }

    async execute(operation) {
        if (!this._allowRequest()) {
            throw new Error('Circuit breaker is open');
        }

        try {
            const result = await operation();
            this._recordSuccess();
            return result;
        } catch (error) {
            this._recordFailure();
            throw error;
        }
    }

    _allowRequest() {
        switch (this.state) {
            case States.CLOSED:
                return true;
            case States.OPEN:
                if (Date.now() - this.lastFailureTime >= this.resetTimeout) {
                    this.state = States.HALF_OPEN;
                    this.halfOpenCalls = 0;
                    return true;
                }
                return false;
            case States.HALF_OPEN:
                return this.halfOpenCalls < this.halfOpenMaxCalls;
            default:
                return false;
        }
    }

    _recordSuccess() {
        switch (this.state) {
            case States.CLOSED:
                this.failureCount = 0;
                break;
            case States.HALF_OPEN:
                this.halfOpenCalls++;
                if (this.halfOpenCalls >= this.halfOpenMaxCalls) {
                    this.state = States.CLOSED;
                    this.failureCount = 0;
                }
                break;
        }
    }

    _recordFailure() {
        switch (this.state) {
            case States.CLOSED:
                this.failureCount++;
                if (this.failureCount >= this.failureThreshold) {
                    this.state = States.OPEN;
                    this.lastFailureTime = Date.now();
                }
                break;
            case States.HALF_OPEN:
                this.state = States.OPEN;
                this.lastFailureTime = Date.now();
                break;
        }
    }
}

module.exports = CircuitBreaker;