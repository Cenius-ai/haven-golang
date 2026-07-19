# Haven — Usage Guide

## Dashboard (`/`)

The landing page shows a grid of the top 50 cryptocurrencies by market cap, fetched from CoinGecko. Each card displays:

- Symbol chip (e.g. **BTC**)
- 24-hour price change (green for positive, red for negative)
- Coin name
- Current price in USD
- Market cap (abbreviated: B = billions, M = millions)

Click any coin card to view its detail page. Data refreshes on each page load and is cached for 5 minutes.

## Portfolio (`/portfolio`)

Track your cryptocurrency holdings:

- **Summary cards** — Total portfolio value and total P&L
- **Holdings table** — Each position with amount, entry price, current price, current value, and P&L
- **Add holding** — Use the form at the bottom: select a coin, enter amount and entry price
- **Remove holding** — Click the × button on any row (confirmation required)

Holdings are persisted to `data/holdings.json` and survive restarts.

## Price Alerts (`/alerts`)

Set threshold alerts that trigger visually when the market hits your target:

- **Alert list** — Each alert shows the coin, condition (above/below target), current price, and status
- **Status pills**:
  - **Active** — Monitoring; price hasn't crossed the threshold
  - **Triggered** — Price has crossed the threshold (left border turns red)
  - **Inactive** — Alert is disabled
- **Create alert** — Select a coin, choose direction (above/below), set target price
- **Delete alert** — Click × to remove

Alerts are checked on page load against current prices.

## Coin Detail (`/coins/:id`)

View detailed information for any coin:

- Full name, symbol, and CoinGecko ID
- Current price with 24h change
- Market cap
- Your position (if you hold this coin)
- Breadcrumb navigation back to Dashboard

## Settings (`/settings`)

- **Appearance** — Toggle between light and dark mode. Choice is saved to `localStorage` and persists across sessions and page reloads.
- **Data** — Market data status indicator
- **About** — Version and project info

## Theme

The theme toggle is available in the header on every page. Click the ☀/☾ button or use the toggle on the Settings page. Dark mode uses warm, low-contrast tones inspired by e-ink reading devices.

## API

For programmatic access (all endpoints return JSON):

| Method | Endpoint                      | Description              |
|--------|-------------------------------|--------------------------|
| GET    | `/api/market/prices`          | All coin prices          |
| GET    | `/api/market/coins/:id`       | Single coin detail       |
| GET    | `/api/portfolio`              | List holdings            |
| POST   | `/api/portfolio/add`          | Add holding              |
| DELETE | `/api/portfolio/remove/:id`   | Remove holding           |
| GET    | `/api/alerts`                 | List alerts              |
| POST   | `/api/alerts`                 | Create alert             |
| DELETE | `/api/alerts/:id`             | Delete alert             |
