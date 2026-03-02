package store

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// StrategyStore strategy storage
type StrategyStore struct {
	db *gorm.DB
}

// Strategy strategy configuration
type Strategy struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	UserID        string    `gorm:"column:user_id;not null;default:'';index" json:"user_id"`
	Name          string    `gorm:"not null" json:"name"`
	Description   string    `gorm:"default:''" json:"description"`
	IsActive      bool      `gorm:"column:is_active;default:false;index" json:"is_active"`
	IsDefault     bool      `gorm:"column:is_default;default:false" json:"is_default"`
	IsPublic      bool      `gorm:"column:is_public;default:false;index" json:"is_public"`    // whether visible in strategy market
	ConfigVisible bool      `gorm:"column:config_visible;default:true" json:"config_visible"` // whether config details are visible
	Config        string    `gorm:"not null;default:'{}'" json:"config"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Strategy) TableName() string { return "strategies" }

// StrategyConfig strategy configuration details (JSON structure)
type StrategyConfig struct {
	// Strategy type: "ai_trading" (default) or "grid_trading"
	StrategyType string `json:"strategy_type,omitempty"`

	// language setting: "zh" for Chinese, "en" for English
	// This determines the language used for data formatting and prompt generation
	Language string `json:"language,omitempty"`
	// coin source configuration
	CoinSource CoinSourceConfig `json:"coin_source"`
	// quantitative data configuration
	Indicators IndicatorConfig `json:"indicators"`
	// custom prompt (appended at the end)
	CustomPrompt string `json:"custom_prompt,omitempty"`
	// risk control configuration
	RiskControl RiskControlConfig `json:"risk_control"`
	// circuit breaker configuration
	CircuitBreakers CircuitBreakerConfig `json:"circuit_breakers"`
	// editable sections of System Prompt
	PromptSections PromptSectionsConfig `json:"prompt_sections,omitempty"`

	// Grid trading configuration (only used when StrategyType == "grid_trading")
	GridConfig *GridStrategyConfig `json:"grid_config,omitempty"`
}

// GridStrategyConfig grid trading specific configuration
type GridStrategyConfig struct {
	// Trading pair (e.g., "BTCUSDT")
	Symbol string `json:"symbol"`
	// Number of grid levels (5-50)
	GridCount int `json:"grid_count"`
	// Total investment in USDT
	TotalInvestment float64 `json:"total_investment"`
	// Leverage (1-20)
	Leverage int `json:"leverage"`
	// Upper price boundary (0 = auto-calculate from ATR)
	UpperPrice float64 `json:"upper_price"`
	// Lower price boundary (0 = auto-calculate from ATR)
	LowerPrice float64 `json:"lower_price"`
	// Use ATR to auto-calculate bounds
	UseATRBounds bool `json:"use_atr_bounds"`
	// ATR multiplier for bound calculation (default 2.0)
	ATRMultiplier float64 `json:"atr_multiplier"`
	// Position distribution: "uniform" | "gaussian" | "pyramid"
	Distribution string `json:"distribution"`
	// Maximum drawdown percentage before emergency exit
	MaxDrawdownPct float64 `json:"max_drawdown_pct"`
	// Stop loss percentage per position
	StopLossPct float64 `json:"stop_loss_pct"`
	// Daily loss limit percentage
	DailyLossLimitPct float64 `json:"daily_loss_limit_pct"`
	// Use maker-only orders for lower fees
	UseMakerOnly bool `json:"use_maker_only"`
	// Enable automatic grid direction adjustment based on box breakouts
	EnableDirectionAdjust bool `json:"enable_direction_adjust"`
	// Direction bias ratio for long_bias/short_bias modes (default 0.7 = 70%/30%)
	DirectionBiasRatio float64 `json:"direction_bias_ratio"`
}

// PromptSectionsConfig editable sections of System Prompt
type PromptSectionsConfig struct {
	// role definition (title + description)
	RoleDefinition string `json:"role_definition,omitempty"`
	// trading frequency awareness
	TradingFrequency string `json:"trading_frequency,omitempty"`
	// entry standards
	EntryStandards string `json:"entry_standards,omitempty"`
	// decision process
	DecisionProcess string `json:"decision_process,omitempty"`
}

// CoinSourceConfig coin source configuration
type CoinSourceConfig struct {
	// source type: "static" | "ai500" | "oi_top" | "oi_low" | "mixed"
	SourceType string `json:"source_type"`
	// static coin list (used when source_type = "static")
	StaticCoins []string `json:"static_coins,omitempty"`
	// excluded coins list (filtered out from all sources)
	ExcludedCoins []string `json:"excluded_coins,omitempty"`
	// whether to use AI500 coin pool
	UseAI500 bool `json:"use_ai500"`
	// AI500 coin pool maximum count
	AI500Limit int `json:"ai500_limit,omitempty"`
	// whether to use OI Top (持仓增加榜，适合做多)
	UseOITop bool `json:"use_oi_top"`
	// OI Top maximum count
	OITopLimit int `json:"oi_top_limit,omitempty"`
	// whether to use OI Low (持仓减少榜，适合做空)
	UseOILow bool `json:"use_oi_low"`
	// OI Low maximum count
	OILowLimit int `json:"oi_low_limit,omitempty"`
	// Note: API URLs are now built automatically using NofxOSAPIKey from IndicatorConfig
}

// IndicatorConfig indicator configuration
type IndicatorConfig struct {
	// K-line configuration
	Klines KlineConfig `json:"klines"`
	// raw kline data (OHLCV) - always enabled, required for AI analysis
	EnableRawKlines bool `json:"enable_raw_klines"`
	// technical indicator switches
	EnableEMA         bool `json:"enable_ema"`
	EnableTEMA        bool `json:"enable_tema"`
	EnableMACD        bool `json:"enable_macd"`
	EnableRSI         bool `json:"enable_rsi"`
	EnableATR         bool `json:"enable_atr"`
	EnableBOLL        bool `json:"enable_boll"` // Bollinger Bands
	EnableVolume      bool `json:"enable_volume"`
	EnableOI          bool `json:"enable_oi"`           // open interest
	EnableFundingRate bool `json:"enable_funding_rate"` // funding rate
	// EMA period configuration
	EMAPeriods []int `json:"ema_periods,omitempty"` // default [20, 50]
	// RSI period configuration
	RSIPeriods []int `json:"rsi_periods,omitempty"` // default [7, 14]
	// TEMA period configuration
	TEMAPeriods []int `json:"tema_periods,omitempty"` // default [9]
	// ATR period configuration
	ATRPeriods []int `json:"atr_periods,omitempty"` // default [14]
	// BOLL period configuration (period, standard deviation multiplier is fixed at 2)
	BOLLPeriods []int `json:"boll_periods,omitempty"` // default [20] - can select multiple timeframes
	// external data sources
	ExternalDataSources []ExternalDataSource `json:"external_data_sources,omitempty"`

	// ========== NofxOS Unified API Configuration ==========
	// Unified API Key for all NofxOS data sources
	NofxOSAPIKey string `json:"nofxos_api_key,omitempty"`

	// quantitative data sources (capital flow, position changes, price changes)
	EnableQuantData    bool `json:"enable_quant_data"`    // whether to enable quantitative data
	EnableQuantOI      bool `json:"enable_quant_oi"`      // whether to show OI data
	EnableQuantNetflow bool `json:"enable_quant_netflow"` // whether to show Netflow data

	// OI ranking data (market-wide open interest increase/decrease rankings)
	EnableOIRanking   bool   `json:"enable_oi_ranking"`             // whether to enable OI ranking data
	OIRankingDuration string `json:"oi_ranking_duration,omitempty"` // duration: 1h, 4h, 24h
	OIRankingLimit    int    `json:"oi_ranking_limit,omitempty"`    // number of entries (default 10)

	// NetFlow ranking data (market-wide fund flow rankings - institution/personal)
	EnableNetFlowRanking   bool   `json:"enable_netflow_ranking"`             // whether to enable NetFlow ranking data
	NetFlowRankingDuration string `json:"netflow_ranking_duration,omitempty"` // duration: 1h, 4h, 24h
	NetFlowRankingLimit    int    `json:"netflow_ranking_limit,omitempty"`    // number of entries (default 10)

	// Price ranking data (market-wide gainers/losers)
	EnablePriceRanking   bool   `json:"enable_price_ranking"`             // whether to enable price ranking data
	PriceRankingDuration string `json:"price_ranking_duration,omitempty"` // durations: "1h" or "1h,4h,24h"
	PriceRankingLimit    int    `json:"price_ranking_limit,omitempty"`    // number of entries per ranking (default 10)
}

// KlineConfig K-line configuration
type KlineConfig struct {
	// primary timeframe: "1m", "3m", "5m", "15m", "1h", "4h"
	PrimaryTimeframe string `json:"primary_timeframe"`
	// primary timeframe K-line count
	PrimaryCount int `json:"primary_count"`
	// longer timeframe
	LongerTimeframe string `json:"longer_timeframe,omitempty"`
	// longer timeframe K-line count
	LongerCount int `json:"longer_count,omitempty"`
	// whether to enable multi-timeframe analysis
	EnableMultiTimeframe bool `json:"enable_multi_timeframe"`
	// selected timeframe list (new: supports multi-timeframe selection)
	SelectedTimeframes []string `json:"selected_timeframes,omitempty"`
}

// ExternalDataSource external data source configuration
type ExternalDataSource struct {
	Name        string            `json:"name"`   // data source name
	Type        string            `json:"type"`   // type: "api" | "webhook"
	URL         string            `json:"url"`    // API URL
	Method      string            `json:"method"` // HTTP method
	Headers     map[string]string `json:"headers,omitempty"`
	DataPath    string            `json:"data_path,omitempty"`    // JSON data path
	RefreshSecs int               `json:"refresh_secs,omitempty"` // refresh interval (seconds)
}

// RiskControlConfig risk control configuration
type RiskControlConfig struct {
	// Max number of coins held simultaneously (CODE ENFORCED)
	MaxPositions int `json:"max_positions"`

	// BTC/ETH exchange leverage for opening positions (AI guided)
	BTCETHMaxLeverage int `json:"btc_eth_max_leverage"`
	// Altcoin exchange leverage for opening positions (AI guided)
	AltcoinMaxLeverage int `json:"altcoin_max_leverage"`

	// BTC/ETH single position max value = equity × this ratio (CODE ENFORCED, default: 5)
	BTCETHMaxPositionValueRatio float64 `json:"btc_eth_max_position_value_ratio"`
	// Altcoin single position max value = equity × this ratio (CODE ENFORCED, default: 1)
	AltcoinMaxPositionValueRatio float64 `json:"altcoin_max_position_value_ratio"`

	// Max margin utilization (e.g. 0.9 = 90%) (CODE ENFORCED)
	MaxMarginUsage float64 `json:"max_margin_usage"`
	// Min position size in USDT (CODE ENFORCED)
	MinPositionSize float64 `json:"min_position_size"`

	// Min take_profit / stop_loss ratio (AI guided)
	MinRiskRewardRatio float64 `json:"min_risk_reward_ratio"`
	// Min AI confidence to open position (AI guided)
	MinConfidence int `json:"min_confidence"`

	// === 新增：针对"盈小亏大"和"过度交易"的风控参数 ===

	// Max stop-loss percentage per position (CODE ENFORCED, default: 3.0 means -3%)
	// 当持仓亏损超过该比例时，代码强制平仓，不依赖 AI 判断
	MaxStopLossPct float64 `json:"max_stop_loss_pct"`

	// Min take-profit percentage before closing (AI GUIDED, default: 1.5 means +1.5%)
	// AI 在盈利未达到该比例前，不应主动平仓（除非触发跟踪止盈）
	MinTakeProfitPct float64 `json:"min_take_profit_pct"`

	// Same symbol cooldown in minutes after closing (CODE ENFORCED, default: 10)
	// 平仓后在该冷却期内，不允许对同一标的重新开仓，防止冲动追单
	SameSymbolCooldownMin int `json:"same_symbol_cooldown_min"`

	// Max trades per symbol per day (CODE ENFORCED, default: 3)
	// 每个标的每天最多交易次数（开仓次数），防止同一标的过度交易
	// 当 DailyRiskBudgetPct > 0 时，此字段作为补充保险仍然生效
	MaxTradesPerSymbolPerDay int `json:"max_trades_per_symbol_per_day"`

	// === 移动止损/止盈配置 (Trailing Stop) ===

	// Enable trailing stop (CODE ENFORCED, default: false)
	// 启用移动止损功能
	EnableTrailingStop bool `json:"enable_trailing_stop"`

	// Trailing stop positive percentage (default: 0.02 means 2%)
	// 正向止损比例：当盈利达到偏移后，使用此比例作为止损距离
	TrailingStopPositive float64 `json:"trailing_stop_positive"`

	// Trailing stop positive offset percentage (default: 0.03 means 3%)
	// 正向止损触发偏移：盈利达到此比例后，才切换到正向止损
	TrailingStopPositiveOffset float64 `json:"trailing_stop_positive_offset"`

	// Trailing only when offset is reached (default: false)
	// 仅在偏移达到后才开始移动止损，否则止损固定不动
	TrailingOnlyOffsetIsReached bool `json:"trailing_only_offset_is_reached"`

	// === 盈利池（Profit Pool）风控参数 ===

	// DailyRiskBudgetPct: 每标的每日初始风险预算 = 账户净值 × 此比例 (CODE ENFORCED)
	// 例如：0.01 表示净值的 1%；200 USDT 账户 → 每标的 2 USDT 风险额度
	// 设为 0 则退回传统次数限制（MaxTradesPerSymbolPerDay）
	DailyRiskBudgetPct float64 `json:"daily_risk_budget_pct"`

	// ProfitReinvestRate: 每次盈利平仓后，将利润的此比例补充回当日风险预算 (CODE ENFORCED)
	// 例如：0.30 表示利润的 30% 归还预算，允许在赚钱后继续交易同一标的
	ProfitReinvestRate float64 `json:"profit_reinvest_rate"`
}

// CircuitBreakerConfig circuit breaker configuration
// 熔断机制配置：当市场条件不适合时自动暂停开仓
type CircuitBreakerConfig struct {
	// === 1. StoplossGuard（止损熔断）===
	// 指定时间窗口内止损次数达到阈值，暂停该交易对开仓
	StoplossGuardEnabled    bool `json:"stoploss_guard_enabled"`
	StoplossGuardLookback   int  `json:"stoploss_guard_lookback_min"` // 回看时间窗口（分钟），default: 60
	StoplossGuardCount      int  `json:"stoploss_guard_count"`        // 止损次数阈值，default: 2
	StoplossGuardTradeLimit int  `json:"stoploss_guard_trade_limit"`  // 最少样本数，default: 1
	StoplossGuardDuration   int  `json:"stoploss_guard_duration_min"` // 熔断时长（分钟），default: 60
	StoplossGuardGlobal     bool `json:"stoploss_guard_global"`       // true=全局熔断，false=仅针对该交易对

	// === 2. MaxDrawdownProtection（最大回撤熔断）===
	// 累计最大回撤超过阈值，全局暂停开仓
	MaxDrawdownEnabled  bool    `json:"max_drawdown_enabled"`
	MaxDrawdownPct      float64 `json:"max_drawdown_pct"`          // 最大回撤比例（%），default: 20.0
	MaxDrawdownDuration int     `json:"max_drawdown_duration_min"` // 熔断时长（分钟），default: 240

	// === 3. LowProfitPairs（低盈利熔断）===
	// 指定交易对在时间窗口内均盈利率低于阈值，暂停该交易对
	LowProfitEnabled    bool    `json:"low_profit_enabled"`
	LowProfitLookback   int     `json:"low_profit_lookback_min"` // 回看时间窗口（分钟），default: 1440（24h）
	LowProfitTradeLimit int     `json:"low_profit_trade_limit"`  // 最少样本数，default: 3
	LowProfitMinAvgPct  float64 `json:"low_profit_min_avg_pct"`  // 平均盈利率下限（USDT），default: -0.5
	LowProfitDuration   int     `json:"low_profit_duration_min"` // 熔断时长（分钟），default: 120

	// === 4. CooldownPeriod（冷却期）===
	// 每次交易完成后，自动锁定该交易对一段时间
	// 注意：与 RiskControl.SameSymbolCooldownMin 作用相同，此处提供统一管理入口
	// 建议只使用其中一个，两个同时启用则取较大值
	CooldownEnabled  bool `json:"cooldown_enabled"`
	CooldownDuration int  `json:"cooldown_duration_min"` // 冷却时长（分钟），default: 30
}

// NewStrategyStore creates a new StrategyStore
func NewStrategyStore(db *gorm.DB) *StrategyStore {
	return &StrategyStore{db: db}
}

func (s *StrategyStore) initTables() error {
	// AutoMigrate will add missing columns without dropping existing data
	return s.db.AutoMigrate(&Strategy{})
}

func (s *StrategyStore) initDefaultData() error {
	// No longer pre-populate strategies - create on demand when user configures
	return nil
}

// GetDefaultStrategyConfig returns the default strategy configuration for the given language
func GetDefaultStrategyConfig(lang string) StrategyConfig {
	// Normalize language to "zh" or "en"
	normalizedLang := "en"
	if lang == "zh" {
		normalizedLang = "zh"
	}

	config := StrategyConfig{
		Language: normalizedLang,
		CoinSource: CoinSourceConfig{
			SourceType: "ai500",
			UseAI500:   true,
			AI500Limit: 10,
			UseOITop:   false,
			OITopLimit: 10,
			UseOILow:   false,
			OILowLimit: 10,
		},
		Indicators: IndicatorConfig{
			Klines: KlineConfig{
				PrimaryTimeframe:     "5m",
				PrimaryCount:         30,
				LongerTimeframe:      "4h",
				LongerCount:          10,
				EnableMultiTimeframe: true,
				SelectedTimeframes:   []string{"5m", "15m", "1h", "4h"},
			},
			EnableRawKlines:   true, // Required - raw OHLCV data for AI analysis
			EnableEMA:         false,
			EnableMACD:        false,
			EnableRSI:         false,
			EnableATR:         false,
			EnableBOLL:        false,
			EnableVolume:      true,
			EnableOI:          true,
			EnableFundingRate: true,
			EMAPeriods:        []int{20, 50},
			RSIPeriods:        []int{7, 14},
			ATRPeriods:        []int{14},
			BOLLPeriods:       []int{20},
			// NofxOS unified API key
			NofxOSAPIKey: "cm_568c67eae410d912c54c",
			// Quant data
			EnableQuantData:    true,
			EnableQuantOI:      true,
			EnableQuantNetflow: true,
			// OI ranking data
			EnableOIRanking:   true,
			OIRankingDuration: "1h",
			OIRankingLimit:    10,
			// NetFlow ranking data
			EnableNetFlowRanking:   true,
			NetFlowRankingDuration: "1h",
			NetFlowRankingLimit:    10,
			// Price ranking data
			EnablePriceRanking:   true,
			PriceRankingDuration: "1h,4h,24h",
			PriceRankingLimit:    10,
		},
		RiskControl: RiskControlConfig{
			MaxPositions:                 3,   // Max 3 coins simultaneously (CODE ENFORCED)
			BTCETHMaxLeverage:            5,   // BTC/ETH exchange leverage (AI guided)
			AltcoinMaxLeverage:           5,   // Altcoin exchange leverage (AI guided)
			BTCETHMaxPositionValueRatio:  5.0, // BTC/ETH: max position = 5x equity (CODE ENFORCED)
			AltcoinMaxPositionValueRatio: 1.0, // Altcoin: max position = 1x equity (CODE ENFORCED)
			MaxMarginUsage:               0.9, // Max 90% margin usage (CODE ENFORCED)
			MinPositionSize:              12,  // Min 12 USDT per position (CODE ENFORCED)
			MinRiskRewardRatio:           2.0, // Min 2:1 profit/loss ratio (AI guided)
			MinConfidence:                75,  // Min 75% confidence (AI guided)
			// 新增：防止"盈小亏大"和"过度交易"
			MaxStopLossPct:           3.0, // 亏损超 3% 强制止损 (CODE ENFORCED)
			MinTakeProfitPct:         1.5, // 至少 1.5% 盈利才止盈 (AI GUIDED)
			SameSymbolCooldownMin:    10,  // 平仓后冷却 10 分钟 (CODE ENFORCED)
			MaxTradesPerSymbolPerDay: 3,   // 每标的每天最多 3 次开仓 (CODE ENFORCED)
			// 移动止损配置 (默认禁用)
			EnableTrailingStop:          false, // 禁用移动止损
			TrailingStopPositive:        0.02,  // 正向止损 2%
			TrailingStopPositiveOffset:  0.03,  // 偏移 3% 后触发
			TrailingOnlyOffsetIsReached: false, // 立即开始移动
			// 盈利池（Profit Pool）
			DailyRiskBudgetPct: 0.20, // 每标的每日初始风险预算 = 净值的 20%（如 53U 账户 → 每币 10.6U）
			ProfitReinvestRate: 0.30, // 盈利的 30% 归还预算（亏钱不补充）
		},
		CircuitBreakers: CircuitBreakerConfig{
			// 1. StoplossGuard：60分钟内止损2次 → 熔断60分钟
			StoplossGuardEnabled:    true,
			StoplossGuardLookback:   60,
			StoplossGuardCount:      2,
			StoplossGuardTradeLimit: 1,
			StoplossGuardDuration:   60,
			StoplossGuardGlobal:     false,
			// 2. MaxDrawdownProtection：最大回撤达20% → 全局熔断4小时
			MaxDrawdownEnabled:  true,
			MaxDrawdownPct:      20.0,
			MaxDrawdownDuration: 240,
			// 3. LowProfitPairs：24小时内≥3笔，均亏损<-0.5U → 熔断2小时
			LowProfitEnabled:    true,
			LowProfitLookback:   1440,
			LowProfitTradeLimit: 3,
			LowProfitMinAvgPct:  -0.5,
			LowProfitDuration:   120,
			// 4. CooldownPeriod：默认关闭，由RiskControl.SameSymbolCooldownMin统一控制
			CooldownEnabled:  false,
			CooldownDuration: 30,
		},
	}

	if lang == "zh" {
		config.PromptSections = PromptSectionsConfig{
			RoleDefinition: `# 你是一个专业的加密货币交易AI

