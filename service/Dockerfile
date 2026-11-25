# syntax=docker/dockerfile:1

# ---- 构建阶段：在构建机的原生架构上交叉编译，避免模拟目标架构 ----
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

# 使用 docker build --platform 或 buildx 时由构建器自动注入目标平台信息。
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

# 先只拷贝依赖清单再下载：go.mod/go.sum 不变时命中层缓存，加速重复构建。
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/blog-server ./cmd/server

# ---- 运行阶段：最小化运行时 ----
FROM alpine:3.23

# 健康检查使用 alpine 自带的 busybox wget，无需额外安装任何包。
RUN addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=builder /out/blog-server /usr/local/bin/blog-server
COPY content/articles /app/content/articles

USER app

# 容器内端口固定 8080：对外端口由运行时的端口映射决定，健康检查 URL 依赖此值。
ENV PORT=8080 \
    BLOG_CONTENT_DIR=/app/content/articles

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["wget", "-q", "--spider", "http://127.0.0.1:8080/api/health"]

ENTRYPOINT ["blog-server"]
