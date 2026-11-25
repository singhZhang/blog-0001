#!/usr/bin/env bash
# 在服务器上重建接口镜像并重启容器，然后做健康检查。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SERVICE_DIR="$ROOT/service"
DEPLOY_DIR="$SERVICE_DIR/deploy"
IMAGE="${BLOG_IMAGE:-blog-service:local}"
HEALTH_URL="${DEPLOY_HEALTH_URL:-http://127.0.0.1:8080/api/health}"

if [[ ! -f "$SERVICE_DIR/Dockerfile" ]]; then
  echo "找不到 Dockerfile: $SERVICE_DIR/Dockerfile" >&2
  exit 1
fi
if [[ ! -f "$DEPLOY_DIR/docker-compose.yml" ]]; then
  echo "找不到 docker-compose.yml: $DEPLOY_DIR/docker-compose.yml" >&2
  exit 1
fi
if [[ ! -f "$DEPLOY_DIR/.env" ]]; then
  echo "缺少 $DEPLOY_DIR/.env，请先从 .env.example 复制并填写" >&2
  exit 1
fi

echo "==> 构建镜像 $IMAGE"
cd "$SERVICE_DIR"
docker build -t "$IMAGE" .

echo "==> 启动容器"
cd "$DEPLOY_DIR"
BLOG_IMAGE="$IMAGE" docker compose up -d
docker compose ps

echo "==> 健康检查 $HEALTH_URL"
ok=0
for i in 1 2 3 4 5 6 7 8 9 10; do
  if curl -fsS "$HEALTH_URL" >/dev/null; then
    ok=1
    break
  fi
  sleep 2
done
if [[ "$ok" -ne 1 ]]; then
  echo "健康检查失败" >&2
  docker compose logs --tail=100 blog-service >&2 || true
  exit 1
fi

echo "==> 接口发布完成"