你的任务是根据提供的市场数据做出交易决策。你是一个经验丰富的量化交易员，擅长技术分析和风险管理。`,
			TradingFrequency: `# ⏱️ 交易频率意识

- 优秀交易员：每天2-4笔 ≈ 每小时0.1-0.2笔
- 每小时超过2笔 = 过度交易
- 单笔持仓时间 ≥ 30-60分钟
如果你发现自己每个周期都在交易 → 标准太低；如果持仓不到30分钟就平仓 → 太冲动。`,
			EntryStandards: `# 🎯 入场标准（严格）

只在多个信号共振时入场。自由使用任何有效的分析方法，避免单一指标、信号矛盾、横盘震荡、或平仓后立即重新开仓等低质量行为。

**🚫 禁止区间内震荡开单（最重要规则）**
- 开仓前必须先识别近期价格区间：观察最近 20-30 根K线的高点和低点，确定震荡上轨和下轨
- **只有价格明确突破区间**（收盘价站稳区间边界之外 + 成交量放大）才允许顺势开仓
- **区间内的小反弹、小回调一律不开单**，即使短期看起来有机会也不开
- 判断依据：如果最近价格一直在某个窄幅区间来回震荡（如 ±5% 内），说明还在震荡，等突破再说`,
			DecisionProcess: `# 📋 决策流程

