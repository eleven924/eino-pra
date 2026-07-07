package main

import (
	"bufio"
	"context"
	"einopra/config"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	listTool     = "listTool"
	ReadFileTool = "readFileTool"
)

func NewClient() (*openai.Client, error) {
	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("api key is not set")
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = config.DEEPSEEK_BASE_URL

	return openai.NewClientWithConfig(cfg), nil
}

// RunReAct 管理Agent调度过程，负责沟通大模型，和调用具体的tools
func RunReAct(ctx context.Context, c *openai.Client, question string, stepLimit int) error {

	// 定义Agent角色
	msgs := make([]openai.ChatCompletionMessage, 0, 2)
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: `你是一个本地golang项目分析助手；当你作回复时，先将你的判断写入到 content 字段中（必须，一两句话就行，相当于 ReAct 中的 Thought），
		然后在确实需要时调用工具，工具的返回会作为下一轮的上下问一并提供；已经拿到所有信息后，不在调用任何工具，直接把最终结果卸载 content 中`,
	})
	msgs = append(msgs, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: strings.TrimSpace(question),
	})

	tools, err := buildTools()
	if err != nil {
		panic(err)
	}

	// 循环调用
	for i := 1; i <= stepLimit; i++ {
		fmt.Println(">>> ReAct Step :", i)
		resp, err := c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    config.DEEPSEEK_V4_FLASH,
			Messages: msgs,
			Tools:    tools,
		})
		if err != nil {
			fmt.Println("Emm..., brain breaken... ", err.Error())
			continue
		}

		if len(resp.Choices) == 0 {
			fmt.Println("Emm..., brain empty... ")
			continue
		}

		toolCalls := resp.Choices[0].Message.ToolCalls
		if len(toolCalls) > 0 {
			// 代表还没有完成，有工具需要调用
			toughtContent := resp.Choices[0].Message.Content
			fmt.Println("● Thought: ", toughtContent)
			msgs = append(msgs, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: toughtContent,
				// 要存此项，否则调用报错 message: Messages with role 'tool' must be a response to a preceding message with 'tool_calls'
				ToolCalls: toolCalls,
			})

			// callTools
			for _, toolCall := range toolCalls {
				toolRes, err := callTool(toolCall)
				// 必须在tool的返回中带上ToolCallID,这样大模型才知道时哪个tool的返回结果
				if err != nil {
					msgs = append(msgs, openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						Content:    err.Error(),
						ToolCallID: toolCall.ID,
					})
				} else {
					msgs = append(msgs, openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						Content:    toolRes,
						ToolCallID: toolCall.ID,
					})
				}

			}

		} else {
			fmt.Println("● TaskRes: ", resp.Choices[0].Message.Content)
			return nil
		}
	}

	return fmt.Errorf("达到最大ReAct轮数：%d", stepLimit)
}

// callTool 实现对应tool的具体执行逻辑
func callTool(tool openai.ToolCall) (string, error) {
	fmt.Println("----- call tool, tool function Name:", tool.Function.Name, " tool function params", tool.Function.Arguments)
	switch tool.Function.Name {
	case listTool:
		var params1 struct {
			Root string `json:"root"`
		}
		err := json.Unmarshal([]byte(tool.Function.Arguments), &params1)
		if err != nil {
			return "", err
		}
		return listFiles(params1.Root)
	case ReadFileTool:

		var params2 struct {
			File  string `json:"file"`
			Limit int    `json:"limit"`
		}
		err := json.Unmarshal([]byte(tool.Function.Arguments), &params2)
		if err != nil {
			return "", err
		}
		fileContent, err := execReadFile(params2.File, params2.Limit)
		// fmt.Println("fileContent: ", fileContent)
		return fileContent, err
	}

	return "", fmt.Errorf("tool func not found")
}

func listFiles(target string) (string, error) {
	if target == "" {
		target = "."
	}
	var mdfiles []string
	filepath.WalkDir(target, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".go") {
			mdfiles = append(mdfiles, path)
		}
		return nil
	})

	return strings.Join(mdfiles, ","), nil

}

func execReadFile(file string, limit int) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > limit {
			break
		}
	}

	if scanner.Err() != nil && scanner.Err() != io.EOF {
		return "", scanner.Err()
	}

	return strings.Join(lines, "\n"), nil
}

// buildTools 负责构建tools列表，通知大模型有什么工具可以调用
func buildTools() ([]openai.Tool, error) {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        listTool,
				Description: "列出指定目录下的 .md 和 .go 结尾的文件，返回相对路径列表，不递归 vendor, node_moudles 等目录",
				// 参数是否开启严格模式，强制模型输出的参数完全匹配 JSON Schema
				Strict: false,
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"root": {"type": "string", "description": "目录路径，默认当前目录"}
					}
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        ReadFileTool,
				Description: "读取file中的内容，最多读取limit行",
				// 参数是否开启严格模式，强制模型输出的参数完全匹配 JSON Schema
				Strict: false,
				Parameters: json.RawMessage(`{
					"type": "object",
					"required": ["file", "limit"],
					"properties": {
						"file": {"type": "string", "description": "需要读取的文件路径"},
						"limit": {"type": "integer", "description": "需要读取的文件的前多少行"}
					}
				}`),
			},
		},
	}, nil
}

func main() {
	client, err := NewClient()
	if err != nil {
		panic(err)
	}
	question := "帮我看下pra02主要干了什么事情"
	err = RunReAct(context.Background(), client, question, 10)
	if err != nil {
		panic(err)
	}
}
