package tool

import "math"

// CalcEMA 计算指数移动平均线
func CalcEMA(data []float64, period int) []float64 {
	if len(data) == 0 || period <= 0 {
		return nil
	}
	result := make([]float64, len(data))
	multiplier := 2.0 / float64(period+1)
	result[0] = data[0]
	for i := 1; i < len(data); i++ {
		result[i] = (data[i]-result[i-1])*multiplier + result[i-1]
	}
	return result
}

// CalcSMA 计算简单移动平均线
func CalcSMA(data []float64, period int) []float64 {
	if len(data) < period {
		return nil
	}
	result := make([]float64, len(data)-period+1)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[0] = sum / float64(period)
	for i := period; i < len(data); i++ {
		sum += data[i] - data[i-period]
		result[i-period+1] = sum / float64(period)
	}
	return result
}

// CalcMA 计算移动平均线，返回与输入等长的结果（前 period-1 个为 0）
func CalcMA(data []float64, period int) []float64 {
	if len(data) < period {
		return make([]float64, len(data))
	}
	result := make([]float64, len(data))
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)
	for i := period; i < len(data); i++ {
		sum += data[i] - data[i-period]
		result[i] = sum / float64(period)
	}
	return result
}

// CalcMACD 计算 MACD (12, 26, 9)
// 返回 DIF, DEA, MACD 三个序列
func CalcMACD(closes []float64) (dif, dea, macd []float64) {
	if len(closes) < 26 {
		return nil, nil, nil
	}

	ema12 := CalcEMA(closes, 12)
	ema26 := CalcEMA(closes, 26)

	dif = make([]float64, len(closes))
	for i := range closes {
		dif[i] = ema12[i] - ema26[i]
	}

	dea = CalcEMA(dif, 9)

	macd = make([]float64, len(closes))
	for i := range closes {
		macd[i] = 2 * (dif[i] - dea[i])
	}

	return dif, dea, macd
}

// CalcKDJ 计算 KDJ (9, 3, 3)
// 返回 K, D, J 三个序列
func CalcKDJ(highs, lows, closes []float64) (k, d, j []float64) {
	n := len(closes)
	if n < 9 || n != len(highs) || n != len(lows) {
		return nil, nil, nil
	}

	rsv := make([]float64, n)
	for i := 8; i < n; i++ {
		highest := math.Inf(-1)
		lowest := math.Inf(1)
		for j := i - 8; j <= i; j++ {
			if highs[j] > highest {
				highest = highs[j]
			}
			if lows[j] < lowest {
				lowest = lows[j]
			}
		}
		if highest == lowest {
			rsv[i] = 50
		} else {
			rsv[i] = (closes[i] - lowest) / (highest - lowest) * 100
		}
	}

	k = make([]float64, n)
	d = make([]float64, n)
	j = make([]float64, n)

	k[8] = 50
	d[8] = 50

	for i := 9; i < n; i++ {
		k[i] = 2.0/3.0*k[i-1] + 1.0/3.0*rsv[i]
		d[i] = 2.0/3.0*d[i-1] + 1.0/3.0*k[i]
		j[i] = 3*k[i] - 2*d[i]
	}

	return k, d, j
}

// CalcRSI 计算 RSI
func CalcRSI(closes []float64, period int) []float64 {
	n := len(closes)
	if n < period+1 || period <= 0 {
		return nil
	}

	rsi := make([]float64, n)

	// 计算第一个 RSI
	avgGain := 0.0
	avgLoss := 0.0
	for i := 1; i <= period; i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			avgGain += change
		} else {
			avgLoss -= change
		}
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	if avgLoss == 0 {
		rsi[period] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - 100/(1+rs)
	}

	// 后续 RSI 使用平滑计算
	for i := period + 1; i < n; i++ {
		change := closes[i] - closes[i-1]
		var gain, loss float64
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - 100/(1+rs)
		}
	}

	return rsi
}
