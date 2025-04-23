FROM golang:1.18.2 as build

WORKDIR /app

COPY main.go .
COPY tracing.go .
COPY user.go .
COPY Gopkg.toml .
COPY Gopkg.lock .

RUN go mod init github.com/bortizf/microservice-app-example/tree/master/auth-api
RUN go mod tidy

RUN go build -o auth-api

FROM gcr.io/distroless/base-debian10

WORKDIR /app

COPY --from=build /app/auth-api .

ENV AUTH_API_PORT=8000
ENV USERS_API_ADDRESS=http://127.0.0.1:8083
ENV JWT_SECRET=PRFT

EXPOSE ${AUTH_API_PORT}

CMD ["./auth-api"]