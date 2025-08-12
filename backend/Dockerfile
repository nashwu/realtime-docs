# build
FROM golang:1.22-alpine AS build
WORKDIR /app
RUN apk add --no-cache build-base git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /out/server ./cmd/server

# runtime stage
FROM gcr.io/distroless/base-debian12
WORKDIR /app
USER 65532:65532

COPY --from=build /out/server /app/server

# Defaults (override in compose)
ENV APP_ENV=prod \
    HTTP_ADDR=:8080 \
    PG_URL=postgres://postgres:secret@postgres:5432/docs?sslmode=disable \
    REDIS_ADDR=redis:6379 \
    JWT_SECRET=change-me \
    CORS_ALLOW=http://localhost:8081

EXPOSE 8080
ENTRYPOINT ["/app/server"]
