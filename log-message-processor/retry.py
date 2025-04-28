import time
from functools import wraps

class RetryConfig:
    def __init__(self, max_attempts=3, wait_time=0.1, max_wait_time=2):
        self.max_attempts = max_attempts
        self.wait_time = wait_time
        self.max_wait_time = max_wait_time

def retry(config):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            wait_time = config.wait_time
            last_exception = None

            for attempt in range(config.max_attempts):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    last_exception = e
                    print(f"Attempt {attempt + 1}/{config.max_attempts} failed: {str(e)}")
                    
                    if attempt == config.max_attempts - 1:
                        raise last_exception

                    if wait_time < config.max_wait_time:
                        wait_time *= 2
                    
                    time.sleep(wait_time)
            
            raise last_exception
        return wrapper
    return decorator