---
slug: context-done-must-be-observed
title: 不读 ctx.Done，取消就不会发生
publishedAt: 2026-09-30
category: Golang
tags: context, Cancellation
---

`context.WithCancel` 返回一个子 context 和 cancel 函数。调用 cancel，或者父 context 先被取消，子 context 的 `Done` 通道会被关闭，`Err` 返回 `context.Canceled`。`WithTimeout` 同样会关闭 `Done`，到期时 `Err` 是 `context.DeadlineExceeded`。关闭通道是通知，运行时不会因此停掉正在执行的函数。

函数要停下来，得自己看到这个通知。循环里在下一轮开始前 `select` 一下 `ctx.Done()`；往下调用时把同一个 ctx 传进去。标准库的 `http.NewRequestWithContext`、数据库驱动的 `QueryContext`，内部在等待网络时监听了这个通道，所以上层取消能让它们返回。自己写的计算循环里如果没有 `Done` 的分支，超时到了循环仍会跑完。

不要在下层重新造一个更长的超时把上层盖住。父 context 已经取消时，从它派生的子 context 也是取消状态，不能靠再包一层把它变成有效。需要更短的期限时，用 `WithTimeout` 从父级派生，期限取父级剩余时间和你要的时间里更早的那个。库函数已经这样做了，自己包一层时也要保留父级的取消。

context 表示一次操作的范围，比如一次 HTTP 请求、一次批处理任务。请求结束就 cancel，避免下游还在使用已经结束的请求。把它存在长寿命的结构体里，取消时间和对象寿命就绑在一起：对象还在，取消可能永远不来；对象被提前丢掉，正在进行的操作会被误取消。每次操作开始时派生，操作结束时 cancel。

```go
ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()

select {
case <-ctx.Done():
    return ctx.Err()
default:
}
```

这段 `select` 只检查“此刻是否已经取消”。接下来若进入一个不看 ctx 的长循环，两秒的超时仍然管不到那个循环。超时要放在真正会等待或会循环的地方，而不是只在函数入口看一眼。
