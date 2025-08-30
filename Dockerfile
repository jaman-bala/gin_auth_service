FROM golang:1.24.5-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Устанавливаем swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

# Генерируем Swagger-документацию
RUN /go/bin/swag init -g cmd/server/main.go

# Собираем приложение
RUN go build -o main ./cmd/server

EXPOSE ${SERVER_PORT}

CMD ["./main"]
