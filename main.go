package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"haven/handlers"
	"haven/seed"
	"haven/services"
	"haven/store"

	"github.com/gin-gonic/gin"

	"embed"
	"html/template"
)

//go:embed templates static
var embeddedFS embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Determine data directory
	dataDir := os.Getenv("HAVEN_DATA_DIR")
	if dataDir == "" {
		dataDir = "."
	}

	// Initialise store
	st := store.New(dataDir)
	if err := st.Load(); err != nil {
		log.Printf("store: loading data: %v (will seed fresh)", err)
	}

	// Initialise CoinGecko service with cache
	cachePath := filepath.Join(dataDir, "cache", "coingecko_cache.json")
	cg := services.NewCoinGecko(cachePath)

	// Seed demo data if empty
	if len(st.Holdings) == 0 && len(st.Alerts) == 0 {
		if err := seed.Run(st, cg); err != nil {
			log.Printf("seed: %v", err)
		}
	}

	// Set up Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.LoggerWithWriter(log.Writer()), gin.Recovery())
	r.Use(securityHeaders())

	// Templates
	templatesFS, err := fs.Sub(embeddedFS, "templates")
	if err != nil {
		log.Fatalf("templates sub: %v", err)
	}
	tmpl := template.Must(
		template.New("").Funcs(template.FuncMap{
			"fma": func(a float64) string { return fmt.Sprintf("%.2f", a) },
			"fmc": func(a float64) string {
				if a >= 1_000_000_000_000 {
					return fmt.Sprintf("$%.2fT", a/1_000_000_000_000)
				}
				if a >= 1_000_000_000 {
					return fmt.Sprintf("$%.2fB", a/1_000_000_000)
				}
				if a >= 1_000_000 {
					return fmt.Sprintf("$%.2fM", a/1_000_000)
				}
				return fmt.Sprintf("$%.2f", a)
			},
			"fmpct": func(a float64) string {
				return fmt.Sprintf("%+.2f%%", a)
			},
		}).ParseFS(templatesFS, "*.html", "partials/*.html"),
	)
	r.SetHTMLTemplate(tmpl)

	// Static files
	staticFS, err := fs.Sub(embeddedFS, "static")
	if err != nil {
		log.Fatalf("static sub: %v", err)
	}
	r.StaticFS("/static", http.FS(staticFS))

	// Handler setup
	h := handlers.New(st, cg)

	// Page routes
	r.GET("/", h.Dashboard)
	r.GET("/portfolio", h.PortfolioPage)
	r.GET("/alerts", h.AlertsPage)
	r.GET("/coins/:id", h.CoinPage)
	r.GET("/settings", h.SettingsPage)

	// API routes
	api := r.Group("/api")
	{
		api.GET("/market/prices", h.APIMarketPrices)
		api.GET("/market/coins/:id", h.APICoinDetail)
		api.GET("/portfolio", h.APIPortfolio)
		api.POST("/portfolio/add", h.APIPortfolioAdd)
		api.DELETE("/portfolio/remove/:id", h.APIPortfolioRemove)
		api.GET("/alerts", h.APIAlerts)
		api.POST("/alerts", h.APIAlertsCreate)
		api.DELETE("/alerts/:id", h.APIAlertsDelete)
	}

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})

	log.Printf("Haven starting on 0.0.0.0:%s", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; script-src 'self' 'unsafe-inline'")
		c.Next()
	}
}

// dumpJSON is a helper for writing seed cache
func dumpJSON(path string, v any) error {
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
