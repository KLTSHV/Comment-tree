package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/!burnt!sushi/toml"
)

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalText(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" {
		d.Duration = 0
		return nil
	}
	dd, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = dd
	return nil
}

type Config struct {
	App       AppConfig         `toml:"app"`
	HTTP      HTTPConfig        `toml:"http"`
	DB        DBConfig          `toml:"db"`
	Shortener CommentTreeConfig `toml:"shortener"`
}

type AppConfig struct {
	Env     string `toml:"env"`
	BaseURL string `toml:"base_url"`
}

type HTTPConfig struct {
	Addr            string   `toml:"addr"`
	ReadTimeout     Duration `toml:"read_timeout"`
	WriteTimeout    Duration `toml:"write_timeout"`
	IdleTimeout     Duration `toml:"idle_timeout"`
	ShutdownTimeout Duration `toml:"shutdown_timeout"`
}

type DBConfig struct {
	DSN             string   `toml:"dsn"`
	MaxOpenConns    int      `toml:"max_open_conns"`
	MaxIdleConns    int      `toml:"max_idle_conns"`
	ConnMaxLifetime Duration `toml:"conn_max_lifetime"`
	PingTimeout     Duration `toml:"ping_timeout"`
}

type CommentTreeConfig struct {
}

func Load(path string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("config path is empty")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}

	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("config file not found: %s: %w", abs, err)
	}

	var cfg Config
	if _, err := toml.DecodeFile(abs, &cfg); err != nil {
		return nil, fmt.Errorf("decode toml: %w", err)
	}

	applyDefaults(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(err)
	}
	return cfg
}

func applyDefaults(cfg *Config) {
	if cfg.App.Env == "" {
		cfg.App.Env = "dev"
	}

	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8080"
	}
	if cfg.HTTP.ReadTimeout.Duration == 0 {
		cfg.HTTP.ReadTimeout = Duration{5 * time.Second}
	}
	if cfg.HTTP.WriteTimeout.Duration == 0 {
		cfg.HTTP.WriteTimeout = Duration{10 * time.Second}
	}
	if cfg.HTTP.IdleTimeout.Duration == 0 {
		cfg.HTTP.IdleTimeout = Duration{60 * time.Second}
	}
	if cfg.HTTP.ShutdownTimeout.Duration == 0 {
		cfg.HTTP.ShutdownTimeout = Duration{10 * time.Second}
	}

	if cfg.DB.MaxOpenConns == 0 {
		cfg.DB.MaxOpenConns = 10
	}
	if cfg.DB.MaxIdleConns == 0 {
		cfg.DB.MaxIdleConns = 5
	}
	if cfg.DB.ConnMaxLifetime.Duration == 0 {
		cfg.DB.ConnMaxLifetime = Duration{30 * time.Minute}
	}
	if cfg.DB.PingTimeout.Duration == 0 {
		cfg.DB.PingTimeout = Duration{2 * time.Second}
	}
	// AllowCustom по умолчанию false — если в TOML не задан, остаётся false.
}

func (c *Config) Validate() error {
	env := strings.ToLower(strings.TrimSpace(c.App.Env))
	if env != "dev" && env != "prod" {
		return fmt.Errorf("app.env must be dev or prod, got %q", c.App.Env)
	}

	if strings.TrimSpace(c.App.BaseURL) == "" {
		return errors.New("app.base_url is required (e.g. http://localhost:8080)")
	}
	u, err := url.Parse(c.App.BaseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("app.base_url must be a valid absolute URL, got %q", c.App.BaseURL)
	}

	if strings.TrimSpace(c.HTTP.Addr) == "" {
		return errors.New("http.addr is required")
	}
	if c.HTTP.ShutdownTimeout.Duration <= 0 {
		return errors.New("http.shutdown_timeout must be > 0")
	}

	if strings.TrimSpace(c.DB.DSN) == "" {
		return errors.New("db.dsn is required")
	}
	if c.DB.PingTimeout.Duration <= 0 {
		return errors.New("db.ping_timeout must be > 0")
	}

	return nil
}
