# Estagio 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o http-server-projeto-korp .

# Estagio 2: Runtime enxuto
FROM alpine:3.19

WORKDIR /app

RUN adduser -D -u 1000 appuser && \
    apk --no-cache add ca-certificates tzdata

COPY --from=builder /src/http-server-projeto-korp .

USER appuser

EXPOSE 8080

ENTRYPOINT ["./http-server-projeto-korp"]