1. 检查持仓 → 是否止盈/止损
2. 扫描候选币种 + 多时间框架 → 是否存在强信号
3. 先写思维链，再输出结构化JSON`,
		}
	} else {
		config.PromptSections = PromptSectionsConfig{
			RoleDefinition: `# You are a professional cryptocurrency trading AI

Your task is to make trading decisions based on the provided market data. You are an experienced quantitative trader skilled in technical analysis and risk management.`,
			TradingFrequency: `# ⏱️ Trading Frequency Awareness

- Excellent trader: 2-4 trades per day ≈ 0.1-0.2 trades per hour
- >2 trades per hour = overtrading
- Single position holding time ≥ 30-60 minutes
If you find yourself trading every cycle → standards are too low; if closing positions in <30 minutes → too impulsive.`,
			EntryStandards: `# 🎯 Entry Standards (Strict)

Only enter positions when multiple signals resonate. Freely use any effective analysis methods, avoid low-quality behaviors such as single indicators, contradictory signals, sideways oscillation, or immediately restarting after closing positions.

**🚫 NEVER trade bounces inside a consolidation range (Most Important Rule)**
- Before opening any position, identify the recent price range: look at the high and low of the last 20-30 candles to define the oscillation ceiling and floor
- **Only open positions when price clearly BREAKS OUT of the range** (closing price confirmed beyond range boundary + volume surge)
- **Small bounces or pullbacks WITHIN the range are NOT tradeable** — do not open even if it looks tempting
- How to judge: if price has been bouncing within a narrow band (e.g., within ±5%) recently, it is still oscillating — wait for the breakout`,
			DecisionProcess: `# 📋 Decision Process

