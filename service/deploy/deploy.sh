#!/usr/bin/env bash
# 服务器发布入口：拉最新 main，再按范围发布前端和/或接口。
# 用法：
#   ./deploy.sh           # 等同 all
#   ./deploy.sh all
#   ./deploy.sh web
#   ./deploy.sh api
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TARGET="${1:-all}"

case "$TARGET" in
  all|web|api) ;;
  *)
    echo "用法: $0 [all|web|api]" >&2
    exit 1
    ;;
esac

echo "==> 仓库目录: $ROOT"
cd "$ROOT"

if [[ ! -d .git ]]; then
  echo "当前目录不是 git 仓库: $ROOT" >&2
  exit 1
fi

echo "==> 拉取 origin/main"
git fetch origin
git checkout main
git pull --ff-only origin main

echo "==> 当前提交: $(git rev-parse --short HEAD)"

case "$TARGET" in
  all)
    "$SCRIPT_DIR/deploy-web.sh"
    "$SCRIPT_DIR/deploy-api.sh"
    ;;
  web)
    "$SCRIPT_DIR/deploy-web.sh"
    ;;
  api)
    "$SCRIPT_DIR/deploy-api.sh"
    ;;
esac

echo "==> 发布完成 ($TARGET)"
