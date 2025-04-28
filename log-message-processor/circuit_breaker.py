from enum import Enum
import time
from threading import Lock

class State(Enum):
    CLOSED = "CLOSED"
    HALF_OPEN = "HALF_OPEN"
    OPEN = "OPEN"

class CircuitBreaker:
    def __init__(self, failure_threshold=3, reset_timeout=10):
        self._failure_threshold = failure_threshold
        self._reset_timeout = reset_timeout
        self._state = State.CLOSED
        self._failure_count = 0
        self._last_failure_time = None
        self._half_open_max_calls = 1
        self._half_open_calls = 0
        self._lock = Lock()

    def execute(self, operation):
        if not self._allow_request():
            raise Exception(f"Circuit breaker is {self._state.value}")

        try:
            result = operation()
            self._record_success()
            return result
        except Exception as e:
            self._record_failure()
            raise e

    def _allow_request(self):
        with self._lock:
            if self._state == State.CLOSED:
                return True
            elif self._state == State.OPEN:
                if (self._last_failure_time is None or 
                    time.time() - self._last_failure_time >= self._reset_timeout):
                    self._state = State.HALF_OPEN
                    self._half_open_calls = 0
                    return True
                return False
            elif self._state == State.HALF_OPEN:
                return self._half_open_calls < self._half_open_max_calls
            return False

    def _record_success(self):
        with self._lock:
            if self._state == State.CLOSED:
                self._failure_count = 0
            elif self._state == State.HALF_OPEN:
                self._half_open_calls += 1
                if self._half_open_calls >= self._half_open_max_calls:
                    print("Circuit breaker state changed from HALF_OPEN to CLOSED")
                    self._state = State.CLOSED
                    self._failure_count = 0

    def _record_failure(self):
        with self._lock:
            if self._state == State.CLOSED:
                self._failure_count += 1
                if self._failure_count >= self._failure_threshold:
                    print(f"Circuit breaker state changed from CLOSED to OPEN (failures: {self._failure_count})")
                    self._state = State.OPEN
                    self._last_failure_time = time.time()
            elif self._state == State.HALF_OPEN:
                print("Circuit breaker state changed from HALF_OPEN to OPEN")
                self._state = State.OPEN
                self._last_failure_time = time.time()