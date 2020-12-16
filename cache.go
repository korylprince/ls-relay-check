package main

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache struct {
	FilterConfigs *cache.Cache
	Hosts         *cache.Cache
	Videos        *cache.Cache
	Channels      *cache.Cache
}

func NewCache(expiration, purge time.Duration) *Cache {
	return &Cache{
		FilterConfigs: cache.New(expiration, purge),
		Hosts:         cache.New(expiration, purge),
		Videos:        cache.New(expiration, purge),
		Channels:      cache.New(expiration, purge),
	}
}
