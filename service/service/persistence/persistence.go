package persistence

import "context"

type KeyValue struct {
	Key   string
	Value string
}

type Persistence interface {
	GetAll(context.Context) ([]KeyValue, error)
	PreSet(context.Context, KeyValue) error
	Set(context.Context, KeyValue) error
	Get(context.Context, string) (KeyValue, error)
}
