package main

import (
	"context"
	"einopra/config"
	calculatortool "einopra/eino-pra02/calculator-tool"
	lolhero "einopra/eino-pra02/lol-hero"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	einoOpenApi "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {

	chatModel, err := einoOpenApi.NewChatModel(context.Background(), &einoOpenApi.ChatModelConfig{
		APIKey:      strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		BaseURL:     config.DEEPSEEK_BASE_URL,
		Model:       config.DEEPSEEK_V4_FLASH,
		Temperature: ptr(float32(0.3)),
	})
	if err != nil {
		panic(err)
	}

	agent, err := adk.NewChatModelAgent(context.Background(), &adk.ChatModelAgentConfig{
		Name:          "lol-assistant",
		Description:   "一个LOL游戏助手，查询英雄基本信息，包括姓名，属国，战力，台词",
		Instruction:   "你是一个LOL游戏助手，适用于查询英雄基本信息，包括姓名，属国，战力，台词。不适用于其他属性的查询以及游戏的操作等",
		Model:         chatModel,
		MaxIterations: 8,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{lolhero.HeroCombatPowerTool, calculatortool.CalculatorTool},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	runner := adk.NewRunner(context.Background(), adk.RunnerConfig{
		Agent: agent,
	})

	msgs := []*schema.Message{
		{
			Role:    schema.User,
			Content: "盖伦和亚索谁的战力比较高?",
		},
	}
	resIter := runner.Run(context.Background(), msgs)

	i := 1
	for {
		resEvent, ok := resIter.Next()
		if !ok {
			break
		}

		if resEvent.Err != nil {
			//todo eino 这里如何让错误返回给大模型，让他继续分析？
			//"[NodeRunError] failed to invoke tool[name:search-hero-combit id:call_00_cOmHhM9ik4978mEw1hxS9889]: [LocalFunc] failed to invoke tool, toolName=search-hero-combit, err=hero combat not found\n------------------------\nnode path: [node_1, ToolNode]"
			fmt.Printf("%#v \n", resEvent.Err.Error())
			return
		}

		if resEvent.Output == nil || resEvent.Output.MessageOutput == nil {
			continue
		}

		messageContent, err := resEvent.Output.MessageOutput.GetMessage()
		if err != nil {
			panic(err)
		}
		res, _ := json.Marshal(messageContent)
		fmt.Println("===>>> Agent response count:", i, "content:", string(res))
		i++
	}

}
