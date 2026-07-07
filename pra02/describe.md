# Pra02

## 目标

实现一个 agent 支持调用自定义tool来获取文件列表并读取指定文件内容（限制读取100行）


## 实现过程

1. 初始化 openai.Client
2. 定义 msgs 列表，组装 system 和 user 的 content
3. 构建自定义 tools -> 这里其实就是返回tool的定义 Name，Describe, Arguments
4. 进入 agent 循环流程，直到 LLM 不在调用工具 -> 这里是不是有其他终止条件呢？
5. 定义 tools 的执行流程.


## 对话测试

```bash

PS H:\mySpace\myGoSpace\einoPra> go run .\pra02\
>>> ReAct Step : 1
● Thought:
----- call tool, tool function Name: listTool  tool function params {"root": "pra02"}
>>> ReAct Step : 2
● Thought:  
----- call tool, tool function Name: readFileTool  tool function params {"file": "pra02/describe.md", "limit": 100}
----- call tool, tool function Name: readFileTool  tool function params {"file": "pra02/main.go", "limit": 100}
>>> ReAct Step : 3
● TaskRes:  pra02 是一个 **AI Agent（智能代理）的 Go 语言实现项目**，核心功能是让大语言模型（DeepSeek）能够通过 **ReAct 循环** 自主调用本地工具来解决问题。

### 主要干了什么：

1. **构建了一个 Agent 调度框架** — 通过 `RunReAct` 函数循环与大模型交互，每次判定是继续调用工具还是直接给出最终答案（最多可执行 `stepLimit` 步）。

2. **自定义了两个本地工具**供 LLM 调用：
   - **`listTool`** — 列出指定目录下的 `.md` 和 `.go` 文件。
   - **`readFileTool`** — 读取指定文件的内容（限制 100 行）。

3. **完整的 ReAct 循环流程**：
   - 组装 System + User 消息 → 调用大模型 → 解析返回的 `ToolCalls` → 如果有工具调用需求，则执行本地工具并将结果返回给模型 → 直到模型不再调用工具，输出最终结论。

4. **通过 DeepSeek API** 驱动，使用 `einopra/config` 中的配置（`DEEPSEEK_BASE_URL` 和 `DEEPSEEK_V4_FLASH` 模型）。

简单来说，**pra02 实现了一个能自主调用文件列表和文件读取工具的 AI 助手，用于分析本地 Go 项目结构**。

```