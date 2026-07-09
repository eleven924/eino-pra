package main

import (
	"context"
	"einopra/config"
	lolhero "einopra/eino-pra02/lol-hero"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	eino_openai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func main() {

	ctx := context.Background()

	chatTemplate := prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个{role}游戏助手，负责为用户推荐战力更高的英雄"),
		schema.UserMessage("对面选了{hero}，我应该选什么英雄呀"))

	chatModel, err := eino_openai.NewChatModel(ctx, &eino_openai.ChatModelConfig{
		APIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		BaseURL: config.DEEPSEEK_BASE_URL,
		Model:   config.DEEPSEEK_V4_FLASH,
	})
	if err != nil {
		panic(err)
	}

	heroCP := compose.InvokableLambda(func(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
		heroCombatPower, err := json.Marshal(lolhero.HeroCombatPowerList)
		if err != nil {
			fmt.Println("英雄战力数据追加失败")
			return msgs, nil
		}

		msgs = append(msgs,
			schema.SystemMessage(fmt.Sprintf("英雄战力数据如下：%s", heroCombatPower)),
		)
		return msgs, nil
	})

	chain := compose.NewChain[map[string]any, *schema.Message]().AppendChatTemplate(chatTemplate).AppendLambda(heroCP).AppendChatModel(chatModel)

	runner, err := chain.Compile(ctx)
	if err != nil {
		panic(err)
	}

	res, err := runner.Invoke(ctx, map[string]any{
		"role": "LOL",
		"hero": "亚索",
	})

	if err != nil {
		panic(err)
	}

	fmt.Println(">>> chain result")
	fmt.Println(res.Content)
}
