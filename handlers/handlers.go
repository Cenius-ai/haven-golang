package handlers

import (
	"net/http"

	"haven/services"
	"haven/store"

	"github.com/gin-gonic/gin"
)

// Handlers holds shared dependencies for all HTTP handlers.
type Handlers struct {
	Store *store.Store
	CG    *services.CoinGecko
}

func New(st *store.Store, cg *services.CoinGecko) *Handlers {
	return &Handlers{Store: st, CG: cg}
}

// --- Page helpers ---

type pageData struct {
	Title      string
	Theme      string
	ActivePage string
}

func (h *Handlers) theme(c *gin.Context) string {
	t, err := c.Cookie("theme")
	if err != nil || (t != "light" && t != "dark") {
		return "light"
	}
	return t
}

func (h *Handlers) page(c *gin.Context, title, active string) pageData {
	return pageData{
		Title:      title,
		Theme:      h.theme(c),
		ActivePage: active,
	}
}

// --- Shared view builders --------------------------------------

type holdingView struct {
	store.Holding
	CoinName     string
	CoinSymbol   string
	CurrentPrice float64
	CurrentValue float64
	PnL          float64
	PnLPct       float64
}

type alertView struct {
	store.Alert
	CoinName     string
	CurrentPrice float64
	Triggered    bool
}

type portfolioSummary struct {
	Holdings   []holdingView
	TotalValue float64
	TotalPnL   float64
	TotalPnLPct float64
}

// buildHoldingViews returns enriched holding views plus a summary.
func buildHoldingViews(holdings []store.Holding, coinMap map[string]store.Coin) portfolioSummary {
	var views []holdingView
	totalValue := 0.0
	totalCost := 0.0

	for _, h := range holdings {
		coin, ok := coinMap[h.CoinID]
		if !ok {
			coin = store.Coin{Name: h.CoinID}
		}
		cv := coin.CurrentPrice * h.Amount
		cost := h.PurchasePrice * h.Amount
		pnl := cv - cost
		var pnlPct float64
		if cost > 0 {
			pnlPct = (pnl / cost) * 100
		}
		views = append(views, holdingView{
			Holding:      h,
			CoinName:     coin.Name,
			CoinSymbol:   coin.Symbol,
			CurrentPrice: coin.CurrentPrice,
			CurrentValue: cv,
			PnL:          pnl,
			PnLPct:       pnlPct,
		})
		totalValue += cv
		totalCost += cost
	}

	totalPnL := totalValue - totalCost
	var totalPnLPct float64
	if totalCost > 0 {
		totalPnLPct = (totalPnL / totalCost) * 100
	}
	return portfolioSummary{
		Holdings:    views,
		TotalValue:  totalValue,
		TotalPnL:    totalPnL,
		TotalPnLPct: totalPnLPct,
	}
}

// buildAlertViews returns enriched alert views, marking triggered ones.
func buildAlertViews(alerts []store.Alert, coinMap map[string]store.Coin) []alertView {
	var views []alertView
	for _, a := range alerts {
		coin, ok := coinMap[a.CoinID]
		if !ok {
			coin = store.Coin{Name: a.CoinID}
		}
		triggered := false
		if a.Active {
			if a.Direction == "above" && coin.CurrentPrice >= a.TargetPrice {
				triggered = true
			}
			if a.Direction == "below" && coin.CurrentPrice <= a.TargetPrice {
				triggered = true
			}
		}
		views = append(views, alertView{
			Alert:        a,
			CoinName:     coin.Name,
			CurrentPrice: coin.CurrentPrice,
			Triggered:    triggered,
		})
	}
	return views
}

func coinMapFrom(coins []store.Coin) map[string]store.Coin {
	m := make(map[string]store.Coin, len(coins))
	for _, c := range coins {
		m[c.ID] = c
	}
	return m
}

func filterTriggered(alerts []alertView) []alertView {
	var out []alertView
	for _, a := range alerts {
		if a.Triggered {
			out = append(out, a)
		}
	}
	return out
}

// --- Dashboard (F1) — market + portfolio summary + triggered alerts ---

func (h *Handlers) Dashboard(c *gin.Context) {
	coins, err := h.CG.GetPrices()
	if err != nil {
		c.HTML(http.StatusOK, "dashboard", gin.H{
			"Page":  h.page(c, "Haven — Dashboard", "dashboard"),
			"Error": err.Error(),
		})
		return
	}

	cmap := coinMapFrom(coins)
	holdings := h.Store.GetHoldings()
	portfolio := buildHoldingViews(holdings, cmap)
	alerts := buildAlertViews(h.Store.GetAlerts(), cmap)
	triggered := filterTriggered(alerts)

	// Top coin prices — take the first 5 from the market list for the summary
	topCoins := coins
	if len(topCoins) > 5 {
		topCoins = topCoins[:5]
	}

	c.HTML(http.StatusOK, "dashboard", gin.H{
		"Page":            h.page(c, "Haven — Dashboard", "dashboard"),
		"Coins":           coins,
		"TopCoins":        topCoins,
		"Portfolio":       portfolio,
		"HasHoldings":     len(portfolio.Holdings) > 0,
		"Alerts":          alerts,
		"TriggeredAlerts": triggered,
		"HasTriggered":    len(triggered) > 0,
	})
}

// --- Portfolio page ---

func (h *Handlers) PortfolioPage(c *gin.Context) {
	holdings := h.Store.GetHoldings()
	coins, _ := h.CG.GetPrices()
	cmap := coinMapFrom(coins)
	portfolio := buildHoldingViews(holdings, cmap)

	c.HTML(http.StatusOK, "portfolio", gin.H{
		"Page":        h.page(c, "Haven — Portfolio", "portfolio"),
		"Holdings":    portfolio.Holdings,
		"TotalValue":  portfolio.TotalValue,
		"TotalPnL":    portfolio.TotalPnL,
		"TotalPnLPct": portfolio.TotalPnLPct,
		"Coins":       coins,
	})
}

// --- Alerts page ---

func (h *Handlers) AlertsPage(c *gin.Context) {
	alerts := h.Store.GetAlerts()
	coins, _ := h.CG.GetPrices()
	cmap := coinMapFrom(coins)
	views := buildAlertViews(alerts, cmap)

	c.HTML(http.StatusOK, "alerts", gin.H{
		"Page":   h.page(c, "Haven — Alerts", "alerts"),
		"Alerts": views,
		"Coins":  coins,
	})
}

// --- Coin page ---

func (h *Handlers) CoinPage(c *gin.Context) {
	id := c.Param("id")
	coin, err := h.CG.GetCoin(id)
	if err != nil {
		c.HTML(http.StatusOK, "coin", gin.H{
			"Page":  h.page(c, "Haven — Coin", ""),
			"Error": "Coin not found.",
		})
		return
	}

	holdings := h.Store.GetHoldings()
	var holding *store.Holding
	for _, h := range holdings {
		if h.CoinID == id {
			holding = &h
			break
		}
	}

	c.HTML(http.StatusOK, "coin", gin.H{
		"Page":    h.page(c, "Haven — "+coin.Name, ""),
		"Coin":    coin,
		"Holding": holding,
	})
}

// --- Settings page ---

func (h *Handlers) SettingsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "settings", gin.H{
		"Page":  h.page(c, "Haven — Settings", "settings"),
		"Theme": h.theme(c),
	})
}
