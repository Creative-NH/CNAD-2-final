FROM golang:1.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o frontend_service

EXPOSE 8084

CMD ["./frontend_service"]
