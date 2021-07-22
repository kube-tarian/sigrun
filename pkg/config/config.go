package config

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/sigstore/sigstore/pkg/signature"

	"github.com/devopstoday11/sigrun/pkg/util"

	"github.com/sigstore/cosign/pkg/cosign"
)

type Config struct {
	Moniker    string
	ChainNo    int64
	PublicKey  string
	PrivateKey string
	Images     []string
	Signature  string
}

func (c *Config) Validate() error {
	return nil
}

func (c *Config) Sign(password string, conf *Config) (string, error) {
	conf.Signature = ""
	data, err := json.Marshal(conf)
	if err != nil {
		return "", err
	}

	signer, err := cosign.LoadECDSAPrivateKey([]byte(c.PrivateKey), []byte(password))
	if err != nil {
		return "", err
	}

	sig, _, err := signer.Sign(context.Background(), data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

func VerifySignature(pubKRaw string, conf *Config) error {
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

const FILE_NAME = "sigrun-config.json"

func Get(path string) (*Config, error) {
	configF, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configF.Close()

	var conf Config
	err = json.NewDecoder(configF).Decode(&conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func IsSame(conf1, conf2 *Config) (bool, error) {
	conf1Raw, err := json.Marshal(conf1)
	if err != nil {
		return false, err
	}

	conf2Raw, err := json.Marshal(conf2)
	if err != nil {
		return false, err
	}

	conf1Hash, err := util.SHA256Hash(bytes.NewReader(conf1Raw))
	if err != nil {
		return false, err
	}

	conf2Hash, err := util.SHA256Hash(bytes.NewReader(conf2Raw))
	if err != nil {
		return false, err
	}

	if conf1Hash == conf2Hash {
		return true, nil
	}

	return false, nil
}

func Set(path string, conf *Config) error {
	configF, err := os.Create(path)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(configF)
	encoder.SetIndent("", "	")
	err = encoder.Encode(conf)
	if err != nil {
		return err
	}

	return nil
}

func GetChainHead() (*Config, error) {
	chainFileEntries, err := os.ReadDir(".sigrun")
	if err != nil {
		return nil, err
	}

	var chainHeight int
	for _, cf := range chainFileEntries {
		chainNumRaw := strings.Replace(cf.Name(), ".json", "", -1)
		chainNum, err := strconv.Atoi(chainNumRaw)
		if err != nil {
			return nil, err
		}

		if chainNum > chainHeight {
			chainHeight = chainNum
		}
	}

	return Get(".sigrun/" + fmt.Sprint(chainHeight) + ".json")
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
func ReadRepos(repoUrls ...string) (map[string]*Config, error) {
	pathToConfig := make(map[string]*Config)
	for _, path := range repoUrls {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		var conf Config
		err = json.NewDecoder(resp.Body).Decode(&conf)
		if err != nil {
			return nil, err
		}

		pathToConfig[path] = &conf
	}

	return pathToConfig, nil
}

func Create(conf *Config) error {
	conf.ChainNo = 0
	conf.Signature = ""
	err := Set(FILE_NAME, conf)
	if err != nil {
		return err
	}

	err = os.Mkdir(".sigrun", os.ModePerm)
	if err != nil {
		return err
	}

	return Set(".sigrun/0.json", conf)
}

func VerifyChain(oldPubK, oldPath string, oldChainNo int64, newConf *Config) error {
	currentChainNo := oldChainNo + 1
	prevPubK := oldPubK

	var currConf *Config
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

	isSame, err := IsSame(currConf, newConf)
	if err != nil {
		return err
	}

	if !isSame {
		return fmt.Errorf("chain head is not the same as config file")
	}

	return nil
}
