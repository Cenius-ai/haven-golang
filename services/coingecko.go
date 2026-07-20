package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"haven/store"
)

const (
	coingeckoBase = "https://api.coingecko.com/api/v3"
	cacheDuration = 5 * time.Minute
	// Short timeout — in sandbox with no network we fail fast and use cache.
	requestTimeout = 4 * time.Second
)

// CoinGecko fetches market data and caches results to a JSON file.
type CoinGecko struct {
	cachePath string
	client    *http.Client
	mu        sync.RWMutex
	cache     *cacheData
}

type cacheData struct {
	FetchedAt time.Time    `json:"fetched_at"`
	Coins     []store.Coin `json:"coins"`
}

func NewCoinGecko(cachePath string) *CoinGecko {
	cg := &CoinGecko{
		cachePath: cachePath,
		client: &http.Client{
			Timeout: requestTimeout,
		},
	}
	cg.loadCache()
	return cg
}

func (cg *CoinGecko) loadCache() {
	f, err := os.Open(cg.cachePath)
	if err != nil {
		log.Printf("coingecko: no cache file at %s: %v", cg.cachePath, err)
		return
	}
	defer f.Close()

	var cd cacheData
	if err := json.NewDecoder(f).Decode(&cd); err != nil {
		log.Printf("coingecko: failed to decode cache: %v", err)
		return
	}
	cg.mu.Lock()
	cg.cache = &cd
	cg.mu.Unlock()
	log.Printf("coingecko: loaded %d coins from cache (age: %s)", len(cd.Coins), time.Since(cd.FetchedAt).Round(time.Second))
}

func (cg *CoinGecko) saveCache() {
	cg.mu.RLock()
	cd := cg.cache
	cg.mu.RUnlock()
	if cd == nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(cg.cachePath), 0o755); err != nil {
		log.Printf("coingecko: mkdir cache: %v", err)
		return
	}
	f, err := os.Create(cg.cachePath)
	if err != nil {
		log.Printf("coingecko: create cache: %v", err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cd); err != nil {
		log.Printf("coingecko: encode cache: %v", err)
	}
}

// GetPrices returns all cached coins. Attempts a live fetch; falls back to cache.
func (cg *CoinGecko) GetPrices() ([]store.Coin, error) {
	// Check if cache is fresh enough
	cg.mu.RLock()
	cached := cg.cache
	cg.mu.RUnlock()

	if cached != nil && len(cached.Coins) > 0 && time.Since(cached.FetchedAt) < cacheDuration {
		return cached.Coins, nil
	}

	// Try live fetch
	coins, err := cg.fetchPrices()
	if err == nil && len(coins) > 0 {
		cg.mu.Lock()
		cg.cache = &cacheData{
			FetchedAt: time.Now(),
			Coins:     coins,
		}
		cg.mu.Unlock()
		cg.saveCache()
		return coins, nil
	}
	if err != nil {
		log.Printf("coingecko: live fetch failed (%v), using cache", err)
	}

	// Fall back to cache
	if cached != nil && len(cached.Coins) > 0 {
		if time.Since(cached.FetchedAt) > 24*time.Hour {
			log.Printf("coingecko: warning — cache is %s old", time.Since(cached.FetchedAt).Round(time.Minute))
		}
		return cached.Coins, nil
	}

	return nil, fmt.Errorf("no live data and no cache available")
}

// GetCoin returns a single coin by ID from the cache.
func (cg *CoinGecko) GetCoin(id string) (store.Coin, error) {
	coins, err := cg.GetPrices()
	if err != nil {
		return store.Coin{}, err
	}
	for _, c := range coins {
		if c.ID == id {
			return c, nil
		}
	}
	return store.Coin{}, fmt.Errorf("coin %q not found", id)
}

// SetCache sets the cache directly (used by seed).
func (cg *CoinGecko) SetCache(coins []store.Coin) {
	cg.mu.Lock()
	cg.cache = &cacheData{
		FetchedAt: time.Now(),
		Coins:     coins,
	}
	cg.mu.Unlock()
	cg.saveCache()
}

func (cg *CoinGecko) fetchPrices() ([]store.Coin, error) {
	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&order=market_cap_desc&per_page=50&page=1&sparkline=false&price_change_percentage=24h", coingeckoBase)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := cg.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coingecko request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko returned %d", resp.StatusCode)
	}

	var coins []store.Coin
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return nil, fmt.Errorf("decode coingecko response: %w", err)
	}

	return coins, nil
}
