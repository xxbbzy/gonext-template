# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /server ./cmd/server/

# Run stage
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
	&& addgroup -S gonext \
	&& adduser -S -G gonext -h /app gonext

WORKDIR /app

COPY --from=builder --chown=gonext:gonext /server ./server
COPY --from=builder --chown=gonext:gonext /app/migrations ./migrations
COPY --from=builder --chown=gonext:gonext /app/docs ./docs

RUN mkdir -p /app/data /app/uploads && chown -R gonext:gonext /app

USER gonext

EXPOSE 8080

CMD ["./server"]
