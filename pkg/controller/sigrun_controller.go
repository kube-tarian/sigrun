package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	kubernetesCoreV1 "k8s.io/api/core/v1"

	"github.com/devopstoday11/sigrun/pkg/config"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

const (
	SIGRUN_CONTROLLER_CONFIG    = "sigrun-controller-config"
	SIGRUN_CONTROLLER_NAMESPACE = "default"
)

type sigrunController struct {
}

func (s *sigrunController) Type() string {
	return CONTROLLER_TYPE_SIGRUN
}

func (s *sigrunController) Init() error {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return err
	}

	kclient, err := kubernetes.NewForConfig(kRestConf)
	if err != nil {
		return err
	}

	configMap, err := kclient.CoreV1().ConfigMaps(SIGRUN_CONTROLLER_NAMESPACE).Get(context.Background(), SIGRUN_CONTROLLER_CONFIG, v1.GetOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not find") {
			return err
		}
	} else {
		if configMap.Name == SIGRUN_CONTROLLER_CONFIG {
			return fmt.Errorf("Cluster has already been initialized")
		}
	}

	err = exec.Command("kubectl", "create", "-f", "https://raw.githubusercontent.com/devopstoday11/sigrun/main/sigrun-controller/install.yaml").Run()
	if err != nil {
		return err
	}

	return nil
}

func (s *sigrunController) Add(repoPaths ...string) error {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return err
	}

	kclient, err := kubernetes.NewForConfig(kRestConf)
	if err != nil {
		return err
	}

	configMap, err := kclient.CoreV1().ConfigMaps(SIGRUN_CONTROLLER_NAMESPACE).Get(context.Background(), SIGRUN_CONTROLLER_CONFIG, v1.GetOptions{})
	if err != nil {
		return err
	}

	pathToConfig, err := config.ReadRepos(repoPaths...)
	if err != nil {
		return err
	}

	for path, conf := range pathToConfig {
		guid, err := config.GetGUID(path)
		if err != nil {
			return err
		}

		configMap, err = s.addRepo(configMap, guid, path, conf)
		if err != nil {
			return err
		}
	}

	_, err = kclient.CoreV1().ConfigMaps(SIGRUN_CONTROLLER_NAMESPACE).Update(context.Background(), configMap, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *sigrunController) List() (map[string]*RepoInfo, error) {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return nil, err
	}

	kclient, err := kubernetes.NewForConfig(kRestConf)
	if err != nil {
		return nil, err
	}

	configMap, err := kclient.CoreV1().ConfigMaps(SIGRUN_CONTROLLER_NAMESPACE).Get(context.Background(), SIGRUN_CONTROLLER_CONFIG, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	sigrunReposJSON, err := base64.StdEncoding.DecodeString(configMap.Data["guid_to_repo_info"])
	if err != nil {
		return nil, err
	}
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)

	return guidToRepoMeta, nil
}

