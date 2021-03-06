package config

import (
	"bytes"
	"crypto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/sigstore/pkg/signature"

	"github.com/devopstoday11/sigrun/pkg/util"
)

func set(path string, content interface{}) error {
	configF, err := os.Create(path)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(configF)
	encoder.SetIndent("", "	")
	err = encoder.Encode(content)
	if err != nil {
		return err
	}

	return nil
}

const CONFIG_FILE_NAME = "sigrun-config.json"
const LEDGER_FILE_NAME = "sigrun-ledger.json"

func detectConfigType(encodedConfig []byte) (string, error) {
	var obj struct {
		Mode string
	}
	err := json.Unmarshal(encodedConfig, &obj)
	if err != nil {
		return "", err
	}

	return obj.Mode, nil
}

func parseConfig(encodedConfig []byte) (Config, error) {
	mode, err := detectConfigType(encodedConfig)
	if err != nil {
		return nil, err
	}

	var conf Config
	if mode == CONFIG_MODE_KEYLESS {
		var keylessConfig Keyless
		err = json.Unmarshal(encodedConfig, &keylessConfig)
		if err != nil {
			return nil, err
		}

		conf = &keylessConfig
	} else {
		var defaultConfig KeyPair
		err = json.Unmarshal(encodedConfig, &defaultConfig)
		if err != nil {
			return nil, err
		}

		conf = &defaultConfig
	}

	return conf, err
}

func getChainHead() (Config, error) {
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

	encodedConfig, err := ioutil.ReadFile(".sigrun/" + fmt.Sprint(chainHeight) + ".json")
	if err != nil {
		return nil, err
	}

	return parseConfig(encodedConfig)
}

func isSame(conf1, conf2 Config) (bool, error) {
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
func decodePEM(raw []byte) (signature.Verifier, error) {
	// PEM encoded file.
	ed, err := cosign.PemToECDSAKey(raw)
	if err != nil {
		return nil, errors.Wrap(err, "pem to ecdsa")
	}

	return signature.LoadECDSAVerifier(ed, crypto.SHA256)
}
