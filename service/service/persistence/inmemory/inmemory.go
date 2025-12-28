package inmemory

import "github.com/dkrizic/feature/service/service/persistence"

type Persistence struct {
	data *[]persistence.KeyValue
}

func NewPersistence() *Persistence {
	return &Persistence{}
}

func (p *Persistence) GetAll() ([]persistence.KeyValue, error) {
	return nil, nil
}

func (p *Persistence) PreSet(persistence.KeyValue) error {
	return nil
}

func (p *Persistence) Set(persistence.KeyValue) error {
	return nil
}

func (p *Persistence) Get(key string) (persistence.KeyValue, error) {
	return persistence.KeyValue{}, nil
}
