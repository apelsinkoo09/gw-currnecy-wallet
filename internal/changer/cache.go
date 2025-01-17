package changer

import (
	"time"
	"github.com/patrickmn/go-cache"
)

type GetExchangeRateCache struct{
	cache *cache.Cache
}

// Create cache
func RateCache(defaultExpiration, cleanupInterval time.Duration) *GetExchangeRateCache{
	return &GetExchangeRateCache{
		cache: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get currency from cache
func (c *GetExchangeRateCache) Get(fromCurrency, toCurrency string) (float64, bool){
	key := fromCurrency + "->" + toCurrency
	value, found := c.cache.Get(key)
	if found {
		return value.(float64), true
	}
	return 0, false
}

// Save currency to cahce
func (c *GetExchangeRateCache) Set(fromCurrency, toCurrency string, rate float64) {
	key := fromCurrency + "->" + toCurrency
	c.cache.Set(key, rate, cache.DefaultExpiration)
}