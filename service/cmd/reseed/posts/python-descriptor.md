---
slug: data-descriptors-outrank-instance-dict
title: 资料描述符比实例字典先被看到
publishedAt: 2026-02-08
category: Python
tags: Descriptor, Python
---

读 `obj.attr` 时，解释器不是只查 `obj.__dict__`。查找会经过类，以及类的基类。类属性若实现了描述符协议，还会调用它的方法。协议就两个入口：读走 `__get__`，写走 `__set__`。同时有这两个方法的叫资料描述符。只有 `__get__` 的，比如普通函数，叫非资料描述符。

资料描述符优先于实例字典。`property` 就是资料描述符，所以下面的赋值不会在实例字典里悄悄另存一份：

```python
class Account:
    def __init__(self):
        self._cents = 0

    @property
    def cents(self):
        return self._cents

    @cents.setter
    def cents(self, value):
        if value < 0:
            raise ValueError("negative")
        self._cents = value
```

执行 `account.cents = -1` 时，因为类上的 `cents` 有 `__set__`，赋值进入 setter，抛出 ValueError。实例字典里不会出现键 `cents`。如果它只是非资料描述符，实例字典里的同名键会把类上的对象挡住，校验就不起作用了。

函数只有 `__get__`。`obj.method` 会调用函数的 `__get__`，返回绑定了 obj 的方法。若事先执行 `obj.method = 123`，实例字典有了这个键，再读 `obj.method` 得到的是 123，不再走类上的函数。这是非资料描述符让路给实例字典的结果，不是方法绑定偶尔失效。

排错时按这个顺序看：类上有没有资料描述符；实例字典里有没有同名键；类上有没有非资料描述符；最后才是普通类属性。框架里的字段类型、ORM 列、校验器，多数是把逻辑放进了资料描述符，让赋值绕不开它们。
