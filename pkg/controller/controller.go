package controller

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/config"

	v1 "k8s.io/api/core/v1"
)

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
	config.VerificationInfo
	Path string
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

func ParseSigrunConfigMap(configMap *v1.ConfigMap) (map[string]*RepoInfo, map[string][]string, error) {
	sigrunReposJSON, err := base64.StdEncoding.DecodeString(configMap.Data["guid_to_repo_info"])
	if err != nil {
		return nil, nil, err
	}
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)

	imageToGuidsRaw, err := base64.StdEncoding.DecodeString(configMap.Data["image_to_guids"])
	if err != nil {
		return nil, nil, err
	}
	imageToGuids := make(map[string][]string)
	_ = json.NewDecoder(strings.NewReader(string(imageToGuidsRaw))).Decode(&imageToGuids)

	return guidToRepoMeta, imageToGuids, nil
}
