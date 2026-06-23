FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

FROM alpine:3.20

RUN adduser -D -H appuser
USER appuser
WORKDIR /app
COPY --from=builder /api /app/api
COPY --from=builder /worker /app/worker
COPY openapi.yaml /app/openapi.yaml

EXPOSE 8080
ENTRYPOINT ["/app/api"]
