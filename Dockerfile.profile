FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o profile-service ./cmd/profile

EXPOSE 8002

ENV GIN_MODE=release

CMD ["./profile-service"]