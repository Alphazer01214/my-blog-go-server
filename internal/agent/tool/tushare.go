package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const tushareAPIURL = "https://api.tushare.pro"

type TushareTool struct {
	ApiToken string
}

type TushareParams struct {
	Action    string `json:"action"`
	TsCode    string `json:"ts_code,omitempty"`
	Name      string `json:"name,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	TradeDate string `json:"trade_date,omitempty"`
	Indicator string `json:"indicator,omitempty"`
	Period    int    `json:"period,omitempty"`
}

type tushareRequest struct {
	APIName string                 `json:"api_name"`
	Token   string                 `json:"token"`
	Params  map[string]interface{} `json:"params"`
	Fields  string                 `json:"fields"`
}

type tushareResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Fields []string                 `json:"fields"`
		Items  [][]interface{}          `json:"items"`
		HasMore bool                    `json:"has_more"`
		Count  int                      `json:"count"`
	} `json:"data"`
}

func NewTushareTool(apiToken string) *TushareTool {
	return &TushareTool{ApiToken: apiToken}
}

func (ts *TushareTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "tushare",
		Desc: "获取中国A股市场数据。支持：查询股票基本信息(get_stock_basic)、获取日线行情(get_daily)、计算技术指标(calc_indicator，支持MACD/KDJ/RSI/MA)。当用户询问股票、需要行情数据或技术分析时使用。",
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"action": {
					Type:     "string",
					Desc:     "操作：get_stock_basic(股票信息)/get_daily(日线行情)/calc_indicator(计算指标)",
					Required: true,
				},
				"ts_code": {
					Type: "string",
					Desc: "股票代码，格式如 000001.SZ 或 600000.SH",
				},
				"name": {
					Type: "string",
					Desc: "股票名称，用于模糊搜索",
				},
				"start_date": {
					Type: "string",
					Desc: "开始日期 YYYYMMDD",
				},
				"end_date": {
					Type: "string",
					Desc: "结束日期 YYYYMMDD",
				},
				"trade_date": {
					Type: "string",
					Desc: "交易日期 YYYYMMDD，用于查询某一天的数据",
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
		return ts.getStockBasic(ctx, params)
	case "get_daily":
		return ts.getDaily(ctx, params)
	case "calc_indicator":
		return ts.calcIndicator(ctx, params)
	default:
		return fmt.Sprintf("未知操作: %s，支持的操作: get_stock_basic, get_daily, calc_indicator", params.Action), nil
	}
}

func (ts *TushareTool) getStockBasic(ctx context.Context, params TushareParams) (string, error) {
	reqParams := map[string]interface{}{
		"list_status": "L",
	}
	if params.TsCode != "" {
		reqParams["ts_code"] = params.TsCode
	}
	if params.Name != "" {
		reqParams["name"] = params.Name
	}

	data, err := ts.callAPI(ctx, "stock_basic", reqParams, "ts_code,symbol,name,area,industry,market,list_date,exchange")
	if err != nil {
		return "", err
	}

	if len(data.Data.Items) == 0 {
		return "未找到相关股票信息", nil
	}

	var b strings.Builder
	b.WriteString("## 股票基本信息\n\n")
	for _, item := range data.Data.Items {
		tsCode := safeString(item, 0, data.Data.Fields)
		symbol := safeString(item, 1, data.Data.Fields)
		name := safeString(item, 2, data.Data.Fields)
		area := safeString(item, 3, data.Data.Fields)
		industry := safeString(item, 4, data.Data.Fields)
		market := safeString(item, 5, data.Data.Fields)
		listDate := safeString(item, 6, data.Data.Fields)
		exchange := safeString(item, 7, data.Data.Fields)

		b.WriteString(fmt.Sprintf("- **%s** (%s)\n", name, tsCode))
		b.WriteString(fmt.Sprintf("  代码: %s | 地区: %s | 行业: %s\n", symbol, area, industry))
		b.WriteString(fmt.Sprintf("  市场: %s | 交易所: %s | 上市日期: %s\n", market, exchange, listDate))
	}
	return b.String(), nil
}

func (ts *TushareTool) getDaily(ctx context.Context, params TushareParams) (string, error) {
	if params.TsCode == "" {
		return "请提供股票代码 ts_code", nil
	}

	reqParams := map[string]interface{}{
		"ts_code": params.TsCode,
	}
	if params.TradeDate != "" {
		reqParams["trade_date"] = params.TradeDate
	}
	if params.StartDate != "" {
		reqParams["start_date"] = params.StartDate
	}
	if params.EndDate != "" {
		reqParams["end_date"] = params.EndDate
	}

	data, err := ts.callAPI(ctx, "daily", reqParams, "ts_code,trade_date,open,high,low,close,pre_close,change,pct_chg,vol,amount")
	if err != nil {
		return "", err
	}

	if len(data.Data.Items) == 0 {
		return "未找到行情数据", nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s 日线行情\n\n", params.TsCode))
	b.WriteString("| 日期 | 开盘 | 最高 | 最低 | 收盘 | 涨跌幅 | 成交量(手) | 成交额(千元) |\n")
	b.WriteString("|------|------|------|------|------|--------|------------|------------|\n")

	for _, item := range data.Data.Items {
		tradeDate := safeString(item, 1, data.Data.Fields)
		open := safeString(item, 2, data.Data.Fields)
		high := safeString(item, 3, data.Data.Fields)
		low := safeString(item, 4, data.Data.Fields)
		close := safeString(item, 5, data.Data.Fields)
		pctChg := safeString(item, 7, data.Data.Fields)
		vol := safeString(item, 8, data.Data.Fields)
		amount := safeString(item, 9, data.Data.Fields)
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s%% | %s | %s |\n",
			tradeDate, open, high, low, close, pctChg, vol, amount))
	}
	return b.String(), nil
}

func (ts *TushareTool) calcIndicator(ctx context.Context, params TushareParams) (string, error) {
	if params.TsCode == "" {
		return "请提供股票代码 ts_code", nil
	}
	if params.Indicator == "" {
		return "请提供指标类型 indicator: MACD, KDJ, RSI, MA", nil
	}

	// 默认取最近 120 天数据
	endDate := params.EndDate
	if endDate == "" {
		endDate = time.Now().Format("20060102")
	}
	startDate := params.StartDate
	if startDate == "" {
		t, _ := time.Parse("20060102", endDate)
		startDate = t.AddDate(0, 0, -180).Format("20060102")
	}

	reqParams := map[string]interface{}{
		"ts_code":   params.TsCode,
		"start_date": startDate,
		"end_date":   endDate,
	}

	data, err := ts.callAPI(ctx, "daily", reqParams, "ts_code,trade_date,open,high,low,close,vol")
	if err != nil {
		return "", err
	}

	if len(data.Data.Items) < 30 {
		return "数据不足，无法计算技术指标", nil
	}

	// 提取 OHLCV 数据
	closes := make([]float64, 0, len(data.Data.Items))
	highs := make([]float64, 0, len(data.Data.Items))
	lows := make([]float64, 0, len(data.Data.Items))
	dates := make([]string, 0, len(data.Data.Items))

	for _, item := range data.Data.Items {
		dates = append(dates, safeString(item, 1, data.Data.Fields))
		closes = append(closes, safeFloat(item, 5, data.Data.Fields))
		highs = append(highs, safeFloat(item, 3, data.Data.Fields))
		lows = append(lows, safeFloat(item, 4, data.Data.Fields))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %s 技术指标分析 (%s ~ %s)\n\n", params.TsCode, startDate, endDate))

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
		// 给出判断
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

func (ts *TushareTool) callAPI(ctx context.Context, apiName string, params map[string]interface{}, fields string) (*tushareResponse, error) {
	reqBody := tushareRequest{
		APIName: apiName,
		Token:   ts.ApiToken,
		Params:  params,
		Fields:  fields,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", tushareAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var tResp tushareResponse
	if err := json.Unmarshal(respBytes, &tResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if tResp.Code != 0 {
		return nil, fmt.Errorf("tushare api error: %s", tResp.Msg)
	}

	return &tResp, nil
}

func safeString(item []interface{}, idx int, fields []string) string {
	if idx >= len(item) || idx >= len(fields) {
		return ""
	}
	if item[idx] == nil {
		return ""
	}
	return fmt.Sprintf("%v", item[idx])
}

func safeFloat(item []interface{}, idx int, fields []string) float64 {
	if idx >= len(item) || idx >= len(fields) {
		return 0
	}
	if item[idx] == nil {
		return 0
	}
	switch v := item[idx].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return 0
	}
}
