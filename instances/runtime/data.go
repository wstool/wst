package runtime

import "sync"

// Data is the interface type to allow storage and retrieval of data across different Actions.
type Data interface {
	Store(key string, value interface{})
	Load(key string) (interface{}, bool)
}

// runtimeDataImpl is an implementation of the RuntimeData interface.
type dataImpl struct {
	data sync.Map
}

// Store stores the value for a key.
func (rt *dataImpl) Store(key string, value interface{}) {
	rt.data.Store(key, value)
}

// Load loads the value for a key.
func (rt *dataImpl) Load(key string) (interface{}, bool) {
	return rt.data.Load(key)
}
