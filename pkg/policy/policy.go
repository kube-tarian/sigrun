package policy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/config"

	kyvernoV1 "github.com/kyverno/kyverno/pkg/api/kyverno/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NAME = "sigrun-verify"

func New() *kyvernoV1.ClusterPolicy {
	background := false
	return &kyvernoV1.ClusterPolicy{
		TypeMeta: v1.TypeMeta{
			Kind:       "ClusterPolicy",
			APIVersion: "kyverno.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: NAME,
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

type RepoMetaData struct {
	Moniker   string
	ChainNo   int64
	Path      string
	PublicKey string
}

func RemoveRepo(cpol *kyvernoV1.ClusterPolicy, guid string) (*kyvernoV1.ClusterPolicy, error) {
	sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
	if err != nil {
		return nil, err
	}
	guidToRepoMeta := make(map[string]*RepoMetaData)
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

func AddRepo(cpol *kyvernoV1.ClusterPolicy, guid, path string, conf *config.Config) (*kyvernoV1.ClusterPolicy, error) {

	// add repos to sigrun-repos annotation
	sigrunReposJSON, err := base64.StdEncoding.DecodeString(cpol.Annotations["sigrun-repos-metadata"])
	if err != nil {
		return nil, err
	}
	guidToRepoMeta := make(map[string]*RepoMetaData)
	_ = json.NewDecoder(strings.NewReader(string(sigrunReposJSON))).Decode(&guidToRepoMeta)

	pubKToGUID := make(map[string]string)
	for guid, repoMD := range guidToRepoMeta {
		pubKToGUID[repoMD.PublicKey] = guid
	}

	if guidToRepoMeta[guid] != nil {
		return nil, fmt.Errorf("sigrun repo with guid " + guid + " and moniker " + conf.Name + " has already been added")
	}

	if g := pubKToGUID[conf.PublicKey]; g != "" {
		return nil, fmt.Errorf("sigrun repo with guid " + guid + " and moniker " + conf.Name + " has the same public key as a sigrun repo that has already been added with guid " + g)
	}

	guidToRepoMeta[guid] = &RepoMetaData{
		Moniker:   conf.Name,
		ChainNo:   conf.ChainNo,
		Path:      path,
		PublicKey: conf.PublicKey,
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
