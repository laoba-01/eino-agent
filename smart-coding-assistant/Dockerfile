# ============================================================
# 统一 Dockerfile — go-zero monorepo，通过 --target 指定服务
# 示例: docker build --target auth-rpc -t auth-rpc .
# ============================================================

# ---------- Go 构建基础层 ----------
FROM golang:1.25-alpine AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# ---------- auth-service ----------
FROM go-builder AS auth-build
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./app/auth

FROM alpine:latest AS auth-rpc
RUN apk --no-cache add ca-certificates
COPY --from=auth-build /service /service
COPY app/auth/etc /etc/auth
EXPOSE 50054
CMD ["/service", "-f", "/etc/auth/auth.yaml"]

# ---------- memory-service ----------
FROM go-builder AS memory-build
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./app/memory

FROM alpine:latest AS memory-rpc
RUN apk --no-cache add ca-certificates
COPY --from=memory-build /service /service
COPY app/memory/etc /etc/memory
EXPOSE 50053
CMD ["/service", "-f", "/etc/memory/memory.yaml"]

# ---------- core-service ----------
FROM go-builder AS core-build
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./app/core

FROM alpine:latest AS core-rpc
RUN apk --no-cache add ca-certificates
COPY --from=core-build /service /service
COPY app/core/etc /etc/core
EXPOSE 50051
CMD ["/service", "-f", "/etc/core/core.yaml"]

# ---------- tool-service ----------
FROM go-builder AS tool-build
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./app/tool

FROM alpine:latest AS tool-rpc
RUN apk --no-cache add ca-certificates
COPY --from=tool-build /service /service
COPY app/tool/etc /etc/tool
EXPOSE 50052
EXPOSE 8081
CMD ["/service", "-f", "/etc/tool/tool.yaml"]

# ---------- api-gateway ----------
FROM go-builder AS gateway-build
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./app/gateway

FROM alpine:latest AS api-gateway
RUN apk --no-cache add ca-certificates
COPY --from=gateway-build /service /service
COPY app/gateway/etc /etc/gateway
EXPOSE 8080
CMD ["/service", "-f", "/etc/gateway/gateway.yaml"]

# ---------- frontend ----------
FROM nginx:alpine AS frontend
COPY frontend/ /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
