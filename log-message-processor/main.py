import time
import redis
import os
import json
import requests
from py_zipkin.zipkin import zipkin_span, ZipkinAttrs, generate_random_64bit_string
import time
import random
import sys
from retry import retry, RetryConfig
from circuit_breaker import CircuitBreaker

sys.stdout = os.fdopen(sys.stdout.fileno(), 'w', 1)
sys.stderr = os.fdopen(sys.stderr.fileno(), 'w', 1)

# Configuración de Retry y Circuit Breaker
retry_config = RetryConfig(max_attempts=3, wait_time=0.1, max_wait_time=2.0)
redis_cb = CircuitBreaker(failure_threshold=3, reset_timeout=10)
zipkin_cb = CircuitBreaker(failure_threshold=3, reset_timeout=10)
message_cb = CircuitBreaker(failure_threshold=3, reset_timeout=10)

@retry(retry_config)
def log_message(message):
    def _log():
        time_delay = random.randrange(0, 2000)
        time.sleep(time_delay / 1000)
        print('message received after waiting for {}ms: {}'.format(time_delay, message))
    
    return message_cb.execute(_log)

@retry(retry_config)
def connect_redis(host, port):
    def _connect():
        return redis.Redis(host=host, port=port, db=0)
    
    return redis_cb.execute(_connect)

@retry(retry_config)
def subscribe_to_channel(redis_client, channel):
    def _subscribe():
        pubsub = redis_client.pubsub()
        pubsub.subscribe([channel])
        return pubsub
    
    return redis_cb.execute(_subscribe)

@retry(retry_config)
def http_transport(encoded_span, zipkin_url):
    def _transport():
        return requests.post(
            zipkin_url,
            data=encoded_span,
            headers={'Content-Type': 'application/x-thrift'},
        )
    
    return zipkin_cb.execute(_transport)

@retry(retry_config)
def process_message(item):
    def _process():
        if item['type'] == 'message':
            if isinstance(item['data'], bytes):
                return json.loads(item['data'].decode("utf-8"))
            elif isinstance(item['data'], int):
                return {'data': item['data']}
            else:
                return json.loads(str(item['data']))
        return None
    
    return message_cb.execute(_process)

if __name__ == '__main__':
    redis_host = os.environ['REDIS_HOST']
    redis_port = int(os.environ['REDIS_PORT'])
    redis_channel = os.environ['REDIS_CHANNEL']
    zipkin_url = os.environ['ZIPKIN_URL'] if 'ZIPKIN_URL' in os.environ else ''

    def transport_handler(encoded_span):
        try:
            http_transport(encoded_span, zipkin_url)
        except Exception as e:
            print('Error sending span to Zipkin:', e)

    # Establecer conexión a Redis con retry y circuit breaker
    redis_client = connect_redis(redis_host, redis_port)
    pubsub = subscribe_to_channel(redis_client, redis_channel)

    for item in pubsub.listen():
        try:
            if item['type'] == 'message':  # Solo procesar mensajes reales
                message = process_message(item)
                if message:  # Solo procesar si hay mensaje
                    if not zipkin_url or 'zipkinSpan' not in message:
                        log_message(message)
                        continue

                    span_data = message['zipkinSpan']
                    try:
                        with zipkin_span(
                            service_name='log-message-processor',
                            zipkin_attrs=ZipkinAttrs(
                                trace_id=span_data['_traceId']['value'],
                                span_id=generate_random_64bit_string(),
                                parent_span_id=span_data['_spanId'],
                                is_sampled=span_data['_sampled']['value'],
                                flags=None
                            ),
                            span_name='save_log',
                            transport_handler=transport_handler,
                            sample_rate=100
                        ):
                            log_message(message)
                    except Exception as e:
                        print('did not send data to Zipkin: {}'.format(e))
                        log_message(message)
        except Exception as e:
            print(f"Error processing message: {str(e)}")
            try:
                log_message({"error": str(e)})
            except Exception as log_error:
                print(f"Error logging message: {str(log_error)}")
            continue