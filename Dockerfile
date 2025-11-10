FROM golang:1.25-alpine

WORKDIR /app

RUN apk add --no-cache git build-base bash

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

# Pre-cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]


