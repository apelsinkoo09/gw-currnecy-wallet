package changer

import (
	"context"
	"fmt"

	proto_exchange "github.com/apelsinkoo09/proto-exchange/exchange"
	_ "google.golang.org/grpc/credentials/insecure"
)

// Exchanger Client struct
type ExchangerClient struct {
	client proto_exchange.ExchangeServiceClient
	cache  *GetExchangeRateCache
}

// Create gRPC client with cache
func NewExchangerClient(client proto_exchange.ExchangeServiceClient, cache *GetExchangeRateCache) *ExchangerClient {
	return &ExchangerClient{
		client: client,
		cache:  cache,
	}
}

// Get exchange rate through gRPC from exchange service
func (e *ExchangerClient) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (float64, error) {

	if rate, found := e.cache.Get(fromCurrency, toCurrency); found {
		return rate, nil
	}

	req := &proto_exchange.CurrencyRequest{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
	}

	result, err := e.client.GetExchangeRateForCurrency(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to ger currency: %v", err)
	}

	rate := float64(result.Rate)
	e.cache.Set(fromCurrency, toCurrency, rate)

	return float64(result.Rate), nil
}
