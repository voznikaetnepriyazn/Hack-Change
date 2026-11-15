FROM golang:1.27.1-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY agent ./agent
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/backend ./agent/cmd/agent

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/backend /app/backend
COPY agent/internal/auth/users.json /app/agent/internal/auth/users.json
RUN mkdir -p /app/logs && chown -R app:app /app
USER app
EXPOSE 8080
ENTRYPOINT ["/app/backend"]
