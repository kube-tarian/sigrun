package config

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"

	"github.com/sigstore/cosign/pkg/cosign"
)

type DefaultConfig struct {
	Name       string
	Mode       string
	ChainNo    int64
	PublicKey  string
	PrivateKey string
	Images     []string
	Signature  string
}

func (conf *DefaultConfig) GetChainNo() int64 {
	return conf.ChainNo
}

func (conf *DefaultConfig) Sign(data []byte) (string, error) {
	password, err := cosignCLI.GetPass(true)
	if err != nil {
		return "", err
	}

	signer, err := cosign.LoadECDSAPrivateKey([]byte(conf.PrivateKey), []byte(password))
	if err != nil {
		return "", err
	}

	sig, _, err := signer.Sign(context.Background(), data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

func (conf *DefaultConfig) SignDoc() ([]byte, error) {
	var signDoc = *conf
	signDoc.Signature = ""
	return json.Marshal(signDoc)
}

func (conf *DefaultConfig) CommitRepositoryUpdate() error {
	oldConf, err := getChainHead()
	if err != nil {
		return err
	}

	isSame, err := isSame(conf, oldConf)
	if err != nil {
		return err
	}

	if isSame {
		return fmt.Errorf("config has not changed")
	}

	conf.ChainNo = oldConf.GetChainNo() + 1

	signDoc, err := conf.SignDoc()
	if err != nil {
		return err
	}

	sig, err := oldConf.Sign(signDoc)
	if err != nil {
		return err
	}
	conf.Signature = sig

	err = set(FILE_NAME, conf)
	if err != nil {
		return err
	}

	return set(".sigrun/"+fmt.Sprint(conf.ChainNo)+".json", conf)
}

func (conf *DefaultConfig) SignImages() error {
	tempPrivKeyFile, err := ioutil.TempFile("", "priv-key")
	if err != nil {
		return err
	}
	defer os.Remove(tempPrivKeyFile.Name())
	_, err = io.Copy(tempPrivKeyFile, strings.NewReader(conf.PrivateKey))
	if err != nil {
		return err
	}

	so := cosignCLI.SignOpts{
		KeyRef: tempPrivKeyFile.Name(),
		Pf:     cosignCLI.GetPass,
	}
	ctx := context.Background()
	for _, img := range conf.Images {
		if err := cosignCLI.SignCmd(ctx, so, img, true, "", false, false); err != nil {
			return errors.Wrapf(err, "signing %s", img)
		}
	}

	return nil
}

func (conf *DefaultConfig) InitializeRepository() error {
	conf.ChainNo = 0
	conf.Signature = ""
	err := set(FILE_NAME, conf)
	if err != nil {
		return err
	}

	err = os.Mkdir(".sigrun", os.ModePerm)
	if err != nil {
		return err
	}

	return set(".sigrun/0.json", conf)
}

func (c *DefaultConfig) Validate() error {

	return nil
}
