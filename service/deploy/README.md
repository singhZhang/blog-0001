# 发布到线上服务器

博客分三部分：Nginx 提供前端静态文件并把 `/api` 转到本机接口，Go 接口跑在容器里，文章和主题放在 MySQL。接口不对外直接暴露。

```text
浏览器
  │ HTTPS
  ▼
Nginx :443
  ├── /        → /var/www/blog          （web 构建产物）
  └── /api/    → http://127.0.0.1:8080  （blog-service 容器）
                    └── MySQL blog 库
```

下文命令里，前端在仓库的 `web/`，接口镜像在 `service/`，编排文件在 `service/deploy/`。把 `blog.example.com` 换成实际上线域名。

## 1. 服务器上先准备好

- Docker 与 Docker Compose 插件
- Nginx
- Node.js 20 或更新版本，用来构建前端
- 一台接口容器能够访问的 MySQL。库名用 `blog`。账号需要对这个库有建表、读写权限。库已经存在时不要再执行 `CREATE DATABASE`，部分云数据库会把「库已存在」当成错误
- 域名的 DNS 已经指到这台机器

在 MySQL 里执行 `service/deploy/schema.sql`。服务启动时也会自己建表，手工执行一遍是为了先确认账号权限。

## 2. 把代码放到服务器

```bash
git clone <仓库地址> /opt/blog
cd /opt/blog
```

之后更新代码用 `git pull`。

## 3. 配置接口

```bash
cd /opt/blog/service/deploy
cp .env.example .env
```

按线上环境修改 `.env`，至少改这几项：

| 变量 | 填写 |
| --- | --- |
| `CORS_ALLOWED_ORIGINS` | 站点地址，例如 `https://blog.example.com`。协议、域名、端口要和浏览器地址栏一致，末尾不要加斜杠 |
| `MYSQL_HOST` | MySQL 地址。MySQL 在另一台机器上就填那台机器的地址；在本机且接口跑在容器里时，填宿主机地址，不要填容器里的 `127.0.0.1` |
| `MYSQL_PORT` | 一般是 `3306` |
| `MYSQL_USER` / `MYSQL_PASSWORD` | 数据库账号 |
| `MYSQL_DATABASE` | `blog` |
| `ADMIN_TOKEN` | 后台登录令牌，换成足够长的随机字符串。留空会关闭全部写接口 |
| `BLOG_BIND` | 保持 `127.0.0.1:8080`，只给本机 Nginx 访问 |

`.env` 含密码和令牌，不要提交到仓库。

## 4. 构建并启动接口

```bash
cd /opt/blog/service
docker build -t blog-service:local .

cd deploy
docker compose up -d
docker compose ps
curl -fsS http://127.0.0.1:8080/api/health
```

健康检查应返回 `{"status":"ok"}`。容器内端口固定是 8080，对外只绑定在宿主机的 `127.0.0.1:8080`。开发机上的 `18081` 只用于本地 Vite 代理，不要写进生产配置。

若容器反复重启，看日志：

```bash
docker compose logs --tail=100 blog-service
```

常见原因是 MySQL 地址从容器内不可达，或账号没有权限。库已经存在时，当前版本会直接连上去建表，不会因为 `database exists` 退出。

## 5. 构建前端并放到 Nginx 目录

生产构建使用相对路径 `/api`，由 Nginx 转发，不需要把开发代理 `18081` 写进构建结果。`VITE_SITE_URL` 只影响页面里的规范链接。

```bash
cd /opt/blog/web
npm ci
VITE_SITE_URL=https://blog.example.com npm run build
sudo mkdir -p /var/www/blog
sudo rsync -a --delete dist/ /var/www/blog/
```

`dist/` 里应能看到 `index.html` 和 `favicon.ico`。

## 6. 配置 Nginx

`/var/www/blog` 承接页面，`/api/` 转到本机 8080。前端用的是 History 路由，找不到文件时要回到 `index.html`，否则刷新文章页会 404。

```nginx
server {
    listen 80;
    server_name blog.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name blog.example.com;

    ssl_certificate     /etc/nginx/ssl/blog.example.com.crt;
    ssl_certificate_key /etc/nginx/ssl/blog.example.com.key;

    root /var/www/blog;
    index index.html;

    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

检查并加载：

```bash
sudo nginx -t && sudo systemctl reload nginx
```

还没有证书时，可以先只保留 80 端口，确认页面和接口都通，再补上 443。证书签发后把 `CORS_ALLOWED_ORIGINS` 改成 `https://` 开头的地址，并按下一节重启接口。

## 7. 验收

在服务器上：

```bash
curl -fsS https://blog.example.com/api/health
curl -fsS https://blog.example.com/api/articles
curl -fsS https://blog.example.com/api/categories
```

在浏览器里：

1. 打开首页，文章列表来自数据库。
2. 打开一篇文章，刷新，地址栏仍是该文章而不是 404。
3. 打开 `/admin`，用 `ADMIN_TOKEN` 登录。
4. 新建一篇草稿或一个测试主题，确认能保存；不需要的话再删掉。

## 8. 以后发新版本

接口代码有变化：

```bash
cd /opt/blog
git pull
cd service
docker build -t blog-service:local .
cd deploy
docker compose up -d
curl -fsS http://127.0.0.1:8080/api/health
```

只有前端变化：

```bash
cd /opt/blog
git pull
cd web
npm ci
VITE_SITE_URL=https://blog.example.com npm run build
sudo rsync -a --delete dist/ /var/www/blog/
```

文章和主题保存在 MySQL 里，发版不会覆盖它们。

空库需要初始文章时，在能访问这台 MySQL、并且安装了 Go 的机器上执行：

```bash
cd /opt/blog/service
go run ./cmd/reseed
```

这个命令会先清空 `articles` 和 `categories`，再写入 8 个主题和随代码保存的文章。已经有正式内容的库不要运行它。后台里手工创建的文章不在镜像里，重灌后会丢失。

## 9. 回滚

接口：检出上一版代码，重新构建并启动。

```bash
cd /opt/blog
git checkout <上一版提交>
cd service && docker build -t blog-service:local .
cd deploy && docker compose up -d
```

前端：用上一版重新 `npm run build`，再 `rsync` 到 `/var/www/blog`。

数据库没有随镜像回滚。发版前如果改过表结构，回滚代码前先确认旧程序还能读当前表。

## 10. 排障

| 现象 | 先看 |
| --- | --- |
| 容器起不来 | `docker compose logs --tail=100 blog-service`，多半是 MySQL 地址、账号或库权限 |
| 页面打开但文章一直加载失败 | 浏览器里请求的 `/api/articles` 是否 200；Nginx 是否把 `/api/` 转到 `127.0.0.1:8080` |
| 刷新文章页 404 | Nginx 的 `try_files` 是否回到 `index.html` |
| 后台登录后写文章失败 | `ADMIN_TOKEN` 是否与登录时填写的一致；改 `.env` 后要 `docker compose up -d` 才会生效 |
| 标签图标不更新 | 图标在前端构建产物里，重新构建并同步 `dist/`，再强制刷新 |
