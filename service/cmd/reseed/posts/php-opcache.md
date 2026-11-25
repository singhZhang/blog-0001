---
slug: opcache-caches-opcodes
title: OPcache 里放的是 opcode，不是函数结果
publishedAt: 2025-12-16
category: PHP
tags: OPcache, PHP
---

PHP 执行文件前把源码编译成 opcode。没有 OPcache 时，每个请求都编译一次。OPcache 把编译结果放进共享内存，同一台机器上的 worker 共用这份结果。它不记住函数返回了什么，也不记住 SQL 的结果。那些仍要应用自己缓存。

文件改了之后是否重新编译，由 validate_timestamps 和 revalidate_freq 决定。validate_timestamps = 1 时，超过 revalidate_freq 秒就检查文件时间戳，变了就重新编译。开发机上这样能在保存后看到新代码。生产上若把 validate_timestamps 设为 0，请求路径上不再 stat 源文件，部署新代码后，已经在缓存里的 opcode 继续执行，直到进程重启或主动 reset。

只 reset 一台机器不够。负载均衡后面有三台应用机，其中两台还拿着旧 opcode，用户会交替看到新旧行为。发布脚本要在每一台执行 reset，或滚动重启 php-fpm，让旧 worker 退出。重置前若流量已经进来，重置本身很快，但正在执行的请求仍会把旧代码跑完。

缓存也有容量。opcache.max_accelerated_files 小于项目里的 PHP 文件数时，多出来的文件每次请求都重新编译。memory_consumption 太小，缓存被写满后同样退回编译。`opcache_get_status()` 里可以看到 used memory、命中次数和缓存的脚本数。命中率低，先看是不是缓存太小或时间戳校验太频繁，而不是先改业务代码。

命令行的 `php` 和 fpm 不共享同一块 OPcache，除非使用 file cache 并且配置了同一目录。在 CLI 里 reset 了，fpm 里的缓存可能还在。部署说明里写清 reset 的是哪一个 SAPI。
