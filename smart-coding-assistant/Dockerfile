# ============================================================
# 统一 Dockerfile — 通过 --target 指定要构建的服务
# 示例: docker build --target core-service -t core-service .
# ============================================================

# ---------- Go 构建基础层 ----------
FROM golang:1.25-alpine AS go-base
WORKDIR /src

# ---------- auth-service ----------
FROM go-base AS auth-builder
COPY auth-service/go.mod auth-service/go.sum ./auth-service/
WORKDIR /src/auth-service
RUN go mod download
COPY auth-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service .

FROM alpine:latest AS auth-service
RUN apk --no-cache add ca-certificates
COPY --from=auth-builder /service /service
EXPOSE 50054
CMD ["/service"]

# ---------- memory-service ----------
FROM go-base AS memory-builder
COPY memory-service/go.mod memory-service/go.sum ./memory-service/
WORKDIR /src/memory-service
RUN go mod download
COPY memory-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service .

FROM alpine:latest AS memory-service
RUN apk --no-cache add ca-certificates
COPY --from=memory-builder /service /service
EXPOSE 50053
CMD ["/service"]

# ---------- core-service ----------
FROM go-base AS core-builder
COPY core-service/go.mod core-service/go.sum ./core-service/
WORKDIR /src/core-service
RUN go mod download
COPY core-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service .

FROM alpine:latest AS core-service
RUN apk --no-cache add ca-certificates
COPY --from=core-builder /service /service
EXPOSE 50051
CMD ["/service"]

# ---------- tool-service ----------
FROM go-base AS tool-builder
COPY tool-service/go.mod tool-service/go.sum ./tool-service/
WORKDIR /src/tool-service
RUN go mod download
COPY tool-service/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service .

FROM alpine:latest AS tool-service
RUN apk --no-cache add ca-certificates
COPY --from=tool-builder /service /service
EXPOSE 50052
EXPOSE 8081
CMD ["/service"]

# ---------- api-gateway ----------
FROM go-base AS gateway-builder
COPY api-gateway/go.mod api-gateway/go.sum ./api-gateway/
WORKDIR /src/api-gateway
RUN go mod download
COPY api-gateway/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /service .

FROM alpine:latest AS api-gateway
RUN apk --no-cache add ca-certificates
COPY --from=gateway-builder /service /service
EXPOSE 8080
CMD ["/service"]

# ---------- frontend ----------
FROM nginx:alpine AS frontend
COPY frontend/ /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
