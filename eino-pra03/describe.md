# eino-pra03

> [原课程地址](https://www.bilibili.com/video/BV1auJw6CECD?spm_id_from=333.788.player.switch&vd_source=f6a61857b96ef315872614abf06a4fbf&p=18)

## 目标

使用 eino 框架中的 chatModelAgent 来完成，模型调用，ReAct，以及工具调用


## 问题

### 1. 自己定义的tool失败后，会中断 chatModelAgent 的流转。

由于 tool 的函数的 req 对应的 tag 误将 combat-power 也写到枚举中，再加上可能 LLM 理解上有出入，当我们使用英雄名查询战力信息是，可能查询字段 CompatPower，这样直接在 tool 函数直接返回error就会断开agent的处理循环
```golang
// 由于存在 enum=combat-power ， 会有一定概率， type=combat-power
type HeroCombatPowerReq struct {
	Type string `json:"type" jsonschema:"required,enum=name,enum=combat-power,enum=country,enum=lines,description=查询英雄信息的维度，只限于名字（name）,属国（country）,台词（lines）；如果传递台词进行contains匹配"`
	Key  string `json:"key"  jsonschema:"required,description=要查询的关键字"`
}
```

tool 查询函数在查询不到的时候直接返回error
```golang
func SearchHeroCombatPower(_ context.Context, req *HeroCombatPowerReq) (*HeroCombatPower, error) {
	for _, hero := range heroList {
		switch req.Type {
		case "name":
			if hero.Name == req.Key {
				return &hero, nil
			}
		case "country":
			if hero.Country == req.Key {
				return &hero, nil
			}
		case "lines":
			if hero.Lines == req.Key {
				return &hero, nil
			}
		}
	}
	return nil, errors.New("hero combat not found")
}

```
在 循环event的时候就直接将工具的报错传递给resEvent.Err，进而中断agent流转
```golang
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
```

**解决办法：**
1. 如果查询不到数据，直接返回 nil,nil ；不返回错误，这样agent会在没有查询结果的时候继续再次查询
2. 封装响应数据，将非致命error合并到实际响应数据中，这样 LLM 也会读取到错误，可以继续优化
```golang
type HeroCombatPowerResp struct {
	Item      *HeroCombatPower
	ToolErrorInfo string
	
}
```
调用结果如下
```markdown
===>>> Agent response count: 2 content: {"role":"tool","content":"{\"Item\":null,\"ErrorInfo\":\"hero combat not found\"}","tool_call_id":"call_00_ECRk1dTkQEtmqqr4pkUh8131","tool_name":"search-hero-combit","extra":{"_eino_msg_id":"18611440-d0db-48c6-9dbb-490028ff2230"}}
===>>> Agent response count: 3 content: {"role":"tool","content":"{\"Item\":null,\"ErrorInfo\":\"hero combat not found\"}","tool_call_id":"call_01_UuWjgRtVy1wVt1yreI6N6266","tool_name":"search-hero-combit","extra":{"_eino_msg_id":"e658bcb5-bb52-4631-b9a5-fbb8b85dba0c"}}
===>>> Agent response count: 4 content: {"role":"assistant","content":"看起来直接查战力没有结果，我先查查这两位英雄的基本信息。","tool_calls":[{"index":0,"id":"call_00_dLASmGPKjWns7tzSXlGO7255","type":"function","function":{"name":"search-hero-combit","arguments":"{\"type\": \"name\", \"key\": \"盖伦\"}"}},{"index":1,"id":"call_01_5O7iDNbqbgq9wqVcQEh70250","type":"function","function":{"name":"search-hero-combit","arguments":"{\"type\": \"name\", \"key\": \"亚索\"}"}}],"response_meta":{"finish_reason":"tool_calls","usage":{"prompt_tokens":780,"prompt_token_details":{"cached_tokens":512},"completion_tokens":215,"total_tokens":995,"completion_token_details":{"reasoning_tokens":87}}},"reasoning_content":"查询没有返回结果，可能是因为\"战力\"这个字段在工具中不可用，或者查询方式不对。我再尝试用\"name\"类型来查询看看能否获取基本信息。等等，工具描述说 type 可以是 \"name\", \"combat-power\", \"country\", \"lines\"。但 combat-power 没有返回结果。\n\n我换个思路，用 type 为 \"name\" 来查询盖伦和亚索的信息。","extra":{"_eino_msg_id":"54e11472-1fa8-40b8-b283-519c9610f87e","openai-request-id":"ab57d823-abcb-4e93-83bb-c55d6b6a7b4d","reasoning-content":"查询没有返回结果，可能是因为\"战力\"这个字段在工具中不可用，或者查询方式不对。我再尝试用\"name\"类型来查询看看能否获取基本信息。等等，工具描述说 type 可以是 \"name\", \"combat-power\", \"country\", \"lines\"。但 combat-power 没有返回结果。\n\n我换个思路，用 type 为 \"name\" 来查询盖伦和亚索的信息。"}}
===>>> Agent response count: 5 content: {"role":"tool","content":"{\"Item\":{\"Name\":\"盖伦\",\"CombatPower\":86.5,\"Country\":\"德玛西亚\",\"Lines\":\"人在塔在！\"},\"ErrorInfo\":\"\"}","tool_call_id":"call_00_dLASmGPKjWns7tzSXlGO7255","tool_name":"search-hero-combit","extra":{"_eino_msg_id":"2940f207-6b98-4fe5-a26f-af2339d1d7f4"}}
===>>> Agent response count: 6 content: {"role":"tool","content":"{\"Item\":{\"Name\":\"亚索\",\"CombatPower\":90.5,\"Country\":\"艾欧尼亚\",\"Lines\":\"死亡如风，常伴吾身。\"},\"ErrorInfo\":\"\"}","tool_call_id":"call_01_5O7iDNbqbgq9wqVcQEh70250","tool_name":"search-hero-combit","extra":{"_eino_msg_id":"b6d78622-1e06-4049-9f52-3736e2f4ebd5"}}
===>>> Agent response count: 7 content: {"role":"assistant","content":"查到了！以下是两位英雄的详细对比：\n\n---\n\n### 🛡️ 盖伦 —— 德玛西亚之力\n| 属性 | 内容 |\n|------|------|\n| **属国** | 德玛西亚 |\n| **战力** | **86.5** |\n| **台词** | \"人在塔在！\" |\n\n---\n\n### 🌪️ 亚索 —— 疾风剑豪\n| 属性 | 内容 |\n|------|------|\n| **属国** | 艾欧尼亚 |\n| **战力** | **90.5** |\n| **台词** | \"死亡如风，常伴吾身。\" |\n\n---\n\n### 🏆 结论\n\n**亚索（90.5）\u003e 盖伦（86.5）**\n\n从战力数值来看，**亚索略胜一筹**！不过盖伦那句\"人在塔在\"也是满满的安全感，两位英雄各有千秋~ 😄","response_meta":{"finish_reason":"stop","usage":{"prompt_tokens":1088,"prompt_token_details":{"cached_tokens":768},"completion_tokens":302,"total_tokens":1390,"completion_token_details":{"reasoning_tokens":104}}},"reasoning_content":"现在我有了两位英雄的完整信息：\n\n**盖伦**\n- 属国：德玛西亚\n- 战力：86.5\n- 台词：人在塔在！\n\n**亚索**\n- 属国：艾欧尼亚\n- 战力：90.5\n- 台词：死亡如风，常伴吾身。\n\n既然用户问谁的战力更高，我需要比较一下，亚索的90.5 \u003e 盖伦的86.5，所以亚索战力更高。","extra":{"_eino_msg_id":"7312edb6-1cca-4a71-9f3e-a4f8d586493f","openai-request-id":"4f13315b-577c-4f00-b3fb-15c167771ad0","reasoning-content":"现在我有了两位英雄的完整信息：\n\n**盖伦**\n- 属国：德玛西亚\n- 战力：86.5\n- 台词：人在塔在！\n\n**亚索**\n- 属国：艾欧尼亚\n- 战力：90.5\n- 台词：死亡如风，常伴吾身。\n\n既然用户问谁的战力更高，我需要比较一下，亚索的90.5 \u003e 盖伦的86.5，所以亚索战力更高。"}}
```
3. todo：eino自身是否存在错误回填，继续流转的方案呢？
