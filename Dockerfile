FROM golang:1.23-alpine AS build
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags='-w -s' -o app main.go

# Use standard Alpine with platform specification
FROM --platform=linux/arm64 alpine:latest
WORKDIR /root
COPY --from=build /app/app ./

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

EXPOSE 3001
ENTRYPOINT ["./app"]