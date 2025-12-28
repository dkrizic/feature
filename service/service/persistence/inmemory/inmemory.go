package inmemory

type Persistence struct {
	data string
}

func NewPersistence() *Persistence {
	return &Persistence{}
}

func (p *Persistence) Save(data string) error {
	p.data = data
	return nil
}

func (p *Persistence) Load() (string, error) {
	return p.data, nil
}
