FROM python:3.6

WORKDIR /app

COPY main.py .
COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

ENV REDIS_HOST=127.0.0.1
ENV REDIS_PORT=6379
ENV REDIS_CHANNEL=log_channel

CMD ["python3", "main.py"]