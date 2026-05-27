package tool

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

type AnalyseStock struct {
}

func (as *AnalyseStock) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name:        "analyse_stock",
		Desc:        "",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&jsonschema.Schema{}),
	}, nil
}

func (as *AnalyseStock) Run() {

}