1. Check positions → whether to take profit/stop loss
2. Scan candidate coins + multi-timeframe → whether strong signals exist
3. Write chain of thought first, then output structured JSON`,
		}
	}

	return config
}

// Create create a strategy
func (s *StrategyStore) Create(strategy *Strategy) error {
	return s.db.Create(strategy).Error
}

// Update update a strategy
func (s *StrategyStore) Update(strategy *Strategy) error {
	return s.db.Model(&Strategy{}).
		Where("id = ? AND user_id = ?", strategy.ID, strategy.UserID).
		Updates(map[string]interface{}{
			"name":           strategy.Name,
			"description":    strategy.Description,
			"config":         strategy.Config,
			"is_public":      strategy.IsPublic,
			"config_visible": strategy.ConfigVisible,
			"updated_at":     time.Now().UTC(),
		}).Error
}

// Delete delete a strategy
func (s *StrategyStore) Delete(userID, id string) error {
	// do not allow deleting system default strategy
	var st Strategy
	if err := s.db.Where("id = ?", id).First(&st).Error; err == nil && st.IsDefault {
		return fmt.Errorf("cannot delete system default strategy")
	}

	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Strategy{}).Error
}

// List get user's strategy list
func (s *StrategyStore) List(userID string) ([]*Strategy, error) {
	var strategies []*Strategy
	err := s.db.Where("user_id = ? OR is_default = ?", userID, true).
		Order("is_default DESC, created_at DESC").
		Find(&strategies).Error
	if err != nil {
		return nil, err
	}
	return strategies, nil
}

// ListPublic get all public strategies for the strategy market
func (s *StrategyStore) ListPublic() ([]*Strategy, error) {
	var strategies []*Strategy
	err := s.db.Where("is_public = ?", true).
		Order("created_at DESC").
		Find(&strategies).Error
	if err != nil {
		return nil, err
	}
	return strategies, nil
}

// Get get a single strategy
func (s *StrategyStore) Get(userID, id string) (*Strategy, error) {
	var st Strategy
	err := s.db.Where("id = ? AND (user_id = ? OR is_default = ?)", id, userID, true).
		First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// GetActive get user's currently active strategy
func (s *StrategyStore) GetActive(userID string) (*Strategy, error) {
	var st Strategy
	err := s.db.Where("user_id = ? AND is_active = ?", userID, true).First(&st).Error
	if err == gorm.ErrRecordNotFound {
		// no active strategy, return system default strategy
		return s.GetDefault()
	}
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// GetDefault get system default strategy
func (s *StrategyStore) GetDefault() (*Strategy, error) {
	var st Strategy
	err := s.db.Where("is_default = ?", true).First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// SetActive set active strategy (will first deactivate other strategies)
func (s *StrategyStore) SetActive(userID, strategyID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// first deactivate all strategies for the user
		if err := tx.Model(&Strategy{}).Where("user_id = ?", userID).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// activate specified strategy
		return tx.Model(&Strategy{}).
			Where("id = ? AND (user_id = ? OR is_default = ?)", strategyID, userID, true).
			Update("is_active", true).Error
	})
}

// Duplicate duplicate a strategy (used to create custom strategy based on default strategy)
func (s *StrategyStore) Duplicate(userID, sourceID, newID, newName string) error {
	// get source strategy
	source, err := s.Get(userID, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get source strategy: %w", err)
	}

	// create new strategy
	newStrategy := &Strategy{
		ID:          newID,
		UserID:      userID,
		Name:        newName,
		Description: "Created based on [" + source.Name + "]",
		IsActive:    false,
		IsDefault:   false,
		Config:      source.Config,
	}

	return s.Create(newStrategy)
}

// ParseConfig parse strategy configuration JSON
func (s *Strategy) ParseConfig() (*StrategyConfig, error) {
	var config StrategyConfig
	if err := json.Unmarshal([]byte(s.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to parse strategy configuration: %w", err)
	}
	return &config, nil
}

// SetConfig set strategy configuration
func (s *Strategy) SetConfig(config *StrategyConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize strategy configuration: %w", err)
	}
	s.Config = string(data)
	return nil
}
