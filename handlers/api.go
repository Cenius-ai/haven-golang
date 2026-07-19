package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// --- Market API ---

func (h *Handlers) APIMarketPrices(c *gin.Context) {
	coins, err := h.CG.GetPrices()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "market data unavailable"})
		return
	}
	c.JSON(http.StatusOK, coins)
}

func (h *Handlers) APICoinDetail(c *gin.Context) {
	id := c.Param("id")
	coin, err := h.CG.GetCoin(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "coin not found"})
		return
	}
	c.JSON(http.StatusOK, coin)
}

// --- Portfolio API ---

func (h *Handlers) APIPortfolio(c *gin.Context) {
	c.JSON(http.StatusOK, h.Store.GetHoldings())
}

type addHoldingRequest struct {
	CoinID        string  `json:"coin_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PurchasePrice float64 `json:"purchase_price" binding:"required,gt=0"`
}

func (h *Handlers) APIPortfolioAdd(c *gin.Context) {
	var req addHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: coin_id, amount, and purchase_price are required and must be positive numbers"})
		return
	}

	holding, err := h.Store.AddHolding(req.CoinID, req.Amount, req.PurchasePrice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not add holding"})
		return
	}

	c.JSON(http.StatusCreated, holding)
}

func (h *Handlers) APIPortfolioRemove(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.Store.RemoveHolding(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "holding not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// --- Alerts API ---

func (h *Handlers) APIAlerts(c *gin.Context) {
	c.JSON(http.StatusOK, h.Store.GetAlerts())
}

type createAlertRequest struct {
	CoinID      string  `json:"coin_id" binding:"required"`
	TargetPrice float64 `json:"target_price" binding:"required,gt=0"`
	Direction   string  `json:"direction" binding:"required,oneof=above below"`
}

func (h *Handlers) APIAlertsCreate(c *gin.Context) {
	var req createAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: coin_id, target_price (>0), and direction (above|below) are required"})
		return
	}

	alert, err := h.Store.AddAlert(req.CoinID, req.Direction, req.TargetPrice, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create alert"})
		return
	}

	c.JSON(http.StatusCreated, alert)
}

func (h *Handlers) APIAlertsDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.Store.RemoveAlert(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}
