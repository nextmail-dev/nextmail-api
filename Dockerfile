# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/nextmail-api ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.20

RUN adduser -D -u 10001 app && mkdir -p /data && chown app:app /data

COPY --from=build /out/nextmail-api /usr/local/bin/nextmail-api

USER app

# GeoLite2-Country.mmdb is not redistributable; mount it at /data:
#   docker run -v ./data:/data nextmail-api
ENV GEOIP_DB_PATH=/data/GeoLite2-Country.mmdb
EXPOSE 8080

ENTRYPOINT ["nextmail-api"]
