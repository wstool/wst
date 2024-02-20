package runtime

import "sync"

// Data is the interface type to allow storage and retrieval of data across different Actions.
type Data interface {
	Store(key string, value interface{}) error
	Load(key string) (interface{}, bool)
}

func CreateData() Data {
	return &syncData{}
}

// runtimeDataImpl is an implementation of the RuntimeData interface.
type syncData struct {
	data sync.Map
}

// Store stores the value for a key.
func (rt *syncData) Store(key string, value interface{}) error {
	rt.data.Store(key, value)
	return nil
}

// Load loads the value for a key.
func (rt *syncData) Load(key string) (interface{}, bool) {
	return rt.data.Load(key)
}
