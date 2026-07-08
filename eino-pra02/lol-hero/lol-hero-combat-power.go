package lolhero

import (
	"context"
	"errors"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type HeroCombatPower struct {
	Name        string
	CombatPower float64
	Country     string
	Lines       string
}

type HeroCombatPowerReq struct {
	Type string `json:"type" jsonschema:"required,enum=name,enum=combat-power,enum=country,enum=lines,description=查询英雄信息的维度，只限于名字（name）,属国（country）,台词（lines）；如果传递台词进行contains匹配"`
	Key  string `json:"key"  jsonschema:"required,description=要查询的关键字"`
}

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

var HeroCombatPowerTool tool.InvokableTool

func init() {
	var err error
	HeroCombatPowerTool, err = utils.InferTool(
		"search hero combit",
		"根据英雄信息查询英雄战力，适用于问题中带有关键字 'LoL','战力' 时，不适用于其他情况",
		SearchHeroCombatPower,
	)
	if err != nil {
		panic(err)
	}
}
