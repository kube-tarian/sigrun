package controller

func NewKyvernoController() *kyvernoController {
	return &kyvernoController{}
}

func NewSigrunController() *sigrunController {
	return &sigrunController{}
}

type Controller interface {
	Add(repoPaths ...string) error
	Update() error
	Remove(repoPaths ...string) error
	List() (map[string]*RepoInfo, error)
	Init() error
	Type() string
}

type RepoInfo struct {
	Name        string
	Mode        string
	ChainNo     int64
	Path        string
	PublicKey   string
	Maintainers []string
}

func GetController() (Controller, error) {
	return nil, nil
}
