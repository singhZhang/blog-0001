---
slug: composer-classmap-and-psr4
title: classmap 省掉的是找文件，不是 new 对象
publishedAt: 2026-08-03
category: PHP
tags: Composer, Autoload
---

类第一次被用到时，自动加载函数才运行。开销大的时候，多半是不知道文件在哪，于是按目录拼路径，再对每个候选做一次文件是否存在的检查。失败的 stat 也会进剖面。目录和命名空间对不上时，一次未命中会打出一长串不存在的路径。

PSR-4 规定前缀和目录的对应关系。`App\Billing\` 对应 `src/Billing/` 时，`App\Billing\Invoice` 的路径可以直接拼出来：`src/Billing/Invoice.php`。不需要把 src 扫一遍。前缀没注册对，拼出来的路径就是错的，自动加载会失败，而不是退回去全盘搜索。

Composer 的 classmap 更进一步，生成类名到真实路径的数组。加载时是一次哈希查找。`composer dump-autoload -o` 会把可解析的类尽量放进 classmap。权威依赖适合这样处理。项目自己还在频繁加类时，可以继续用 PSR-4，避免每次加文件都重新生成映射。

这只影响“这个类的文件在哪”。对象已经创建之后，自动加载不会再跑。剖面里若时间在 PDO 和 SQL，换 classmap 看不出差别。剖面里若出现大量 `stat` 失败并且栈在 ClassLoader，再查前缀和 classmap 是否过期。

生成文件是构建产物。`vendor/composer/autoload_classmap.php` 和 `autoload_psr4.php` 随依赖安装产生。只同步了业务源码、没重新安装或 dump，运行中的加载器仍指向旧路径。类找不到时，先对一下这两份文件里的路径和磁盘上的文件是不是同一次部署。
