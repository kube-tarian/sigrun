package config

import (
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/util"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/sigstore/pkg/signature"
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
}

func ReadRepositoryConfig() (Config, error) {
	encodedConfig, err := ioutil.ReadFile(FILE_NAME)
	if err != nil {
		return nil, err
	}

	return parseConfig(encodedConfig)
}

func VerifySignature(pubKRaw string, conf *DefaultConfig) error {
	sig := conf.Signature
	conf.Signature = ""
	data, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	conf.Signature = sig
	pubK, err := cosign.PemToECDSAKey([]byte(pubKRaw))
	if err != nil {
		return err
	}
	verifier := signature.ECDSAVerifier{
		Key:     pubK,
		HashAlg: crypto.SHA256,
	}

	sigRaw, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		return err
	}

	return verifier.Verify(context.Background(), data, sigRaw)
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

func VerifyChain(oldPubK, oldPath string, oldChainNo int64, newConf *DefaultConfig) error {
	currentChainNo := oldChainNo + 1
	prevPubK := oldPubK

	var currConf *DefaultConfig
	var err error
	for currentChainNo <= newConf.ChainNo {
		currPath := strings.Replace(oldPath, FILE_NAME, ".sigrun/"+fmt.Sprint(currentChainNo)+".json", -1)
		confMap, err := ReadRepos(currPath)
		if err != nil {
			return err
		}
		currConf = confMap[currPath]
		err = VerifySignature(prevPubK, currConf)
		if err != nil {
			return err
		}
		prevPubK = currConf.PublicKey
		currentChainNo = currConf.ChainNo + 1
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
