# STAGE 1: Builder
FROM golang:1.22-alpine3.19 AS builder
RUN apk add --no-cache ca-certificates=20261020-r0 tzdata=2025b-r0
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/medcore-service main.go

# STAGE 2: Runtime
FROM alpine:3.19 AS runner
RUN addgroup -g 10001 -S medcore && \
    adduser -u 10001 -S medcore -G medcore
RUN apk add --no-cache ca-certificates=20261020-r0 tzdata=2025b-r0
WORKDIR /app
COPY --from=builder /bin/medcore-service /app/medcore-service
RUN chown -R medcore:medcore /app
USER 10001:10001
ENTRYPOINT ["/app/medcore-service"]
