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
	AuthURL        string
	FileDir        string
	S3             Storage
}

type Storage struct {
	Endpoint     string
	Bucket       string
	AccessKey    string
	SecretKey    string
	SessionToken string
	Region       string
	Insecure     bool
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		PublicOrigin:   strings.TrimRight(os.Getenv("PUBLIC_ORIGIN"), "/"),
		InternalSecret: os.Getenv("INTERNAL_SECRET"),
		Address:        ":8080",
		AuthURL:        os.Getenv("AUTH_URL"),
		FileDir:        os.Getenv("FILE_DIR"),
		S3: Storage{
			Endpoint:     os.Getenv("S3_ENDPOINT"),
			Bucket:       os.Getenv("S3_BUCKET"),
			AccessKey:    os.Getenv("S3_ACCESS_KEY"),
			SecretKey:    os.Getenv("S3_SECRET_KEY"),
			SessionToken: os.Getenv("S3_SESSION_TOKEN"),
			Region:       os.Getenv("S3_REGION"),
		},
	}
	if cfg.DatabaseURL == "" || cfg.PublicOrigin == "" || cfg.InternalSecret == "" {
		return Config{}, fmt.Errorf("DATABASE_URL, PUBLIC_ORIGIN and INTERNAL_SECRET are required")
	}
	origin, err := url.Parse(cfg.PublicOrigin)
	if err != nil || origin.Host == "" || (origin.Scheme != "http" && origin.Scheme != "https") || origin.Path != "" || origin.RawQuery != "" || origin.ForceQuery || origin.Fragment != "" || origin.User != nil {
		return Config{}, fmt.Errorf("PUBLIC_ORIGIN must contain only an HTTP scheme and host")
	}
	cfg.S3.Insecure, err = boolean("S3_INSECURE")
	if err != nil {
		return Config{}, err
	}
	if cfg.AuthURL == "" {
		return Config{}, fmt.Errorf("AUTH_URL is required")
	}
	endpoint, err := url.Parse(cfg.AuthURL)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.RawQuery != "" || endpoint.ForceQuery || endpoint.Fragment != "" || endpoint.User != nil {
		return Config{}, fmt.Errorf("AUTH_URL must be an absolute HTTP URL without credentials, query or fragment")
	}
	cfg.AuthURL = strings.TrimRight(cfg.AuthURL, "/")
	if cfg.S3.Endpoint == "" {
		if cfg.FileDir == "" {
			return Config{}, fmt.Errorf("FILE_DIR is required for local storage")
		}
	} else if cfg.S3.Bucket == "" || cfg.S3.AccessKey == "" || cfg.S3.SecretKey == "" {
		return Config{}, fmt.Errorf("S3_BUCKET, S3_ACCESS_KEY and S3_SECRET_KEY are required for S3 storage")
	}
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
