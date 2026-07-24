package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Environment string `yaml:"environment"`
		AutoMigrate bool   `yaml:"auto_migrate"`
	} `yaml:"app"`
	Server struct {
		Host         string        `yaml:"host"`
		Port         int           `yaml:"port"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
	} `yaml:"server"`
	Database struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		Name         string `yaml:"name"`
		SSLMode      string `yaml:"sslmode"`
		MaxOpenConns int    `yaml:"max_open_conns"`
	} `yaml:"database"`
	Redis struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	JWT struct {
		Secret     string        `yaml:"secret"`
		AccessTTL  time.Duration `yaml:"access_ttl"`
		RefreshTTL time.Duration `yaml:"refresh_ttl"`
		Issuer     string        `yaml:"issuer"`
	} `yaml:"jwt"`
	Admin struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"admin"`
	Agent struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"agent"`
	Storage struct {
		Directory string `yaml:"directory"`
		MaxBytes  int64  `yaml:"max_bytes"`
	} `yaml:"storage"`
	InternalToken string `yaml:"internal_token"`
}

func Load() (*Config, error) {
	path := "config/app.dev.yaml"
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		path = v
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.JWT.Secret = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		c.Redis.Address = v
	}
	if v := os.Getenv("ADMIN_EMAIL"); v != "" {
		c.Admin.Email = v
	}
	if v := os.Getenv("ADMIN_PASSWORD"); v != "" {
		c.Admin.Password = v
	}
	if v := os.Getenv("INTERNAL_TOKEN"); v != "" {
		c.InternalToken = v
	}
	return &c, nil
}
