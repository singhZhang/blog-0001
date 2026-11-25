---
slug: tool-schema-rejects-before-execution
title: 工具参数先校验，再执行
publishedAt: 2026-03-14
category: Agent
tags: Tool Calling, JSON Schema
---

函数调用返回的是一段结构化参数。运行时拿到它之后，第一件事是对照 schema 做校验。函数名不在允许列表、缺必填字段、类型不对、枚举值不在集合里，都不要执行函数。把校验错误原样作为这一步的结果交回去，模型才有依据改参数。运行时若自作主张填默认值，错误会被掩盖，下一次仍会以同样的方式错。

schema 里的描述要写约束，不要写愿望。日期字段写 `format: date`，并在描述里给 `2026-03-14` 这种例子。状态字段用 enum，列出 `pending`、`paid`、`cancelled`，不要写“订单状态字符串”。数量用 integer 且 `minimum: 1`。模型仍然可能输出字符串 `"1"`，校验会拒绝，拒绝信息里说明期望 integer。这比在函数内部悄悄转型更容易查。

一个最小的订单查询可以写成：

```json
{
  "name": "get_order",
  "parameters": {
    "type": "object",
    "required": ["id"],
    "additionalProperties": false,
    "properties": {
      "id": { "type": "string", "minLength": 1 }
    }
  }
}
```

`additionalProperties: false` 让多余字段失败，而不是被忽略。忽略多余字段时，模型会以为那些字段生效了。

副作用决定能不能重试。`get_order` 超时后可以再调一次。`refund_order` 超时后不能假设没退成。退款接口需要幂等键，同一次业务请求携带同一个键，服务端只处理一次。工具描述里写上“timeout 之后不要自动重试，把 timeout 交回循环”。循环看到 timeout，可以停下来问人，而不是换个说法再退一次。

权限错误和参数错误也要分开。参数错误让模型修改调用。权限错误让它换一条允许的路径或停止。两种都返回 500 时，模型只能继续猜。
