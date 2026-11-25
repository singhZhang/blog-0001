#!/usr/bin/env bash
# 在服务器上构建前端并同步到 Nginx 静态目录。
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
WEB_DIR="$ROOT/web"
WEB_ROOT="${DEPLOY_WEB_ROOT:-/var/www/blog}"
SITE_URL="${VITE_SITE_URL:?请设置 VITE_SITE_URL，例如 https://blog.example.com}"

if [[ ! -d "$WEB_DIR" ]]; then
  echo "找不到前端目录: $WEB_DIR" >&2
  exit 1
fi

echo "==> 构建前端 ($SITE_URL)"
cd "$WEB_DIR"
npm ci
VITE_SITE_URL="$SITE_URL" npm run build

if [[ ! -f dist/index.html ]]; then
  echo "构建失败：dist/index.html 不存在" >&2
  exit 1
fi

echo "==> 同步到 $WEB_ROOT"
mkdir_cmd=(mkdir -p "$WEB_ROOT")
rsync_cmd=(rsync -a --delete dist/ "$WEB_ROOT/")
if [[ ! -w "$(dirname "$WEB_ROOT")" ]] || { [[ -d "$WEB_ROOT" ]] && [[ ! -w "$WEB_ROOT" ]]; }; then
  mkdir_cmd=(sudo mkdir -p "$WEB_ROOT")
  rsync_cmd=(sudo rsync -a --delete dist/ "$WEB_ROOT/")
fi
"${mkdir_cmd[@]}"
"${rsync_cmd[@]}"

echo "==> 前端发布完成"
