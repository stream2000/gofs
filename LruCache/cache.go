package cache

//return status of cache
type cacheStatus struct {
	Gets        int64
	Hits        int64
	MaxItemSize int
	CurrentSize int
}

//this is an interface which defines some common functions
type cache interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	Delete(key string)
	Status() *cacheStatus
}
