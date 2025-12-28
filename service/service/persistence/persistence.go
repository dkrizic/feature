package persitence

type Persistence interface {
	Save(data string) error
	Load() (string, error)
}
