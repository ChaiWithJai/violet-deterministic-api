FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/server

FROM alpine:3.20
RUN adduser -D -g '' appuser
USER appuser
WORKDIR /home/appuser
COPY --from=build /out/app /usr/local/bin/app
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/app"]
