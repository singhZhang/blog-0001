# 零玖肆玖肆

个人技术博客。前端在 `web/`，接口在 `service/`，文章和主题存在 MySQL 里。两边放在同一个仓库，一次发布。

```text
浏览器
  ├── web/          Vue 3 页面，开发时由 Vite 提供
  └── /api          Go 接口，读写 MySQL
```

线上由 Nginx 提供前端静态文件，并把 `/api` 转到本机接口。完整步骤见 [service/deploy/README.md](service/deploy/README.md)。推送到 `main` 后由 GitHub Actions 自动打包前端并重启接口，Secrets 与手动脚本说明也在该文档的「以后发新版本」一节。

## 目录

| 路径 | 内容 |
| --- | --- |
| `web/` | Vue 3、TypeScript、Vite 前端，含后台 `/admin` |
| `service/` | Go HTTP 接口 |
| `service/deploy/` | 生产环境的 Compose、表结构和发布说明 |
| `go.work` | 让仓库根目录能识别 `service` 这个 Go 模块 |

## 本地运行

先起接口。开发时代理指向 `18081`，和生产用的 `8080` 分开：

```bash
cd service
PORT=18081 \
MYSQL_HOST=127.0.0.1 \
MYSQL_PORT=3306 \
MYSQL_USER=root \
MYSQL_PASSWORD=你的密码 \
MYSQL_DATABASE=blog \
go run ./cmd/server
```

密码没有默认值，必须通过环境变量传入。库已存在时服务会直接连接并建表。

再起前端：

```bash
cd web
npm install
npm run dev
```

打开 [http://127.0.0.1:5173](http://127.0.0.1:5173)。后台在 `/admin`，本地默认令牌是 `blog-admin`，对应环境变量 `ADMIN_TOKEN`。

## 检查

```bash
cd service && go test ./...
cd web && npm test
```

接口健康检查：`curl -fsS http://127.0.0.1:18081/api/health`，应返回 `{"status":"ok"}`。

## 不要提交的文件

`service/deploy/.env` 里有数据库密码和后台令牌，已在根目录 `.gitignore` 中忽略。线上配置从 `service/deploy/.env.example` 复制后再填写。
