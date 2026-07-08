package calculatortool

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

var CalculatorTool tool.InvokableTool

type CalculatorReq struct {
	ParamA float64 `json:"param_a"`
	ParamB float64 `json:"param_b"`
	Op     string  `json:"op"`
}

type CalculatorResp struct {
	Expression string  `json:"expression"`
	Result     float64 `json:"result"`
}

func Calaulate(_ context.Context, req CalculatorReq) (CalculatorResp, error) {
	switch req.Op {
	case "add":
		return CalculatorResp{
			Expression: fmt.Sprintf("%g add %g", req.ParamA, req.ParamB),
			Result:     req.ParamA + req.ParamB,
		}, nil
	case "sub":
		return CalculatorResp{
			Expression: fmt.Sprintf("%g sub %g", req.ParamA, req.ParamB),
			Result:     req.ParamA - req.ParamB,
		}, nil

	case "mul":
		return CalculatorResp{
			Expression: fmt.Sprintf("%g mul %g", req.ParamA, req.ParamB),
			Result:     req.ParamA * req.ParamB,
		}, nil
	case "div":
		if req.ParamB == 0 {
			return CalculatorResp{}, errors.New("ParamB 在 div 运算时不能为 0")
		}
		return CalculatorResp{
			Expression: fmt.Sprintf("%g div %g", req.ParamA, req.ParamB),
			Result:     req.ParamA / req.ParamB,
		}, nil
	}
	return CalculatorResp{}, errors.New("unknown op type")
}

func init() {
	CalculatorTool = utils.NewTool(
		// NewTool 第一个参数是 ToolInfo 负责定义tool的描述和参数，类似元数据
		&schema.ToolInfo{
			Name: "two number calculator",
			Desc: "一个计算器，适用：两个数字之间的加、减、乘、除运算，不适用：复杂计算、多参数计算、大数据计算; example: param_a=1,param_b=2,op=add",
			ParamsOneOf: schema.NewParamsOneOfByParams(
				map[string]*schema.ParameterInfo{
					"param_a": {
						Type:     schema.Number,
						Desc:     "用于计算的第一个数字",
						Required: true,
					},
					"param_b": {
						Type:     schema.Number,
						Desc:     "用于计算的第二个数字",
						Required: true,
					},
					"op": {
						Type:     schema.String,
						Desc:     "运算符",
						Required: true,
						Enum:     []string{"add", "sub", "mul", "div"},
					},
				},
			),
		},
		// NewTool 的第二个参数是计算器函数，函数的定义必须符合
		// type InvokeFunc[T, D any] func(ctx context.Context, input T) (output D, err error)
		Calaulate,
	)
}
