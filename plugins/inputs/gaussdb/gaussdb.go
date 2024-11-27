package plugin

import (
	"context"
	"database/sql"
	"fmt"
	_ "gitee.com/opengauss/openGauss-connector-go-pq" // GaussDB Driver
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// GaussDBPlugin  Structure of the plugin about GaussDB
type GaussDBPlugin struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Dbname   string `toml:"dbname"`
	ctx      context.Context
	cancel   context.CancelFunc

	Log telegraf.Logger `toml:"-"`
}

func init() {
	inputs.Add("gauss", func() telegraf.Input {
		return &GaussDBPlugin{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Dbname:   "postgres",
		}
	})
}

func (p *GaussDBPlugin) Init() error {
	return nil
}

// SampleConfig Return a default configuration
func (p *GaussDBPlugin) SampleConfig() string {
	return `
  ## GaussDB connection settings
  host = "localhost"
  port = 5432
  user = "postgres"
  password = "password"
  dbname = "postgres"
`
}

// Description Return a short description of the plugin Returns a short description of the plugin
func (p *GaussDBPlugin) Description() string {
	return "Collects metrics from GaussDB"
}

// Start Plugin
func (p *GaussDBPlugin) Start(ctx context.Context) error {
	// Check configuration
	if p.Host == "" || p.Port == 0 || p.User == "" || p.Dbname == "" {
		return fmt.Errorf("invalid configuration: host, port, user, and dbname are required")
	}
	return nil
}

func (p *GaussDBPlugin) Gather(accumulator telegraf.Accumulator) error {
	p.sendMetric(accumulator)
	return nil
}

func (p *GaussDBPlugin) Stop() {
	p.cancel()
}

func (p *GaussDBPlugin) sendMetric(accumulator telegraf.Accumulator) error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.Dbname)

	// connect to GaussDB
	db, err := sql.Open("opengauss", connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to GaussDB: %w", err)
	}
	defer db.Close()

	// Query GaussDB View pg_stat_activity
	rows, err := db.Query("SELECT COUNT(*) FROM pg_stat_activity;")
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Result extraction and addition to accumulator
	var connections int
	if rows.Next() {
		if err := rows.Scan(&connections); err != nil {
			return fmt.Errorf("failed to scan result: %w", err)
		}
		accumulator.AddFields("gaussdb_connections", map[string]interface{}{"count": connections}, nil)
	}

	return nil
}
