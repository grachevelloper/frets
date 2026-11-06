package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DatabaseURL    string
	PublicOrigin   string
	InternalSecret string
	Address        string
	Google         OAuth
	Yandex         OAuth
	LocalLogin     bool
}

type OAuth struct {
	ClientID     string
	ClientSecret string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		PublicOrigin:   strings.TrimRight(os.Getenv("PUBLIC_ORIGIN"), "/"),
		InternalSecret: os.Getenv("INTERNAL_SECRET"),
		Address:        ":8080",
		Google: OAuth{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		},
		Yandex: OAuth{
			ClientID:     os.Getenv("YANDEX_CLIENT_ID"),
			ClientSecret: os.Getenv("YANDEX_CLIENT_SECRET"),
		},
	}

	if cfg.DatabaseURL == "" || cfg.PublicOrigin == "" || cfg.InternalSecret == "" {
		return Config{}, fmt.Errorf("DATABASE_URL, PUBLIC_ORIGIN and INTERNAL_SECRET are required")
	}
	origin, err := url.Parse(cfg.PublicOrigin)
	if err != nil || origin.Host == "" || (origin.Scheme != "http" && origin.Scheme != "https") || origin.Path != "" || origin.RawQuery != "" || origin.ForceQuery || origin.Fragment != "" || origin.User != nil {
		return Config{}, fmt.Errorf("PUBLIC_ORIGIN must contain only an HTTP scheme and host")
	}
	cfg.LocalLogin, err = boolean("LOCAL_LOGIN")
	if err != nil {
		return Config{}, err
	}
	cfg.LocalLogin = cfg.LocalLogin && cfg.PublicOrigin == "http://localhost:6666"
	return cfg, nil
}

func boolean(name string) (bool, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return false, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", name)
	}
	return value, nil
}
