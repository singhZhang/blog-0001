---
slug: readiness-is-not-liveness
title: 存活探针失败会重启，就绪探针失败只会摘流量
publishedAt: 2026-01-19
category: 运维
tags: Kubernetes, Probe
---

Kubernetes 里这两类探针的后果不同。livenessProbe 失败，kubelet 杀掉容器，按重启策略拉起新进程。readinessProbe 失败，Pod 从 Service 的 Endpoints 里去掉，不再接收新的连接，进程本身还在。启动阶段还有 startupProbe，在它成功之前，存活和就绪都可以先不判，用来挡住启动很慢的进程被误杀。

把“依赖还没连上”写成存活失败，滚动发布时容易看到重启次数上涨。新容器起来，数据库连接还在建立，存活检查已经失败，容器被杀，连接又从头开始。就绪失败则只是这个 Pod 暂时不进 Endpoints。旧 Pod 继续接流量，新 Pod 准备好再加入。这才是滚动更新在用的门禁。

探针路径不要做重活。periodSeconds 常见是 10 秒，failureThreshold 设为 3 时，大约 30 秒连续失败才动作。如果这个路径每次都打数据库、扫磁盘，探针自己就会成为负载。存活检查确认进程还能在超时内返回即可。就绪检查确认关键依赖可用、监听端口已打开、必要的预热做完。预热没做完就返回 HTTP 503，调用方会看到摘除，而不是一串业务 500。

收到 SIGTERM 时，进程应先把就绪置失败，再等待已经接受的请求结束，然后退出。terminationGracePeriodSeconds 要比这个等待短的话，请求仍会被掐断。只在代码里监听信号、却不把就绪状态翻掉，Service 仍会把新请求送进来，这些请求会撞上正在退出的进程。

```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  periodSeconds: 10
  failureThreshold: 3
livenessProbe:
  httpGet:
    path: /live
    port: 8080
  periodSeconds: 10
  failureThreshold: 3
```

两条路径分开实现。/live 不检查数据库。/ready 检查数据库，失败时返回非 2xx。这样数据库抖动会把实例摘掉，而不会把整个部署重启一遍。
