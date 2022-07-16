package cache

type ICache interface {
	Store(key string, value string) error
	Load(key string) (string, error)
}

type RedisCache struct {
}
