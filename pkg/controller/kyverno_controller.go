package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	kyvernoV1 "github.com/kyverno/kyverno/pkg/api/kyverno/v1"

	"github.com/devopstoday11/sigrun/pkg/config"

	kyvernoclient "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type kyvernoController struct {
}

func (k *kyvernoController) Type() string {
	return CONTROLLER_TYPE_KYVERNO
}

func (k *kyvernoController) Add(repoPaths ...string) error {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return err
	}

	kClient, err := kyvernoclient.NewForConfig(kRestConf)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cpol, err := kClient.KyvernoV1().ClusterPolicies().Get(ctx, KYVERNO_POLICY_NAME, v1.GetOptions{})
	if err != nil {
		return err
	}

	pathToConfig, err := config.ReadRepos(repoPaths...)
	if err != nil {
		return err
	}

	for path, conf := range pathToConfig {
		if conf.GetVerificationInfo().Mode == config.CONFIG_MODE_KEYLESS {
			return fmt.Errorf("kyverno controller does not support keyless config yet")
		}

		guid, err := config.GetGUID(path)
		if err != nil {
			return err
		}

		cpol, err = k.addRepo(cpol, guid, path, conf)
		if err != nil {
			return err
		}
	}

	_, err = kClient.KyvernoV1().ClusterPolicies().Update(ctx, cpol, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (k *kyvernoController) Update() error {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return err
	}

	kClient, err := kyvernoclient.NewForConfig(kRestConf)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cpol, err := kClient.KyvernoV1().ClusterPolicies().Get(ctx, KYVERNO_POLICY_NAME, v1.GetOptions{})
	if err != nil {
		return err
	}

	// add repos to sigrun-repos annotation
	sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
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
			oldConf := config.GetVerificationConfigFromVerificationInfo(&md.VerificationInfo)
			fmt.Println("verifying sigrun repo with guid " + guid + " and name " + md.Name + " from chain no " + fmt.Sprint(md.ChainNo) + " to " + fmt.Sprint(newConfInfo.ChainNo))
			err = config.VerifyChain(md.Path, oldConf, newConf)
			if err != nil {
				return err
			}

			fmt.Println("updating sigrun repo with guid " + guid + " and name " + md.Name + " from chain no " + fmt.Sprint(md.ChainNo) + " to " + fmt.Sprint(newConfInfo.ChainNo))
			cpol, err = k.removeRepo(cpol, guid)
			if err != nil {
				return err
			}

			cpol, err = k.addRepo(cpol, guid, md.Path, newConf)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("sigrun repo with guid " + guid + " and name " + md.Name + " is already upto date")
		}
	}

	_, err = kClient.KyvernoV1().ClusterPolicies().Update(ctx, cpol, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (k *kyvernoController) Remove(repoPaths ...string) error {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return err
	}

	kClient, err := kyvernoclient.NewForConfig(kRestConf)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cpol, err := kClient.KyvernoV1().ClusterPolicies().Get(ctx, KYVERNO_POLICY_NAME, v1.GetOptions{})
	if err != nil {
		return err
	}

	for _, path := range repoPaths {
		guid, err := config.GetGUID(path)
		if err != nil {
			return err
		}

		cpol, err = k.removeRepo(cpol, guid)
		if err != nil {
			return err
		}
	}

	_, err = kClient.KyvernoV1().ClusterPolicies().Update(ctx, cpol, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (k *kyvernoController) List() (map[string]*RepoInfo, error) {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return nil, err
	}

	kClient, err := kyvernoclient.NewForConfig(kRestConf)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	cpol, err := kClient.KyvernoV1().ClusterPolicies().Get(ctx, KYVERNO_POLICY_NAME, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
	if err != nil {
		return nil, err
	}
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)

	return guidToRepoMeta, nil
}

func (k *kyvernoController) Init() error {
	kRestConf, err := genericclioptions.NewConfigFlags(true).ToRESTConfig()
	if err != nil {
		return err
	}

	kClient, err := kyvernoclient.NewForConfig(kRestConf)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cpol, err := kClient.KyvernoV1().ClusterPolicies().Get(ctx, KYVERNO_POLICY_NAME, v1.GetOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "not find") {
			return err
		}
	} else {
		if cpol.Name == KYVERNO_POLICY_NAME {
			return fmt.Errorf("Cluster has already been initialized")
		}
	}

	err = exec.Command("kubectl", "create", "-f", "https://raw.githubusercontent.com/kyverno/kyverno/13caaed8b778a977ceed7c041a83a5642ff98cf5/definitions/install.yaml").Run()
	if err != nil {
		return err
	}

	_, err = kClient.KyvernoV1().ClusterPolicies().Create(ctx, newKyvernoPolicy(), v1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

const KYVERNO_POLICY_NAME = "sigrun-verify"

func newKyvernoPolicy() *kyvernoV1.ClusterPolicy {
	background := false
	return &kyvernoV1.ClusterPolicy{
		TypeMeta: v1.TypeMeta{
			Kind:       "ClusterPolicy",
			APIVersion: "kyverno.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: KYVERNO_POLICY_NAME,
			Annotations: map[string]string{
				"sigrun-keys":  "",
				"sigrun-repos": "",
			},
		},
		Spec: kyvernoV1.Spec{
			Rules: []kyvernoV1.Rule{
				{
					Name: "sigrun",
					MatchResources: kyvernoV1.MatchResources{
						ResourceDescription: kyvernoV1.ResourceDescription{
							Kinds: []string{"Pod"},
						},
					},
				},
			},
			ValidationFailureAction: "enforce",
			Background:              &background,
		},
		Status: kyvernoV1.PolicyStatus{},
	}
}

func (k *kyvernoController) removeRepo(cpol *kyvernoV1.ClusterPolicy, guid string) (*kyvernoV1.ClusterPolicy, error) {
	sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
	if err != nil {
		return nil, err
	}
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)
	verifyImages := cpol.Spec.Rules[0].VerifyImages

	if guidToRepoMeta[guid] == nil {
		return nil, fmt.Errorf("sigrun repo with guid " + guid + " does not exist ")
	}
	var buf []*kyvernoV1.ImageVerification
	for _, vi := range verifyImages {
		if vi.Key != guidToRepoMeta[guid].PublicKey {
			buf = append(buf, vi)
		}
	}
	verifyImages = buf
	delete(guidToRepoMeta, guid)

	guidToRepoRaw, err := json.Marshal(guidToRepoMeta)
	if err != nil {
		return nil, err
	}
	cpol.Annotations["sigrun-repos-metadata"] = base64.StdEncoding.EncodeToString(guidToRepoRaw)
	cpol.Spec.Rules[0].VerifyImages = verifyImages

	return cpol, nil
}

func (k *kyvernoController) addRepo(cpol *kyvernoV1.ClusterPolicy, guid, path string, c config.Config) (*kyvernoV1.ClusterPolicy, error) {
	conf := c.GetVerificationInfo()

	// add repos to sigrun-repos annotation
	sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
	if err != nil {
		return nil, err
	}
	guidToRepoMeta := make(map[string]*RepoInfo)
	_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)

	pubKToGUID := make(map[string]string)
	for guid, repoMD := range guidToRepoMeta {
		pubKToGUID[repoMD.PublicKey] = guid
	}

	if guidToRepoMeta[guid] != nil {
		return nil, fmt.Errorf("sigrun repo with guid " + guid + " and name " + conf.Name + " has already been added")
	}

	if g := pubKToGUID[conf.PublicKey]; g != "" {
		return nil, fmt.Errorf("sigrun repo with guid " + guid + " and name " + conf.Name + " has the same public key as a sigrun repo that has already been added with guid " + g)
	}

	guidToRepoMeta[guid] = &RepoInfo{
		VerificationInfo: *conf,
		Path:             path,
	}
	for _, confImg := range conf.Images {
		cpol.Spec.Rules[0].VerifyImages = append(cpol.Spec.Rules[0].VerifyImages, &kyvernoV1.ImageVerification{
			Image: confImg + "*",
			Key:   conf.PublicKey,
		})
	}
	guidToRepoRaw, err := json.Marshal(guidToRepoMeta)
	if err != nil {
		return nil, err
	}
	cpol.Annotations["sigrun-repos-metadata"] = base64.StdEncoding.EncodeToString(guidToRepoRaw)

	return cpol, nil
}
