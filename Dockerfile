FROM golang:1.25.0-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY seeds/ ./seeds/

RUN go build -o feature-vote ./cmd/server
RUN go build -o seed ./seeds

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/feature-vote .
COPY --from=builder /app/seed .

EXPOSE 8080

CMD ["./feature-vote"]