package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/util"
)

const (
	CONFIG_MODE_DEFAULT = "default"
	CONFIG_MODE_KEYLESS = "keyless"
)

func NewDefaultConfig(name, pubKey, privKey string, images []string) *DefaultConfig {
	return &DefaultConfig{
		Name:       name,
		Mode:       CONFIG_MODE_DEFAULT,
		ChainNo:    0,
		PublicKey:  pubKey,
		PrivateKey: privKey,
		Images:     images,
		Signature:  "",
	}
}

func NewKeylessConfig(name string, maintainers, images []string) *KeylessConfig {
	return &KeylessConfig{
		Name:        name,
		Mode:        CONFIG_MODE_KEYLESS,
		ChainNo:     0,
		Maintainers: maintainers,
		Images:      images,
		Signature:   "",
	}
}

type Config interface {
	InitializeRepository() error
	SignImages() error
	CommitRepositoryUpdate() error
	GetChainNo() int64
	Sign([]byte) (string, error)
	SignDoc() ([]byte, error)
	Validate() error
	GetVerificationInfo() *VerificationInfo
	VerifySuccessorConfig(Config) error
	GetSignature() string
}

type VerificationInfo struct {
	Name        string
	Mode        string
	ChainNo     int64
	PublicKey   string
	Maintainers []string
	Images      []string
}

func ReadRepositoryConfig() (Config, error) {
	encodedConfig, err := ioutil.ReadFile(FILE_NAME)
	if err != nil {
		return nil, err
	}

	return parseConfig(encodedConfig)
}

func GetGUID(path string) (string, error) {
	genesisConfPath := strings.Replace(path, FILE_NAME, ".sigrun/0.json", -1)

	resp, err := http.Get(genesisConfPath)
	if err != nil {
		return "", err
	}

	return util.SHA256Hash(resp.Body)
}

// TODO should be repo urls, currentl config file urls
func ReadRepos(repoUrls ...string) (map[string]Config, error) {
	pathToConfig := make(map[string]Config)
	for _, path := range repoUrls {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}

		confRaw, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		pathToConfig[path], err = parseConfig(confRaw)
		if err != nil {
			return nil, err
		}
	}

	return pathToConfig, nil
}

func GetVerificationConfigFromVerificationInfo(info *VerificationInfo) Config {
	if info.Mode == CONFIG_MODE_KEYLESS {
		return &KeylessConfig{
			Name:        info.Name,
			Mode:        info.Mode,
			ChainNo:     info.ChainNo,
			Maintainers: info.Maintainers,
			Images:      info.Images,
			Signature:   "",
		}
	} else {
		return &DefaultConfig{
			Name:       info.Name,
			Mode:       info.Mode,
			ChainNo:    info.ChainNo,
			PublicKey:  info.PublicKey,
			PrivateKey: "",
			Images:     info.Images,
			Signature:  "",
		}
	}
}

func VerifyChain(repoPath string, oldConf, newConf Config) error {
	var err error

	currentChainNo := oldConf.GetChainNo() + 1
	prevConf := oldConf
	var currConf Config
	for currentChainNo <= newConf.GetChainNo() {
		currPath := strings.Replace(repoPath, FILE_NAME, ".sigrun/"+fmt.Sprint(currentChainNo)+".json", -1)
		confMap, err := ReadRepos(currPath)
		if err != nil {
			return err
		}
		currConf = confMap[currPath]

		err = prevConf.VerifySuccessorConfig(currConf)
		if err != nil {
			return err
		}
		prevConf = currConf
		currentChainNo = currConf.GetChainNo() + 1
	}

	isSame, err := isSame(currConf, newConf)
	if err != nil {
		return err
	}

	if !isSame {
		return fmt.Errorf("chain head is not the same as config file")
	}

	return nil
}
