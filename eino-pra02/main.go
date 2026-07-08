package main

import (
	"context"
	calculatortool "einopra/eino-pra02/calculator-tool"
	lolhero "einopra/eino-pra02/lol-hero"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func main() {
	// invokeCalculatorTool()
	// invokeHeroCombatPowerTool()
	invokeTools()
}

func invokeCalculatorTool() {
	info, err := calculatortool.CalculatorTool.Info(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Calculator Tool Info: %+v \n ", info)

	reqstr, err := json.Marshal(calculatortool.CalculatorReq{
		ParamA: 2,
		ParamB: 5,
		Op:     "add",
	})

	if err != nil {
		panic(err)
	}

	res, err := calculatortool.CalculatorTool.InvokableRun(context.Background(), string(reqstr))
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}

func invokeHeroCombatPowerTool() {
	info, err := lolhero.HeroCombatPowerTool.Info(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hero Combat Power Tool Info: %+v \n ", info)

	reqstr, err := json.Marshal(lolhero.HeroCombatPowerReq{
		Type: "name",
		Key:  "嘉文四世",
	})

	if err != nil {
		panic(err)
	}

	res, err := lolhero.HeroCombatPowerTool.InvokableRun(context.Background(), string(reqstr))
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}

// invokeTools 模拟 大模型 响应，使用 ToolNode 调用两个工具
func invokeTools() {
	// 添加多个工具
	toolNode, err := compose.NewToolNode(context.Background(), &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{
			lolhero.HeroCombatPowerTool,
			calculatortool.CalculatorTool,
		},
	})

	if err != nil {
		panic(err)
	}

	reqstr, _ := json.Marshal(calculatortool.CalculatorReq{
		ParamA: 2,
		ParamB: 5,
		Op:     "add",
	})
	// mock LLM returned message
	llmMessage := &schema.Message{
		Role: schema.Assistant,
		ToolCalls: []schema.ToolCall{
			{
				Function: schema.FunctionCall{Name: "search hero combit", Arguments: `{"type":"name","key":"亚索"}`},
			},
			{
				Function: schema.FunctionCall{Name: "two number calculator", Arguments: string(reqstr)},
			},
		},
	}

	resp, err := toolNode.Invoke(context.Background(), llmMessage)
	if err != nil {
		panic(err)
	}

	for _, v := range resp {
		fmt.Printf("> toolName: %v,toolRes: %v \n", v.ToolName, v.Content)

	}
}
