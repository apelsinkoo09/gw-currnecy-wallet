package storages

type Storage interface {
	GetAllExchangeRates() (map[string]float32, error)
	GetExchangeRates(from_currency, to_currnecy string) (float32, error)
}
