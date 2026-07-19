package seed

import (
	"fmt"
	"log"

	"haven/services"
	"haven/store"
)

// Run seeds demo data into the store and CoinGecko cache. Idempotent.
func Run(st *store.Store, cg *services.CoinGecko) error {
	coins := seedCoins()

	// Seed cache so market data is available
	cg.SetCache(coins)

	// Seed holdings if empty
	if len(st.GetHoldings()) == 0 {
		if err := seedHoldings(st); err != nil {
			return fmt.Errorf("seed holdings: %w", err)
		}
		log.Printf("seed: created %d holdings", len(st.GetHoldings()))
	}

	// Seed alerts if empty
	if len(st.GetAlerts()) == 0 {
		if err := seedAlerts(st); err != nil {
			return fmt.Errorf("seed alerts: %w", err)
		}
		log.Printf("seed: created %d alerts", len(st.GetAlerts()))
	}

	return nil
}

func seedCoins() []store.Coin {
	return []store.Coin{
		{ID: "bitcoin", Symbol: "btc", Name: "Bitcoin", CurrentPrice: 67432.18, MarketCap: 1_324_000_000_000, PriceChangePercentage24h: 2.34},
		{ID: "ethereum", Symbol: "eth", Name: "Ethereum", CurrentPrice: 3521.67, MarketCap: 423_000_000_000, PriceChangePercentage24h: -1.12},
		{ID: "solana", Symbol: "sol", Name: "Solana", CurrentPrice: 142.83, MarketCap: 65_000_000_000, PriceChangePercentage24h: 5.67},
		{ID: "cardano", Symbol: "ada", Name: "Cardano", CurrentPrice: 0.4512, MarketCap: 16_000_000_000, PriceChangePercentage24h: -0.89},
		{ID: "polkadot", Symbol: "dot", Name: "Polkadot", CurrentPrice: 7.23, MarketCap: 10_200_000_000, PriceChangePercentage24h: 1.45},
		{ID: "chainlink", Symbol: "link", Name: "Chainlink", CurrentPrice: 14.82, MarketCap: 8_700_000_000, PriceChangePercentage24h: 3.21},
		{ID: "avalanche-2", Symbol: "avax", Name: "Avalanche", CurrentPrice: 35.67, MarketCap: 14_100_000_000, PriceChangePercentage24h: -2.10},
		{ID: "uniswap", Symbol: "uni", Name: "Uniswap", CurrentPrice: 7.89, MarketCap: 5_980_000_000, PriceChangePercentage24h: 0.55},
		{ID: "aave", Symbol: "aave", Name: "Aave", CurrentPrice: 92.45, MarketCap: 1_420_000_000, PriceChangePercentage24h: 4.33},
		{ID: "maker", Symbol: "mkr", Name: "Maker", CurrentPrice: 2845.12, MarketCap: 2_630_000_000, PriceChangePercentage24h: -0.34},
	}
}

func seedHoldings(st *store.Store) error {
	type holding struct {
		CoinID        string
		Amount        float64
		PurchasePrice float64
	}
	items := []holding{
		{CoinID: "bitcoin", Amount: 0.25, PurchasePrice: 42000},
		{CoinID: "ethereum", Amount: 2.5, PurchasePrice: 2100},
		{CoinID: "solana", Amount: 25.0, PurchasePrice: 95},
		{CoinID: "cardano", Amount: 5000.0, PurchasePrice: 0.32},
		{CoinID: "chainlink", Amount: 100.0, PurchasePrice: 11.50},
	}
	for _, item := range items {
		if _, err := st.AddHolding(item.CoinID, item.Amount, item.PurchasePrice); err != nil {
			return err
		}
	}
	return nil
}

func seedAlerts(st *store.Store) error {
	type alert struct {
		CoinID      string
		TargetPrice float64
		Direction   string
		Active      bool
	}
	items := []alert{
		{CoinID: "bitcoin", TargetPrice: 70000, Direction: "above", Active: true},
		{CoinID: "ethereum", TargetPrice: 3000, Direction: "below", Active: true},
		{CoinID: "solana", TargetPrice: 150, Direction: "above", Active: true},
		{CoinID: "cardano", TargetPrice: 0.50, Direction: "above", Active: false},
		{CoinID: "chainlink", TargetPrice: 12, Direction: "below", Active: true},
	}
	for _, item := range items {
		if _, err := st.AddAlert(item.CoinID, item.Direction, item.TargetPrice, item.Active); err != nil {
			return err
		}
	}
	return nil
}
