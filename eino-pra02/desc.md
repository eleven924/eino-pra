# eino-pra02

> [B站原视频](https://www.bilibili.com/video/BV1auJw6CECD?spm_id_from=333.788.player.switch&vd_source=f6a61857b96ef315872614abf06a4fbf&p=16)

## 目标

1. 通过 util.NewTool 添加工具定义以及对应函数来创建工具
2. 通过 utils.InferTool 来根据定义的函数以及其请求参数中的tag来创建工具
3. 使用 ToolNode 来调用 LLM 返回的 ToolCalls 中的多个工具
