FROM node:20-alpine AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY --from=web /internal/http/ui ./internal/http/ui
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/vda ./cmd/api

FROM golang:1.22-alpine
RUN adduser -D -g '' appuser
USER appuser
WORKDIR /home/appuser
COPY --from=build /out/vda /usr/local/bin/vda
EXPOSE 4020
ENTRYPOINT ["/usr/local/bin/vda"]
