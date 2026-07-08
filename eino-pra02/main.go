package main

import (
	"context"
	calculatortool "einopra/eino-pra02/calculator-tool"
	"encoding/json"
	"fmt"
)

func main() {
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
