FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/proxy ./proxy

FROM alpine:3.22

RUN apk --no-cache add ca-certificates \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/proxy ./proxy

USER app
EXPOSE 8090
CMD ["./proxy"]
