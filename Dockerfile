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
CMD ["/app/go-redis"]
