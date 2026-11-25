# 零玖肆玖肆 · 技术博客

基于 Vue 3、TypeScript 与 Vite 的博客前端。文章和主题来自 `service` 提供的接口。

## 本地运行

```bash
npm install
npm run dev
```

开发服务器把 `/api` 代理到 `http://127.0.0.1:18081`。后台入口是 `/admin`。

## 验证

```bash
npm test
```

## 页面

| 路径 | 页面 |
| --- | --- |
| `/` | 首页 |
| `/articles/:slug` | 文章详情 |
| `/archive` | 归档 |
| `/topics` | 主题总览 |
| `/topics/:topic` | 单个主题 |
| `/about` | 关于 |
| `/admin` | 后台 |

路由使用 HTML5 History。部署时把未命中的路径回退到 `index.html`。

## 目录

```text
src/
├── components/   站点组件，ui/ 放按钮和输入框
├── composables/  文章、主题、登录态与页面信息
├── data/         站点固定文案
├── lib/          可单测的纯函数
├── services/     接口访问
├── styles/       设计变量、全局样式和后台样式
├── types/        数据契约
├── views/        前台页面，admin/ 为后台
└── router.ts
```
