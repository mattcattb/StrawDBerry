FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go build -o straw-dberry .
LABEL org.opencontainers.image.source="https://github.com/mattcattb/StrawDBerry"
FROM alpine:latest
COPY --from=build /app/straw-dberry /app/straw-dberry
RUN mkdir -p /data
WORKDIR /data
EXPOSE 6479
CMD ["/app/straw-dberry"]
