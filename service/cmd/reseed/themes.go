package main

import "github.com/singhZhang/blog-0001/service/internal/category"

func themeList() []category.Category {
	return []category.Category{
		{Name: "AI", Description: "模型如何表示、采样和在推理时省下计算。关心机制，而不是口号。"},
		{Name: "运维", Description: "发布、探针和错误预算。把线上行为约束成可以执行的规则。"},
		{Name: "FDE", Description: "在客户环境里交付产品时，接口、数据和成功标准该怎么划清。"},
		{Name: "Agent", Description: "让模型调用工具并完成多步任务时，循环、契约和上下文怎么设计。"},
		{Name: "Skills", Description: "把一项可复用能力写成规程：何时触发、输入输出、以及不允许做什么。"},
		{Name: "PHP", Description: "请求如何穿过 FPM、OPcache 和自动加载，以及这些层次各自保证什么。"},
		{Name: "Python", Description: "解释器的执行模型：GIL、属性查找，以及对象为什么会这样表现。"},
		{Name: "Golang", Description: "调度、接口和取消。理解运行时在替你做什么，以及它不替你做什么。"},
	}
}
