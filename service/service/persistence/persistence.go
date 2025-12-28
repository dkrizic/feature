package persistence

type KeyValue struct {
	Name  string
	Value string
}

type Persistence interface {
	GetAll() ([]KeyValue, error)
	PreSet(KeyValue) error
	Set(KeyValue) error
	Get(key string) (KeyValue, error)
}
