package main

import (
	"context"
	"einopra/config"
	"fmt"
	"io"
	"os"
	"strings"

	einoOpenApi "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {

	// 构建chatModel，chatModel负责和大模型进行对话
	chatModel, err := einoOpenApi.NewChatModel(context.Background(), &einoOpenApi.ChatModelConfig{
		APIKey:      strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
		BaseURL:     config.DEEPSEEK_BASE_URL,
		Model:       config.DEEPSEEK_V4_FLASH,
		Temperature: ptr(float32(0.3)),
	})
	if err != nil {
		panic(err)
	}

	// 构建msgs,这里我们还是用 pra01 中的例子
	msgs := []*schema.Message{
		{
			Role:    schema.System,
			Content: "你是一个文学家，作答时需要引用古诗词。",
		},
		{
			Role:    schema.User,
			Content: "我要上撤所",
		},
	}

	// 等待全部返回使用 Generate
	// resp, err := chatModel.Generate(context.Background(), msgs)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(resp.Content)
	// fmt.Printf("%+v", resp.ResponseMeta.Usage)

	fmt.Println(">>>我是分割线<<<")
	// 流式返回使用 Stream， 入参其实和Generate是一样的
	// model.WithTemperature(float32(1.0) 可以再此次修改配置，不影响原配置
	streamResp, err := chatModel.Stream(context.Background(), msgs, model.WithTemperature(float32(1.0)))
	if err != nil {
		panic(err)
	}
	defer streamResp.Close()
	var respmsgs []schema.Message
	for {
		chunk, err := streamResp.Recv()
		if err == io.EOF {
			fmt.Println("END")
			// 如果结束了，这里chunk是nil，如何统计token的用量呢
			// 再最后一条消息中携带 num: 1992  FinishReason: stop  usage:TotalTokens 2019
			fmt.Printf("%+v", chunk)
			break
		}
		if err != nil {
			panic(err)
		}
		respmsgs = append(respmsgs, *chunk)
		fmt.Print(chunk.Content)
	}

	fmt.Println(">>>我是分割线<<<")
	for i, v := range respmsgs {
		if v.ResponseMeta != nil && v.ResponseMeta.Usage != nil {
			fmt.Println("num:", i, " FinishReason:", v.ResponseMeta.FinishReason, " usage:TotalTokens", v.ResponseMeta.Usage.TotalTokens)
			continue
		}
		if v.ResponseMeta != nil {
			fmt.Println("num:", i, " FinishReason:", v.ResponseMeta.FinishReason)
			continue
		}

	}

}