func (s *sigrunController) Remove(repoPaths ...string) error {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return err
	}

	kclient, err := kubernetes.NewForConfig(kRestConf)
	if err != nil {
		return err
	}

	configMap, err := kclient.CoreV1().ConfigMaps(SIGRUN_CONTROLLER_NAMESPACE).Get(context.Background(), SIGRUN_CONTROLLER_CONFIG, v1.GetOptions{})
	if err != nil {
		return err
	}

	for _, path := range repoPaths {
		guid, err := config.GetGUID(path)
		if err != nil {
			return err
		}

		configMap, err = s.removeRepo(configMap, guid)
		if err != nil {
			return err
		}
	}

	_, err = kclient.CoreV1().ConfigMaps(SIGRUN_CONTROLLER_NAMESPACE).Update(context.Background(), configMap, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *sigrunController) Update() error {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return err
	}

	kclient, err := kubernetes.NewForConfig(kRestConf)
	if err != nil {
		return err
	}

	configMap, err := kclient.CoreV1().ConfigMaps(SIGRUN_CONTROLLER_NAMESPACE).Get(context.Background(), SIGRUN_CONTROLLER_CONFIG, v1.GetOptions{})
	if err != nil {
		return err
	}

	sigrunReposJSON, err := base64.StdEncoding.DecodeString(configMap.Data["guid_to_repo_info"])
	if err != nil {
		return err
	}
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)

	for guid, md := range guidToRepoMeta {
		confMap, err := config.ReadRepos(md.Path)
		if err != nil {
			return err
		}
		newConf := confMap[md.Path]
		newConfInfo := newConf.GetVerificationInfo()

		if newConfInfo.ChainNo > md.ChainNo {
			oldConf := config.GetVerificationConfigFromVerificationInfo(&config.VerificationInfo{
				Name:        md.Name,
				Mode:        md.Mode,
				ChainNo:     md.ChainNo,
				PublicKey:   md.PublicKey,
				Maintainers: md.Maintainers,
				Images:      nil,
			})
			fmt.Println("verifying sigrun repo with guid " + guid + " and name " + md.Name + " from chain no " + fmt.Sprint(md.ChainNo) + " to " + fmt.Sprint(newConfInfo.ChainNo))
			err = config.VerifyChain(md.Path, oldConf, newConf)
			if err != nil {
				return err
			}

			fmt.Println("updating sigrun repo with guid " + guid + " and name " + md.Name + " from chain no " + fmt.Sprint(md.ChainNo) + " to " + fmt.Sprint(newConfInfo.ChainNo))
			configMap, err = s.removeRepo(configMap, guid)
			if err != nil {
				return err
			}

			configMap, err = s.addRepo(configMap, guid, md.Path, newConf)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("sigrun repo with guid " + guid + " and name " + md.Name + " is already upto date")
		}
	}

	_, err = kclient.CoreV1().ConfigMaps(SIGRUN_CONTROLLER_NAMESPACE).Update(context.Background(), configMap, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *sigrunController) addRepo(configMap *kubernetesCoreV1.ConfigMap, guid, path string, c config.Config) (*kubernetesCoreV1.ConfigMap, error) {
	conf := c.GetVerificationInfo()

	// add repos to sigrun-repos annotation
	sigrunReposJSON, err := base64.StdEncoding.DecodeString(configMap.Data["guid_to_repo_info"])
	if err != nil {
		return nil, err
	}
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)

	if guidToRepoMeta[guid] != nil {
		return nil, fmt.Errorf("sigrun repo with guid " + guid + " and name " + conf.Name + " has already been added")
	}

	guidToRepoMeta[guid] = &RepoInfo{
		Name:        conf.Name,
		Mode:        conf.Mode,
		ChainNo:     conf.ChainNo,
		Path:        path,
		PublicKey:   conf.PublicKey,
		Maintainers: conf.Maintainers,
	}

	imageToGuidsRaw, err := base64.StdEncoding.DecodeString(configMap.Data["image_to_guids"])
	if err != nil {
		return nil, err
	}
	imageToGuids := make(map[string][]string)
	_ = json.NewDecoder(strings.NewReader(string(imageToGuidsRaw))).Decode(&imageToGuids)
	imageToGuids[guid] = append(imageToGuids[guid], guid)

	guidToRepoRaw, err := json.Marshal(guidToRepoMeta)
	if err != nil {
		return nil, err
	}
	configMap.Data["guid_to_repo_info"] = base64.StdEncoding.EncodeToString(guidToRepoRaw)

	imageToGuidsRaw, err = json.Marshal(imageToGuids)
	if err != nil {
		return nil, err
	}
	configMap.Data["image_to_guids"] = base64.StdEncoding.EncodeToString(imageToGuidsRaw)

	return configMap, nil
}

func (s *sigrunController) removeRepo(configMap *kubernetesCoreV1.ConfigMap, guid string) (*kubernetesCoreV1.ConfigMap, error) {
	// add repos to sigrun-repos annotation
	sigrunReposJSON, err := base64.StdEncoding.DecodeString(configMap.Data["guid_to_repo_info"])
	if err != nil {
		return nil, err
	}
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)
	delete(guidToRepoMeta, guid)

	imageToGuidsRaw, err := base64.StdEncoding.DecodeString(configMap.Data["image_to_guids"])
	if err != nil {
		return nil, err
	}
	imageToGuids := make(map[string][]string)
	_ = json.NewDecoder(strings.NewReader(string(imageToGuidsRaw))).Decode(&imageToGuids)
	newImageToGuids := make(map[string][]string)

	for image, guids := range imageToGuids {
		var newGuids []string
		for _, v := range guids {
			if v != guid {
				newGuids = append(newGuids, v)
			}
		}

		if len(newGuids) <= 0 {
			continue
		}

		newImageToGuids[image] = newGuids
	}

	guidToRepoRaw, err := json.Marshal(guidToRepoMeta)
	if err != nil {
		return nil, err
	}
	configMap.Data["guid_to_repo_info"] = base64.StdEncoding.EncodeToString(guidToRepoRaw)

	imageToGuidsRaw, err = json.Marshal(newImageToGuids)
	if err != nil {
		return nil, err
	}
	configMap.Data["image_to_guids"] = base64.StdEncoding.EncodeToString(imageToGuidsRaw)

	return configMap, nil
}
