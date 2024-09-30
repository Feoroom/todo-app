FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY . .
RUN go build -o ./bin/app ./cmd/api


EXPOSE 8000

CMD ["./bin/app"]