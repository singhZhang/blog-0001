# CI/CD 自动发布设计

日期：2026-11-25

## 目标

推送到 GitHub `main` 后，自动在生产服务器上重新打包前端，并重建、重启接口容器。同时保留在 Actions 里单独只发前端或只发接口的能力。

## 决策

| 项 | 选择 |
| --- | --- |
| 触发 | 推送到 `main`；另支持 `workflow_dispatch` |
| 执行位置 | GitHub Actions SSH 到服务器，在服务器上构建 |
| 推送时范围 | 前端 + 接口都发布 |
| 手动范围 | `all` / `web` / `api` |
| 数据 | 不碰 MySQL，不跑 `reseed`，不改服务器 `.env` 与 Nginx |

## 组件

1. `.github/workflows/deploy.yml`  
   - `push` 到 `main` → 发布 `all`  
   - 手动触发时按输入选择范围  
   - 用 SSH 在服务器执行仓库内的 `service/deploy/deploy.sh`

2. `service/deploy/deploy.sh`  
   - `git fetch` / `checkout main` / `pull --ff-only`  
   - 按参数调用 `deploy-web.sh`、`deploy-api.sh`

3. `service/deploy/deploy-web.sh`  
   - `npm ci`、`VITE_SITE_URL=... npm run build`  
   - `rsync -a --delete dist/` 到 `/var/www/blog`（需要 sudo 时使用 sudo）

4. `service/deploy/deploy-api.sh`  
   - `docker build -t blog-service:local .`  
   - `docker compose up -d`  
   - `curl` 检查 `http://127.0.0.1:8080/api/health`，失败则非 0 退出

## Secrets

| 名称 | 说明 |
| --- | --- |
| `DEPLOY_HOST` | 服务器主机名或 IP |
| `DEPLOY_USER` | SSH 用户 |
| `DEPLOY_SSH_KEY` | SSH 私钥 |
| `DEPLOY_PATH` | 服务器上的仓库路径，默认 `/opt/blog` |
| `VITE_SITE_URL` | 前端构建用的站点 URL，如 `https://blog.example.com` |

可选：`DEPLOY_PORT`（SSH 端口，默认 22）、`DEPLOY_WEB_ROOT`（静态目录，默认 `/var/www/blog`）。

## 成功标准

- 推 `main` 后 Actions 成功，线上页面与 `/api/health` 可用  
- 手动只选 `web` 时接口容器不被重建  
- 手动只选 `api` 时 `/var/www/blog` 不被 rsync  
- `service/deploy/.env` 永不进入仓库或 Actions 日志明文打印密码
