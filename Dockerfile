FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o foolock .

FROM alpine:3.21
COPY --from=builder /app/foolock /foolock
ENTRYPOINT ["/foolock"]
