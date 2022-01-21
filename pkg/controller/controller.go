package controller

import (
	"encoding/json"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/config"

	corev1 "k8s.io/api/core/v1"
)

const (
	CONTROLLER_TYPE_SIGRUN = "sigrun"
	GUID_TO_REPO_INFO      = "guid_to_repo_info"
	IMAGE_TO_GUIDS         = "image_to_guids"
)

func New() *sigrunController {
	return &sigrunController{}
}

type RepoInfo struct {
	config.VerificationInfo
	Path string
}

func ParseSigrunConfigMap(configMap *corev1.ConfigMap) (map[string]*RepoInfo, map[string][]string, error) {
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(configMap.Data[GUID_TO_REPO_INFO])).Decode(&guidToRepoMeta)

	imageToGuids := make(map[string][]string)
	_ = json.NewDecoder(strings.NewReader(configMap.Data[IMAGE_TO_GUIDS])).Decode(&imageToGuids)

	return guidToRepoMeta, imageToGuids, nil
}
