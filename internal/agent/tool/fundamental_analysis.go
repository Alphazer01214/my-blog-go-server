package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/parquet-go/parquet-go"
)

// FundamentalAnalysisTool 基本面分析工具
type FundamentalAnalysisTool struct {
	DataDir string
}

// FundamentalParams 参数
type FundamentalParams struct {
	Action string `json:"action"` // get_financial / get_valuation / get_profitability
	TsCode string `json:"ts_code"`
}

// FinancialRow 财务数据行
type FinancialRow struct {
	Symbol      string  `parquet:"symbol"`
	Name        string  `parquet:"name"`
	Date        string  `parquet:"date"`
	Revenue     float64 `parquet:"revenue"`      // 营收(亿)
	NetProfit   float64 `parquet:"net_profit"`    // 净利润(亿)
	GrowthRate  float64 `parquet:"growth_rate"`   // 营收增长率(%)
	ProfitRate  float64 `parquet:"profit_rate"`   // 净利润率(%)
	ROE         float64 `parquet:"roe"`           // 净资产收益率(%)
	PE          float64 `parquet:"pe"`            // 市盈率
	PB          float64 `parquet:"pb"`            // 市净率
	TotalMV     float64 `parquet:"total_mv"`      // 总市值(亿)
	CircMV      float64 `parquet:"circ_mv"`       // 流通市值(亿)
	DebtRatio   float64 `parquet:"debt_ratio"`    // 资产负债率(%)
	CurrentRatio float64 `parquet:"current_ratio"` // 流动比率
}

func NewFundamentalAnalysisTool(dataDir string) *FundamentalAnalysisTool {
	if dataDir == "" {
		dataDir = "resources"
	}
	return &FundamentalAnalysisTool{DataDir: dataDir}
}

func (fa *FundamentalAnalysisTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "fundamental_analysis",
		Desc: `获取股票基本面分析数据。支持：财务数据(get_financial)、估值指标(get_valuation)、盈利能力(get_profitability)。
当用户询问股票基本面、估值、财务状况、盈利能力时使用此工具。
注意：数据来自本地数据集，可能不是最新数据。`,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"action": {
					Type:     "string",
					Desc:     "操作类型：get_financial(财务数据)/get_valuation(估值指标)/get_profitability(盈利能力)",
					Required: true,
				},
				"ts_code": {
					Type:     "string",
					Desc:     "股票代码，如 sh600519 或 600519",
					Required: true,
				},
			},
		),
	}, nil
}

func (fa *FundamentalAnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params FundamentalParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	symbol := normalizeSymbol(params.TsCode)

	// 尝试从 financial.parquet 读取
	rows, err := fa.readFinancialData(symbol)
	if err != nil {
		return fmt.Sprintf("无法获取 %s 的基本面数据: %v", symbol, err), nil
	}

	if len(rows) == 0 {
		return fmt.Sprintf("未找到 %s 的基本面数据", symbol), nil
	}

	row := rows[len(rows)-1] // 取最新一条

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s (%s) 基本面分析\n\n", row.Name, symbol))

	switch params.Action {
	case "get_financial":
		b.WriteString("### 财务数据\n\n")
		b.WriteString(fmt.Sprintf("- **报告期**: %s\n", row.Date))
		b.WriteString(fmt.Sprintf("- **营业收入**: %.2f 亿元\n", row.Revenue))
		b.WriteString(fmt.Sprintf("- **净利润**: %.2f 亿元\n", row.NetProfit))
		b.WriteString(fmt.Sprintf("- **营收增长率**: %.2f%%\n", row.GrowthRate))
		b.WriteString(fmt.Sprintf("- **净利润率**: %.2f%%\n", row.ProfitRate))
		b.WriteString(fmt.Sprintf("- **资产负债率**: %.2f%%\n", row.DebtRatio))
		b.WriteString(fmt.Sprintf("- **流动比率**: %.2f\n", row.CurrentRatio))

	case "get_valuation":
		b.WriteString("### 估值指标\n\n")
		b.WriteString(fmt.Sprintf("- **市盈率(PE)**: %.2f\n", row.PE))
		b.WriteString(fmt.Sprintf("- **市净率(PB)**: %.2f\n", row.PB))
		b.WriteString(fmt.Sprintf("- **总市值**: %.2f 亿元\n", row.TotalMV))
		b.WriteString(fmt.Sprintf("- **流通市值**: %.2f 亿元\n", row.CircMV))
		b.WriteString("\n**估值分析**:\n")
		if row.PE > 0 && row.PE < 15 {
			b.WriteString("- PE < 15，估值偏低\n")
		} else if row.PE >= 15 && row.PE <= 30 {
			b.WriteString("- PE 15~30，估值适中\n")
		} else if row.PE > 30 {
			b.WriteString("- PE > 30，估值偏高\n")
		}
		if row.PB < 1 {
			b.WriteString("- PB < 1，可能被低估或基本面较弱\n")
		} else if row.PB >= 1 && row.PB <= 3 {
			b.WriteString("- PB 1~3，估值合理\n")
		} else if row.PB > 3 {
			b.WriteString("- PB > 3，估值偏高\n")
		}

	case "get_profitability":
		b.WriteString("### 盈利能力\n\n")
		b.WriteString(fmt.Sprintf("- **净资产收益率(ROE)**: %.2f%%\n", row.ROE))
		b.WriteString(fmt.Sprintf("- **净利润率**: %.2f%%\n", row.ProfitRate))
		b.WriteString(fmt.Sprintf("- **营收增长率**: %.2f%%\n", row.GrowthRate))
		b.WriteString("\n**盈利能力分析**:\n")
		if row.ROE > 15 {
			b.WriteString("- ROE > 15%，盈利能力优秀\n")
		} else if row.ROE >= 10 {
			b.WriteString("- ROE 10~15%，盈利能力良好\n")
		} else if row.ROE > 0 {
			b.WriteString("- ROE < 10%，盈利能力一般\n")
		} else {
			b.WriteString("- ROE 为负，公司亏损\n")
		}
		if row.GrowthRate > 20 {
			b.WriteString("- 营收增长率 > 20%，高成长\n")
		} else if row.GrowthRate > 0 {
			b.WriteString("- 营收正增长\n")
		} else {
			b.WriteString("- 营收负增长，需关注\n")
		}

	default:
		return "未知操作类型，支持: get_financial, get_valuation, get_profitability", nil
	}

	return b.String(), nil
}

func (fa *FundamentalAnalysisTool) readFinancialData(symbol string) ([]FinancialRow, error) {
	// 尝试读取 financial.parquet
	filePath := fa.DataDir + "/financial.parquet"
	rows, err := parquet.ReadFile[FinancialRow](filePath)
	if err != nil {
		return nil, err
	}

	// 过滤
	var filtered []FinancialRow
	for _, row := range rows {
		if row.Symbol == symbol {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}
