package config

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sigstore/cosign/pkg/cosign"
)

type Config struct {
	ChainNo    int64
	PublicKey  string
	PrivateKey string
	Images     []string
	Signature  string
}

func (c *Config) Validate() error {
	return nil
}

const CONFIG_FILE_NAME = "sigrun-config.json"

func Read() (*Config, error) {
	configF, err := os.Open(CONFIG_FILE_NAME)
	if err != nil {
		return nil, err
	}

	var conf Config
	err = json.NewDecoder(configF).Decode(&conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

func GetGUID(path string) (string, error) {
	genesisConfPath := strings.Replace(path, CONFIG_FILE_NAME, ".sigrun/0.json", -1)

	resp, err := http.Get(genesisConfPath)
	if err != nil {
		return "", err
	}

	hasher := sha256.New()
	_, err = io.Copy(hasher, resp.Body)
	if err != nil {
		return "", err
	}

	return string(hasher.Sum(nil)), nil
}

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

func Create(conf *Config, password string) error {
	configF, err := os.Create(CONFIG_FILE_NAME)
	if err != nil {
		return err
	}
	conf.Signature = ""
	conf.ChainNo = 0

	confRaw, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	signer, err := cosign.LoadECDSAPrivateKey([]byte(conf.PrivateKey), []byte(password))
	if err != nil {
		return err
	}

	sig, _, err := signer.Sign(context.Background(), confRaw)
	if err != nil {
		return err
	}

	conf.Signature = base64.StdEncoding.EncodeToString(sig)

	encoder := json.NewEncoder(configF)
	encoder.SetIndent("", "	")
	err = encoder.Encode(conf)
	if err != nil {
		return err
	}

	err = os.Mkdir(".sigrun", os.ModeDir)
	if err != nil {
		return err
	}

	chainF, err := os.Create(filepath.Join(".sigrun", "0.json"))
	if err != nil {
		return err
	}

	_, err = io.Copy(chainF, configF)
	if err != nil {
		return err
	}

	return nil
}
