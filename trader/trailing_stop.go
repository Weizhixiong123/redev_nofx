package trader

import (
	"fmt"
	"nofx/logger"
	"nofx/store"
	"time"
)

// TrailingStopManager 移动止损管理器
type TrailingStopManager struct {
	config *store.RiskControlConfig
	// 记录每个持仓的最高价格（用于计算回撤）
	peakPrices map[string]float64 // key: symbol_side
	// 记录每个持仓的初始止损价格
	initialStopLoss map[string]float64
}

// NewTrailingStopManager 创建移动止损管理器
func NewTrailingStopManager(config *store.RiskControlConfig) *TrailingStopManager {
	return &TrailingStopManager{
		config:          config,
		peakPrices:      make(map[string]float64),
		initialStopLoss: make(map[string]float64),
	}
}

// UpdateTrailingStop 更新移动止损
// 返回：(新止损价格, 是否更新, 错误)
func (tsm *TrailingStopManager) UpdateTrailingStop(
	symbol string,
	side string, // "long" or "short"
	entryPrice float64,
	currentPrice float64,
	currentStopLoss float64,
	leverage int,
) (float64, bool, error) {

	// 检查是否启用移动止损
	if !tsm.config.EnableTrailingStop {
		return currentStopLoss, false, nil
	}

	posKey := fmt.Sprintf("%s_%s", symbol, side)

	// 初始化峰值价格
	if _, exists := tsm.peakPrices[posKey]; !exists {
		tsm.peakPrices[posKey] = currentPrice
		tsm.initialStopLoss[posKey] = currentStopLoss
	}

	// 计算当前盈利百分比（基于入场价格）
	var currentProfit float64
	if side == "long" {
		currentProfit = (currentPrice - entryPrice) / entryPrice
	} else { // short
		currentProfit = (entryPrice - currentPrice) / entryPrice
	}

	// 更新峰值价格
	if side == "long" {
		if currentPrice > tsm.peakPrices[posKey] {
			tsm.peakPrices[posKey] = currentPrice
		}
	} else { // short
		if currentPrice < tsm.peakPrices[posKey] {
			tsm.peakPrices[posKey] = currentPrice
		}
	}

	// 判断是否应该调整止损
	if !tsm.shouldAdjustStopLoss(currentProfit) {
		return currentStopLoss, false, nil
	}

	// 获取当前应使用的止损比例
	stopLossRatio := tsm.getStopLossRatio(currentProfit)

	// 计算新的止损价格
	var newStopLoss float64
	if side == "long" {
		// 做多：止损价格 = 当前价格 × (1 - 止损比例)
		newStopLoss = currentPrice * (1 - stopLossRatio)
	} else { // short
		// 做空：止损价格 = 当前价格 × (1 + 止损比例)
		newStopLoss = currentPrice * (1 + stopLossRatio)
	}

	// 检查是否向有利方向移动
	shouldMove := false
	if side == "long" {
		// 做多：新止损价格应该高于当前止损价格
		shouldMove = newStopLoss > currentStopLoss
	} else { // short
		// 做空：新止损价格应该低于当前止损价格
		shouldMove = newStopLoss < currentStopLoss
	}

	if shouldMove {
		// 计算止损距离百分比
		var stopDistance float64
		if side == "long" {
			stopDistance = (currentPrice - newStopLoss) / currentPrice * 100
		} else {
			stopDistance = (newStopLoss - currentPrice) / currentPrice * 100
		}

		logger.Infof("📊 [Trailing Stop] %s %s: 盈利 %.2f%% → 移动止损至 %.4f (距离 %.2f%%)",
			symbol, side, currentProfit*100, newStopLoss, stopDistance)

		return newStopLoss, true, nil
	}

	return currentStopLoss, false, nil
}

// shouldAdjustStopLoss 判断是否应该调整止损
func (tsm *TrailingStopManager) shouldAdjustStopLoss(currentProfit float64) bool {
	// 如果设置了"仅偏移达到后才移动"
	if tsm.config.TrailingOnlyOffsetIsReached {
		// 盈利未达到偏移，不移动
		if currentProfit < tsm.config.TrailingStopPositiveOffset {
			return false
		}
	}

	return true
}

