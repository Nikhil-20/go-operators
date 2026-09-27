# Dockerfile 

# Stage 1 Build the go application
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the go application 

RUN CGO_ENABLED=0 GOOS=linux go build -o /k8s-watcher ./main.go

FROM alpine:latest

RUN addgroup -S watcher && adduser -S watcher -G watcher

WORKDIR /app

COPY --from=builder --chown=watcher:watcher /k8s-watcher .

USER watcher

CMD ["/app/k8s-watcher"]
