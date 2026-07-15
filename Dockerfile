FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o go-redis .
LABEL org.opencontainers.image.source="https://github.com/mattcattb/go-redis"
FROM alpine:latest
COPY --from=build /app/go-redis /app/go-redis
RUN mkdir -p /data
WORKDIR /data
EXPOSE 6479
# EXPOSE is documentation only. Set REDIS_PORT and publish the same container
# port to listen elsewhere, for example: -e REDIS_PORT=6379 -p 6379:6379.
CMD ["/app/go-redis"]