// getStopLossRatio 获取当前应使用的止损比例
func (tsm *TrailingStopManager) getStopLossRatio(currentProfit float64) float64 {
	// 检查是否应该使用正向止损
	if tsm.config.TrailingStopPositive > 0 &&
		currentProfit >= tsm.config.TrailingStopPositiveOffset {
		// 使用正向止损比例
		return tsm.config.TrailingStopPositive
	}

	// 使用基础止损比例
	if tsm.config.MaxStopLossPct > 0 {
		return tsm.config.MaxStopLossPct / 100.0
	}

	// 默认 3%
	return 0.03
}

// CheckStopLossHit 检查是否触发止损
// 返回：(是否触发, 触发类型)
func (tsm *TrailingStopManager) CheckStopLossHit(
	symbol string,
	side string,
	currentPrice float64,
	stopLossPrice float64,
	low float64, // K线最低价（可选）
	high float64, // K线最高价（可选）
) (bool, string) {

	if side == "long" {
		// 做多：价格跌破止损线
		triggerPrice := currentPrice
		if low > 0 {
			triggerPrice = low
		}

		if triggerPrice <= stopLossPrice {
			// 判断是否为移动止损触发
			posKey := fmt.Sprintf("%s_%s", symbol, side)
			if _, exists := tsm.peakPrices[posKey]; exists {
				return true, "TRAILING_STOP_LOSS"
			}
			return true, "STOP_LOSS"
		}
	} else { // short
		// 做空：价格突破止损线向上
		triggerPrice := currentPrice
		if high > 0 {
			triggerPrice = high
		}

		if triggerPrice >= stopLossPrice {
			// 判断是否为移动止损触发
			posKey := fmt.Sprintf("%s_%s", symbol, side)
			if _, exists := tsm.peakPrices[posKey]; exists {
				return true, "TRAILING_STOP_LOSS"
			}
			return true, "STOP_LOSS"
		}
	}

	return false, ""
}

// ClearPosition 清除持仓记录（平仓后调用）
func (tsm *TrailingStopManager) ClearPosition(symbol string, side string) {
	posKey := fmt.Sprintf("%s_%s", symbol, side)
	delete(tsm.peakPrices, posKey)
	delete(tsm.initialStopLoss, posKey)
	logger.Infof("🧹 [Trailing Stop] 清除 %s 移动止损记录", posKey)
}

// GetPositionStats 获取持仓统计信息（用于日志和调试）
func (tsm *TrailingStopManager) GetPositionStats(symbol string, side string, currentPrice float64, entryPrice float64) map[string]interface{} {
	posKey := fmt.Sprintf("%s_%s", symbol, side)

	stats := make(map[string]interface{})
	stats["symbol"] = symbol
	stats["side"] = side
	stats["current_price"] = currentPrice
	stats["entry_price"] = entryPrice

	if peakPrice, exists := tsm.peakPrices[posKey]; exists {
		stats["peak_price"] = peakPrice

		// 计算从峰值的回撤
		var drawdown float64
		if side == "long" {
			drawdown = (peakPrice - currentPrice) / peakPrice
		} else {
			drawdown = (currentPrice - peakPrice) / peakPrice
		}
		stats["drawdown_from_peak"] = drawdown * 100 // 百分比

		// 计算峰值盈利
		var peakProfit float64
		if side == "long" {
			peakProfit = (peakPrice - entryPrice) / entryPrice
		} else {
			peakProfit = (entryPrice - peakPrice) / entryPrice
		}
		stats["peak_profit"] = peakProfit * 100 // 百分比
	}

	if initialSL, exists := tsm.initialStopLoss[posKey]; exists {
		stats["initial_stop_loss"] = initialSL
	}

	return stats
}

// ResetDaily 每日重置（如果需要）
func (tsm *TrailingStopManager) ResetDaily() {
	logger.Info("📅 [Trailing Stop] 每日重置移动止损管理器")
	tsm.peakPrices = make(map[string]float64)
	tsm.initialStopLoss = make(map[string]float64)
}
