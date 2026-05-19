FROM golang:1.25-alpine AS builder

ARG SERVICE
WORKDIR /build

RUN apk add --no-cache git

COPY . .

WORKDIR /build/${SERVICE}
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

COPY --from=builder /app/service /app/service

EXPOSE 8080 50051 50052 50053 50054

ENTRYPOINT ["/app/service"]
