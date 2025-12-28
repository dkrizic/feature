package configmap

import "github.com/dkrizic/feature/service/service/persistence"

type Persistence struct {
	configMapName string
}

func NewPersistence(configMapName string) *Persistence {
	return &Persistence{
		configMapName: configMapName,
	}
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
