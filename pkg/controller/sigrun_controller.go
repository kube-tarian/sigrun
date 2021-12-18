package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	} else {
		if configMap.Name == SIGRUN_CONTROLLER_CONFIG {
			return fmt.Errorf("cluster has already been initialized")
		}
	}

	caCert, caPriv, err := GenerateCACert(24 * 365 * time.Hour)
	if err != nil {
		return err
	}

	whCert, whPriv, err := GenerateCertPem(caCert, caPriv, 24*365*time.Hour)
	if err != nil {
		return err
	}

	resp, err := http.Get("https://raw.githubusercontent.com/devopstoday11/sigrun/main/install.yaml")
	if err != nil {
		return err
	}

	templateBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//
	//templateBytes, err := ioutil.ReadFile("./install.yaml")
	//if err != nil {
	//	return err
	//}

	template := strings.Replace(string(templateBytes), "{{caCert}}", base64.StdEncoding.EncodeToString(CertificateToPem(caCert)), -1)
	template = strings.Replace(template, "{{whCert}}", base64.StdEncoding.EncodeToString(CertificateToPem(whCert)), -1)
	template = strings.Replace(template, "{{whKey}}", base64.StdEncoding.EncodeToString(PrivateKeyToPem(whPriv)), -1)

	f, err := ioutil.TempFile(os.TempDir(), "sigrun")
	if err != nil {
		return err
	}

	_, err = io.Copy(f, strings.NewReader(template))
	if err != nil {
		return err
	}

	fPath, err := filepath.Abs(f.Name())
	if err != nil {
		return err
	}

	err = exec.Command("kubectl", "apply", "-f", fPath).Run()
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

		configMap, err = s.AddRepo(configMap, guid, path, conf)
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

	sigrunReposJSON := configMap.Data["guid_to_repo_info"]
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

	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(configMap.Data["guid_to_repo_info"]))).Decode(&guidToRepoMeta)

	for guid, md := range guidToRepoMeta {
		confMap, err := config.ReadRepos(md.Path)
		if err != nil {
			return err
		}
		newConf := confMap[md.Path]
		newConfInfo := newConf.GetVerificationInfo()

		if newConfInfo.ChainNo > md.ChainNo {
			oldConf := config.GetVerificationConfigFromVerificationInfo(&md.VerificationInfo)
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

			configMap, err = s.AddRepo(configMap, guid, md.Path, newConf)
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

func (s *sigrunController) AddRepo(configMap *kubernetesCoreV1.ConfigMap, guid, path string, c config.Config) (*kubernetesCoreV1.ConfigMap, error) {
	conf := c.GetVerificationInfo()

	// add repos to sigrun-repos annotation
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(configMap.Data["guid_to_repo_info"]))).Decode(&guidToRepoMeta)

	if guidToRepoMeta[guid] != nil {
		return nil, fmt.Errorf("sigrun repo with guid " + guid + " and name " + conf.Name + " has already been added")
	}

	guidToRepoMeta[guid] = &RepoInfo{
		VerificationInfo: *conf,
		Path:             path,
	}

	imageToGuids := make(map[string][]string)
	_ = json.NewDecoder(strings.NewReader(configMap.Data["image_to_guids"])).Decode(&imageToGuids)
	for _, img := range conf.Images {
		imageToGuids[img] = append(imageToGuids[img], guid)
	}

	guidToRepoRaw, err := json.Marshal(guidToRepoMeta)
	if err != nil {
		return nil, err
	}
	configMap.Data["guid_to_repo_info"] = string(guidToRepoRaw)

	imageToGuidsRaw, err := json.Marshal(imageToGuids)
	if err != nil {
		return nil, err
	}
	configMap.Data["image_to_guids"] = string(imageToGuidsRaw)

	return configMap, nil
}

func (s *sigrunController) removeRepo(configMap *kubernetesCoreV1.ConfigMap, guid string) (*kubernetesCoreV1.ConfigMap, error) {
	// add repos to sigrun-repos annotation
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(configMap.Data["guid_to_repo_info"])).Decode(&guidToRepoMeta)
	delete(guidToRepoMeta, guid)

	imageToGuids := make(map[string][]string)
	_ = json.NewDecoder(strings.NewReader(configMap.Data["image_to_guids"])).Decode(&imageToGuids)
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
	configMap.Data["guid_to_repo_info"] = string(guidToRepoRaw)

	imageToGuidsRaw, err := json.Marshal(newImageToGuids)
	if err != nil {
		return nil, err
	}
	configMap.Data["image_to_guids"] = string(imageToGuidsRaw)

	return configMap, nil
}
