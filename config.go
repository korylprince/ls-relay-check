package main

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	CustomerID string `required:"true"`
	ConfigURL  string `required:"true"`
	ConfigKey  string `required:"true"`
	CheckURL   string `required:"true"`
	CheckKey   string `required:"true"`

	AllowedEmails []string `default:","`
	allowedEmails map[string]struct{}

	CacheExpiration time.Duration `default:"5m"`
	CachePurge      time.Duration `default:"10m"`
	Cache           *Cache

	ListenAddr string `default:":8080"`
	Prefix     string
}

func NewConfig() *Config {
	config := new(Config)
	envconfig.MustProcess("", config)
	config.allowedEmails = make(map[string]struct{})
	for _, e := range config.AllowedEmails {
		config.allowedEmails[e] = struct{}{}
	}
	config.Cache = NewCache(config.CacheExpiration, config.CachePurge)
	return config
}
