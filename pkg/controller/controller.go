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
	Remove() error
	List() ([]*RepoMetaData, error)
	Init() error
	Type() string
}

type RepoMetaData struct {
	Name      string
	ChainNo   int64
	Path      string
	PublicKey string
}

func GetController() (Controller, error) {
	return nil, nil
}
