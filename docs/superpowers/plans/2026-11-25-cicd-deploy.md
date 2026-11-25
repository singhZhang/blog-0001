# CI/CD 自动发布 Implementation Plan

> **For agentic workers:** 按任务顺序实现；步骤用 checkbox 跟踪。

**Goal:** 推送 `main` 时通过 GitHub Actions SSH 到服务器，重新打包前端并重启接口；支持手动只发其一。

**Architecture:** Actions 只负责触发与 SSH；服务器上的 `deploy.sh` 拉代码并调用 `deploy-web.sh` / `deploy-api.sh`。

**Tech Stack:** GitHub Actions、`appleboy/ssh-action`、bash、Docker Compose、npm/Vite。

## Global Constraints

- 仅 `main` 自动发布；手动可选 `all` / `web` / `api`
- 构建在服务器完成，与 `service/deploy/README.md` 一致
- 不提交、不打印 `service/deploy/.env` 中的密钥

---

### Task 1: 发布脚本

- [ ] 写 `deploy-web.sh`、`deploy-api.sh`、`deploy.sh`，`set -euo pipefail`，可执行
- [ ] `deploy.sh` 支持参数 `all|web|api`，默认 `all`

### Task 2: GitHub Actions

- [ ] 写 `.github/workflows/deploy.yml`：push main + workflow_dispatch
- [ ] 传入 `DEPLOY_PATH`、`VITE_SITE_URL`、`DEPLOY_WEB_ROOT` 等环境变量后执行脚本

### Task 3: 文档

- [ ] 在 `service/deploy/README.md` 增加 CI/CD 与 Secrets 说明
- [ ] 根 README 链到该节（可选一行）

### Task 4: 提交推送

- [ ] 提交设计、脚本、workflow、文档并 push `main`
