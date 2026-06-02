package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/parquet-go/parquet-go"
)

// TushareTool 从本地 parquet 文件读取 A 股数据
type TushareTool struct {
	DataDir string
}

type TushareParams struct {
	Action    string `json:"action"`
	TsCode    string `json:"ts_code,omitempty"`
	Name      string `json:"name,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Indicator string `json:"indicator,omitempty"`
	Period    int    `json:"period,omitempty"`
}

// StockBasic 股票基本信息（从 stock_list.parquet 读取）
type StockBasic struct {
	Code string `parquet:"code"`
	Name string `parquet:"name"`
}

// DailyRow 日线行情（从 daily.parquet 读取）
type DailyRow struct {
	Date             string  `parquet:"date"`
	Symbol           string  `parquet:"symbol"`
	Name             string  `parquet:"name"`
	Open             float64 `parquet:"open"`
	High             float64 `parquet:"high"`
	Low              float64 `parquet:"low"`
	Close            float64 `parquet:"close"`
	Volume           float64 `parquet:"volume"`
	Amount           float64 `parquet:"amount"`
	OutstandingShare float64 `parquet:"outstanding_share"`
	Turnover         float64 `parquet:"turnover"`
}

func NewTushareTool(dataDir string) *TushareTool {
	if dataDir == "" {
		dataDir = "resources"
	}
	return &TushareTool{DataDir: dataDir}
}

func (ts *TushareTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "tushare",
		Desc: `获取中国A股市场数据。支持：查询股票基本信息(get_stock_basic)、获取日线行情(get_daily)、计算技术指标(calc_indicator，支持MACD/KDJ/RSI/MA)。当用户询问股票、需要行情数据或技术分析时使用。
注意：数据仅有20240101至20260601，若超出范围或无法获取，则放弃查询历史数据`,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"action": {
					Type:     "string",
					Desc:     "操作：get_stock_basic(股票信息)/get_daily(日线行情)/calc_indicator(计算指标)",
					Required: true,
				},
				"ts_code": {
					Type: "string",
					Desc: "股票代码，格式如 sh600519 或 sz000001（带交易所前缀），或纯数字如 600519",
				},
				"name": {
					Type: "string",
					Desc: "股票名称，用于模糊搜索",
				},
				"start_date": {
					Type: "string",
					Desc: "开始日期 YYYY-MM-DD 或 YYYYMMDD",
				},
				"end_date": {
					Type: "string",
					Desc: "结束日期 YYYY-MM-DD 或 YYYYMMDD",
				},
				"indicator": {
					Type: "string",
					Desc: "技术指标：MACD/KDJ/RSI/MA",
				},
				"period": {
					Type: "integer",
					Desc: "RSI周期，默认14",
				},
			},
		),
	}, nil
}

func (ts *TushareTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params TushareParams
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("参数解析失败: %w", err)
	}

	switch params.Action {
	case "get_stock_basic":
		return ts.getStockBasic(params)
	case "get_daily":
		return ts.getDaily(params)
	case "calc_indicator":
		return ts.calcIndicator(params)
	default:
		return fmt.Sprintf("未知操作: %s，支持的操作: get_stock_basic, get_daily, calc_indicator", params.Action), nil
	}
}

func (ts *TushareTool) getStockBasic(params TushareParams) (string, error) {
	filePath := filepath.Join(ts.DataDir, "stock_list.parquet")

	rows, err := parquet.ReadFile[StockBasic](filePath)
	if err != nil {
		return "", fmt.Errorf("读取股票列表失败: %w", err)
	}

	// 过滤
	filtered := make([]StockBasic, 0)
	code := normalizeCode(params.TsCode)
	name := params.Name

	for _, row := range rows {
		if code != "" && !strings.Contains(row.Code, code) {
			continue
		}
		if name != "" && !strings.Contains(row.Name, name) {
			continue
		}
		filtered = append(filtered, row)
	}

	if len(filtered) == 0 {
		return "未找到相关股票信息", nil
	}

	// 限制返回数量
	if len(filtered) > 20 {
		filtered = filtered[:20]
	}

	var b strings.Builder
	b.WriteString("## 股票基本信息\n\n")
	for _, row := range filtered {
		symbol := codeToSymbol(row.Code)
		b.WriteString(fmt.Sprintf("- **%s** (%s)\n", row.Name, symbol))
		b.WriteString(fmt.Sprintf("  代码: %s\n", row.Code))
	}
	return b.String(), nil
}

func (ts *TushareTool) getDaily(params TushareParams) (string, error) {
	if params.TsCode == "" {
		return "请提供股票代码 ts_code", nil
	}

	symbol := normalizeSymbol(params.TsCode)
	startDate := normalizeDate(params.StartDate)
	endDate := normalizeDate(params.EndDate)

	rows, err := ts.readDailyData(symbol, startDate, endDate)
	if err != nil {
		return "", err
	}

	if len(rows) == 0 {
		return "未找到行情数据", nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s 日线行情\n\n", symbol))
	b.WriteString("| 日期 | 开盘 | 最高 | 最低 | 收盘 | 成交量 | 成交额 |\n")
	b.WriteString("|------|------|------|------|------|--------|--------|\n")

	for _, row := range rows {
		b.WriteString(fmt.Sprintf("| %s | %.2f | %.2f | %.2f | %.2f | %.0f | %.0f |\n",
			row.Date, row.Open, row.High, row.Low, row.Close, row.Volume, row.Amount))
	}
	return b.String(), nil
}

func (ts *TushareTool) calcIndicator(params TushareParams) (string, error) {
	if params.TsCode == "" {
		return "请提供股票代码 ts_code", nil
	}
	if params.Indicator == "" {
		return "请提供指标类型 indicator: MACD, KDJ, RSI, MA", nil
	}

	symbol := normalizeSymbol(params.TsCode)
	endDate := normalizeDate(params.EndDate)
	startDate := normalizeDate(params.StartDate)

	// 默认取最近 180 天数据
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	if startDate == "" {
		t, _ := time.Parse("2006-01-02", endDate)
		startDate = t.AddDate(0, 0, -180).Format("2006-01-02")
	}

	rows, err := ts.readDailyData(symbol, startDate, endDate)
	if err != nil {
		return "", err
	}

	if len(rows) < 30 {
		return "数据不足，无法计算技术指标（需要至少30条数据）", nil
	}

	// 提取 OHLCV 数据
	closes := make([]float64, len(rows))
	highs := make([]float64, len(rows))
	lows := make([]float64, len(rows))
	dates := make([]string, len(rows))

	for i, row := range rows {
		dates[i] = row.Date
		closes[i] = row.Close
		highs[i] = row.High
		lows[i] = row.Low
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s 技术指标分析 (%s ~ %s)\n\n", symbol, startDate, endDate))

	switch strings.ToUpper(params.Indicator) {
	case "MACD":
		dif, dea, macd := CalcMACD(closes)
		b.WriteString("### MACD (12, 26, 9)\n\n")
		b.WriteString("| 日期 | DIF | DEA | MACD |\n")
		b.WriteString("|------|-----|-----|------|\n")
		start := len(dif) - 20
		if start < 0 {
			start = 0
		}
		for i := start; i < len(dif); i++ {
			b.WriteString(fmt.Sprintf("| %s | %.4f | %.4f | %.4f |\n", dates[i], dif[i], dea[i], macd[i]))
		}
		if len(dif) >= 2 {
			lastDIF := dif[len(dif)-1]
			lastDEA := dea[len(dea)-1]
			lastMACD := macd[len(macd)-1]
			prevMACD := macd[len(macd)-2]
			b.WriteString("\n**分析**：\n")
			if lastDIF > lastDEA && lastMACD > 0 {
				b.WriteString("- 当前处于多头排列（DIF > DEA，MACD > 0）\n")
			} else if lastDIF < lastDEA && lastMACD < 0 {
				b.WriteString("- 当前处于空头排列（DIF < DEA，MACD < 0）\n")
			}
			if lastMACD > 0 && prevMACD <= 0 {
				b.WriteString("- **金叉信号**：MACD 刚由负转正\n")
			} else if lastMACD < 0 && prevMACD >= 0 {
				b.WriteString("- **死叉信号**：MACD 刚由正转负\n")
			}
		}

	case "KDJ":
		k, d, j := CalcKDJ(highs, lows, closes)
		b.WriteString("### KDJ (9, 3, 3)\n\n")
		b.WriteString("| 日期 | K | D | J |\n")
		b.WriteString("|------|---|---|---|\n")
		start := len(k) - 20
		if start < 0 {
			start = 0
		}
		for i := start; i < len(k); i++ {
			b.WriteString(fmt.Sprintf("| %s | %.2f | %.2f | %.2f |\n", dates[i], k[i], d[i], j[i]))
		}
		if len(k) >= 2 {
			lastK := k[len(k)-1]
			lastD := d[len(d)-1]
			lastJ := j[len(j)-1]
			b.WriteString("\n**分析**：\n")
			if lastJ > 80 {
				b.WriteString("- J 值 > 80，处于超买区域\n")
			} else if lastJ < 20 {
				b.WriteString("- J 值 < 20，处于超卖区域\n")
			}
			if lastK > lastD {
				b.WriteString("- K 线在 D 线之上，短期偏多\n")
			} else {
				b.WriteString("- K 线在 D 线之下，短期偏空\n")
			}
		}

	case "RSI":
		period := params.Period
		if period == 0 {
			period = 14
		}
		rsi := CalcRSI(closes, period)
		b.WriteString(fmt.Sprintf("### RSI (%d)\n\n", period))
		b.WriteString("| 日期 | RSI |\n")
		b.WriteString("|------|-----|\n")
		start := len(rsi) - 20
		if start < 0 {
			start = 0
		}
		for i := start; i < len(rsi); i++ {
			b.WriteString(fmt.Sprintf("| %s | %.2f |\n", dates[i], rsi[i]))
		}
		if len(rsi) >= 1 {
			lastRSI := rsi[len(rsi)-1]
			b.WriteString("\n**分析**：\n")
			if lastRSI > 70 {
				b.WriteString(fmt.Sprintf("- RSI = %.2f > 70，处于超买区域，注意回调风险\n", lastRSI))
			} else if lastRSI < 30 {
				b.WriteString(fmt.Sprintf("- RSI = %.2f < 30，处于超卖区域，可能存在反弹机会\n", lastRSI))
			} else {
				b.WriteString(fmt.Sprintf("- RSI = %.2f，处于正常区间\n", lastRSI))
			}
		}

	case "MA":
		ma5 := CalcMA(closes, 5)
		ma10 := CalcMA(closes, 10)
		ma20 := CalcMA(closes, 20)
		ma60 := CalcMA(closes, 60)
		b.WriteString("### 移动平均线 (MA5, MA10, MA20, MA60)\n\n")
		b.WriteString("| 日期 | 收盘价 | MA5 | MA10 | MA20 | MA60 |\n")
		b.WriteString("|------|--------|-----|------|------|------|\n")
		start := len(ma60) - 20
		if start < 0 {
			start = 0
		}
		for i := start; i < len(ma60); i++ {
			b.WriteString(fmt.Sprintf("| %s | %.2f | %.2f | %.2f | %.2f | %.2f |\n",
				dates[i], closes[i], ma5[i], ma10[i], ma20[i], ma60[i]))
		}
		if len(closes) >= 1 {
			lastClose := closes[len(closes)-1]
			b.WriteString("\n**分析**：\n")
			if lastClose > ma5[len(ma5)-1] && lastClose > ma20[len(ma20)-1] {
				b.WriteString("- 股价在 MA5 和 MA20 之上，趋势偏多\n")
			} else if lastClose < ma5[len(ma5)-1] && lastClose < ma20[len(ma20)-1] {
				b.WriteString("- 股价在 MA5 和 MA20 之下，趋势偏空\n")
			}
		}

	default:
		return fmt.Sprintf("不支持的指标类型: %s，支持: MACD, KDJ, RSI, MA", params.Indicator), nil
	}

	return b.String(), nil
}

// readDailyData 从 daily.parquet 读取指定股票和日期范围的数据
func (ts *TushareTool) readDailyData(symbol, startDate, endDate string) ([]DailyRow, error) {
	filePath := filepath.Join(ts.DataDir, "daily.parquet")

	rows, err := parquet.ReadFile[DailyRow](filePath)
	if err != nil {
		return nil, fmt.Errorf("读取日线数据失败: %w", err)
	}

	// 过滤
	filtered := make([]DailyRow, 0)
	for _, row := range rows {
		if row.Symbol != symbol {
			continue
		}
		if startDate != "" && row.Date < startDate {
			continue
		}
		if endDate != "" && row.Date > endDate {
			continue
		}
		filtered = append(filtered, row)
	}

	return filtered, nil
}

// normalizeCode 将用户输入的代码标准化为 parquet 中的 code 格式（纯6位数字）
func normalizeCode(code string) string {
	code = strings.TrimSpace(code)
	code = strings.ToUpper(code)
	// 移除 .SH .SZ 后缀
	code = strings.Replace(code, ".SH", "", 1)
	code = strings.Replace(code, ".SZ", "", 1)
	// 移除 sh sz 前缀
	code = strings.TrimPrefix(code, "SH")
	code = strings.TrimPrefix(code, "SZ")
	return code
}

// normalizeSymbol 将用户输入的代码标准化为 parquet 中的 symbol 格式（sh600519 / sz000001）
func normalizeSymbol(code string) string {
	code = strings.TrimSpace(code)
	code = strings.ToLower(code)
	// 如果已经是 sh/sz 开头，直接返回
	if strings.HasPrefix(code, "sh") || strings.HasPrefix(code, "sz") {
		return code
	}
	// 移除 .sh .sz 后缀
	code = strings.Replace(code, ".sh", "", 1)
	code = strings.Replace(code, ".sz", "", 1)
	// 根据代码判断交易所
	if len(code) == 6 {
		if code[0] == '6' {
			return "sh" + code
		} else if code[0] == '0' || code[0] == '3' {
			return "sz" + code
		}
	}
	return code
}

// codeToSymbol 将 code（纯数字）转为 symbol（带交易所前缀）
func codeToSymbol(code string) string {
	if len(code) == 6 {
		if code[0] == '6' {
			return "sh" + code
		} else if code[0] == '0' || code[0] == '3' {
			return "sz" + code
		}
	}
	return code
}

// normalizeDate 标准化日期格式为 YYYY-MM-DD
func normalizeDate(date string) string {
	date = strings.TrimSpace(date)
	if date == "" {
		return ""
	}
	// YYYYMMDD -> YYYY-MM-DD
	if len(date) == 8 {
		return date[:4] + "-" + date[4:6] + "-" + date[6:]
	}
	return date
}
