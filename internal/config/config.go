package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
)

// Config is the root configuration model for this project.
//
// Hard safety constraint:
// This repo is designed to automate ONLY the locally-hosted mock app.
// Validation enforces a localhost base URL.
type Config struct {
	Mocknet MocknetConfig `koanf:"mocknet"`
	Auth    AuthConfig    `koanf:"auth"`
	Storage StorageConfig `koanf:"storage"`
	Browser BrowserConfig `koanf:"browser"`
	Logging LoggingConfig `koanf:"logging"`
	Run     RunConfig     `koanf:"run"`
}

type MocknetConfig struct {
	BaseURL string `koanf:"base_url"`
	Port    int    `koanf:"port"`
	BrandName string `koanf:"brand_name"`
}

type AuthConfig struct {
	Username string `koanf:"username"`
	Password string `koanf:"password"`
}

type StorageConfig struct {
	SQLitePath string `koanf:"sqlite_path"`
}

type BrowserConfig struct {
	Headless bool          `koanf:"headless"`
	SlowMo   time.Duration `koanf:"slow_mo"`
}

type LoggingConfig struct {
	Level string `koanf:"level"`
	JSON  bool   `koanf:"json"`
}

type RunConfig struct {
	Deterministic bool          `koanf:"deterministic"`
	Seed          int64         `koanf:"seed"`
	Timeout       time.Duration `koanf:"timeout"`
}

func Default() Config {
	return Config{
		Mocknet: MocknetConfig{
			BaseURL: "http://localhost:8080",
			Port:    8080,
			BrandName: "Mock Professional Network",
		},
		Auth: AuthConfig{},
		Storage: StorageConfig{
			SQLitePath: filepath.FromSlash("data/subspace.db"),
		},
		Browser: BrowserConfig{
			Headless: true,
			SlowMo:   0,
		},
		Logging: LoggingConfig{
			Level: "info",
			JSON:  true,
		},
		Run: RunConfig{
			Deterministic: false,
			Seed:          1,
			Timeout:       45 * time.Second,
		},
	}
}

// Load reads configuration from a YAML file (optional) and environment variables.
//
// Precedence (highest first):
//  1) Environment variables
//  2) Config file
//  3) Defaults
func Load(configPath string) (Config, error) {
	k := koanf.New(".")

	// Defaults first.
	if err := k.Load(structs.Provider(Default(), "koanf"), nil); err != nil {
		return Config{}, fmt.Errorf("load defaults: %w", err)
	}

	// Optional YAML file.
	if strings.TrimSpace(configPath) != "" {
		if _, err := os.Stat(configPath); err != nil {
			return Config{}, fmt.Errorf("stat config file: %w", err)
		}
		if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
			return Config{}, fmt.Errorf("load config file: %w", err)
		}
	}

	// Environment overrides using SUBSPACE_ prefix.
	//	- Use double-underscore for nesting: SUBSPACE_LOGGING__LEVEL=debug
	//	- For a few common fields we also accept single-underscore convenience env vars.
	applyEnvOverrides(k)

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}


func applyEnvOverrides(k *koanf.Koanf) {
	if v := os.Getenv("SUBSPACE_AUTH_USERNAME"); strings.TrimSpace(v) != "" {
		_ = k.Set("auth.username", v)
	}
	if v := os.Getenv("SUBSPACE_AUTH_PASSWORD"); strings.TrimSpace(v) != "" {
		_ = k.Set("auth.password", v)
	}
	if v := os.Getenv("SUBSPACE_MOCKNET_BASE_URL"); strings.TrimSpace(v) != "" {
		_ = k.Set("mocknet.base_url", v)
	}
	if v := os.Getenv("SUBSPACE_STORAGE_SQLITE_PATH"); strings.TrimSpace(v) != "" {
		_ = k.Set("storage.sqlite_path", v)
	}

	for _, kv := range os.Environ() {
		key, val, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		if !strings.HasPrefix(key, "SUBSPACE_") {
			continue
		}
		// Convenience vars are handled above; the nested form requires '__'.
		if !strings.Contains(key, "__") {
			continue
		}
		if strings.TrimSpace(val) == "" {
			continue
		}

		cfgKey := strings.TrimPrefix(key, "SUBSPACE_")
		cfgKey = strings.ToLower(cfgKey)
		cfgKey = strings.ReplaceAll(cfgKey, "__", ".")

		// Parse known types to avoid decode surprises.
		switch cfgKey {
		case "browser.headless":
			if b, err := strconv.ParseBool(val); err == nil {
				_ = k.Set(cfgKey, b)
				continue
			}
		case "browser.slow_mo":
			if d, err := time.ParseDuration(val); err == nil {
				_ = k.Set(cfgKey, d)
				continue
			}
		case "mocknet.port":
			if p, err := strconv.Atoi(val); err == nil {
				_ = k.Set(cfgKey, p)
				continue
			}
		case "logging.json":
			if b, err := strconv.ParseBool(val); err == nil {
				_ = k.Set(cfgKey, b)
				continue
			}
		case "run.deterministic":
			if b, err := strconv.ParseBool(val); err == nil {
				_ = k.Set(cfgKey, b)
				continue
			}
		case "run.seed":
			if n, err := strconv.ParseInt(val, 10, 64); err == nil {
				_ = k.Set(cfgKey, n)
				continue
			}
		case "run.timeout":
			if d, err := time.ParseDuration(val); err == nil {
				_ = k.Set(cfgKey, d)
				continue
			}
		}

		_ = k.Set(cfgKey, val)
	}
}

var (
	localhostHostRe = regexp.MustCompile(`^(localhost|127\.0\.0\.1)(:\d+)?$`)
)

func Validate(cfg Config) error {
	var problems []string

	if strings.TrimSpace(cfg.Mocknet.BaseURL) == "" {
		problems = append(problems, "mocknet.base_url is required")
	} else {
		u, err := url.Parse(cfg.Mocknet.BaseURL)
		if err != nil {
			problems = append(problems, fmt.Sprintf("mocknet.base_url is invalid: %v", err))
		} else {
			if u.Scheme != "http" {
				problems = append(problems, "mocknet.base_url must use http")
			}
			if !localhostHostRe.MatchString(u.Host) {
				problems = append(problems, "mocknet.base_url must point to localhost/127.0.0.1 (local mock app only)")
			}
		}
	}

	if strings.TrimSpace(cfg.Storage.SQLitePath) == "" {
		problems = append(problems, "storage.sqlite_path is required")
	}

	if cfg.Browser.SlowMo < 0 {
		problems = append(problems, "browser.slow_mo must be >= 0")
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Logging.Level)) {
	case "debug", "info", "warn", "warning", "error":
		// ok
	case "":
		problems = append(problems, "logging.level is required")
	default:
		problems = append(problems, "logging.level must be one of: debug, info, warn, error")
	}

	if cfg.Run.Timeout <= 0 {
		problems = append(problems, "run.timeout must be > 0")
	}

	if len(problems) > 0 {
		return errors.New("config validation failed: " + strings.Join(problems, "; "))
	}
	return nil
}
