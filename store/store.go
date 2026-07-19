package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Coin mirrors the CoinGecko market data shape.
type Coin struct {
	ID                       string  `json:"id"`
	Symbol                   string  `json:"symbol"`
	Name                     string  `json:"name"`
	CurrentPrice             float64 `json:"current_price"`
	MarketCap                float64 `json:"market_cap"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
}

// Holding tracks a user's position in a coin.
type Holding struct {
	ID            int     `json:"id"`
	CoinID        string  `json:"coin_id"`
	Amount        float64 `json:"amount"`
	PurchasePrice float64 `json:"purchase_price"`
}

// Alert is a price threshold alert.
type Alert struct {
	ID          int     `json:"id"`
	CoinID      string  `json:"coin_id"`
	TargetPrice float64 `json:"target_price"`
	Direction   string  `json:"direction"` // "above" or "below"
	Active      bool    `json:"active"`
}

// Store holds portfolio and alert data in memory, persisted to JSON files.
type Store struct {
	mu        sync.RWMutex
	Holdings  []Holding `json:"holdings"`
	Alerts    []Alert   `json:"alerts"`

	nextHoldingID int
	nextAlertID   int
	dataDir       string
}

func New(dataDir string) *Store {
	return &Store{
		dataDir:       dataDir,
		nextHoldingID: 1,
		nextAlertID:   1,
	}
}

// --- Persistence ---

func (s *Store) holdingsPath() string { return filepath.Join(s.dataDir, "data", "holdings.json") }
func (s *Store) alertsPath() string   { return filepath.Join(s.dataDir, "data", "alerts.json") }

func (s *Store) Load() error {
	if err := os.MkdirAll(filepath.Join(s.dataDir, "data"), 0o755); err != nil {
		return err
	}

	if err := s.loadFile(s.holdingsPath(), &s.Holdings); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load holdings: %w", err)
	}
	for _, h := range s.Holdings {
		if h.ID >= s.nextHoldingID {
			s.nextHoldingID = h.ID + 1
		}
	}

	if err := s.loadFile(s.alertsPath(), &s.Alerts); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load alerts: %w", err)
	}
	for _, a := range s.Alerts {
		if a.ID >= s.nextAlertID {
			s.nextAlertID = a.ID + 1
		}
	}

	return nil
}

func (s *Store) Save() error {
	if err := s.saveFile(s.holdingsPath(), s.Holdings); err != nil {
		return err
	}
	return s.saveFile(s.alertsPath(), s.Alerts)
}

func (s *Store) loadFile(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func (s *Store) saveFile(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// --- Holdings ---

func (s *Store) GetHoldings() []Holding {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Holding, len(s.Holdings))
	copy(out, s.Holdings)
	return out
}

func (s *Store) AddHolding(coinID string, amount, purchasePrice float64) (Holding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := Holding{
		ID:            s.nextHoldingID,
		CoinID:        coinID,
		Amount:        amount,
		PurchasePrice: purchasePrice,
	}
	s.nextHoldingID++
	s.Holdings = append(s.Holdings, h)
	if err := s.saveFile(s.holdingsPath(), s.Holdings); err != nil {
		return Holding{}, err
	}
	return h, nil
}

func (s *Store) RemoveHolding(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, h := range s.Holdings {
		if h.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("holding %d not found", id)
	}
	s.Holdings = append(s.Holdings[:idx], s.Holdings[idx+1:]...)
	return s.saveFile(s.holdingsPath(), s.Holdings)
}

// --- Alerts ---

func (s *Store) GetAlerts() []Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Alert, len(s.Alerts))
	copy(out, s.Alerts)
	return out
}

func (s *Store) AddAlert(coinID, direction string, targetPrice float64, active bool) (Alert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a := Alert{
		ID:          s.nextAlertID,
		CoinID:      coinID,
		TargetPrice: targetPrice,
		Direction:   direction,
		Active:      active,
	}
	s.nextAlertID++
	s.Alerts = append(s.Alerts, a)
	if err := s.saveFile(s.alertsPath(), s.Alerts); err != nil {
		return Alert{}, err
	}
	return a, nil
}

func (s *Store) RemoveAlert(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, a := range s.Alerts {
		if a.ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("alert %d not found", id)
	}
	s.Alerts = append(s.Alerts[:idx], s.Alerts[idx+1:]...)
	return s.saveFile(s.alertsPath(), s.Alerts)
}
