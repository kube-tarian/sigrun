package controller

const (
	CONTROLLER_TYPE_KYVERNO = "kyverno"
	CONTROLLER_TYPE_SIGRUN  = "sigrun"
)

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

func GetController(contType string) (Controller, error) {
	switch contType {
	case CONTROLLER_TYPE_KYVERNO:
		return NewKyvernoController(), nil
	default:
		return NewSigrunController(), nil
	}
}

//
//func detectControllerType() (string, error) {
//	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
//	if err != nil {
//		return "", err
//	}
//	kClient, err := kyvernoclient.NewForConfig(kRestConf)
//	if err != nil {
//		return "", err
//	}
//
//	ctx := context.Background()
//	cpol, err := kClient.KyvernoV1().ClusterPolicies().Get(ctx, KYVERNO_POLICY_NAME, v1.GetOptions{})
//	if err != nil {
//		if !strings.Contains(err.Error(), "not find") {
//			return "", err
//		}
//	} else {
//		if cpol.Name == KYVERNO_POLICY_NAME {
//			return CONTROLLER_TYPE_KYVERNO, nil
//		}
//	}
//
//	return "", nil
//}
