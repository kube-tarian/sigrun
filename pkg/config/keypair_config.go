package config

import (
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/sigstore/sigstore/pkg/signature"

	"github.com/pkg/errors"
	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"

	"github.com/sigstore/cosign/pkg/cosign"
)

type KeyPair struct {
	Name       string
	Mode       string
	ChainNo    int64
	PublicKey  string
	PrivateKey string
	Images     []string
	Signature  string
}

func (conf *KeyPair) GetSignature() string {
	return conf.Signature
}

func (conf *KeyPair) VerifySuccessorConfig(config Config) error {
	data, err := config.SignDoc()
	if err != nil {
		return err
	}

	pubK, err := cosign.PemToECDSAKey([]byte(conf.PublicKey))
	if err != nil {
		return err
	}
	verifier, err := signature.LoadECDSAVerifier(pubK, crypto.SHA256)
	if err != nil {
		return err
	}

	sigRaw, err := base64.StdEncoding.DecodeString(config.GetSignature())
	if err != nil {
		return err
	}

	return verifier.VerifySignature(bytes.NewReader(sigRaw), bytes.NewReader(data))
}

func (conf *KeyPair) GetVerificationInfo() *VerificationInfo {
	return &VerificationInfo{
		Name:        conf.Name,
		Mode:        conf.Mode,
		ChainNo:     conf.ChainNo,
		PublicKey:   conf.PublicKey,
		Maintainers: nil,
		Images:      conf.Images,
	}
}

func (conf *KeyPair) GetChainNo() int64 {
	return conf.ChainNo
}

func (conf *KeyPair) Sign(data []byte) (string, error) {
	password, err := cosignCLI.GetPass(true)
	if err != nil {
		return "", err
	}

	signer, err := cosignCLI.LoadECDSAPrivateKey([]byte(conf.PrivateKey), []byte(password))
	if err != nil {
		return "", err
	}

	sig, err := signer.SignMessage(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}

func (conf *KeyPair) SignDoc() ([]byte, error) {
	var signDoc = *conf
	signDoc.Signature = ""
	return json.Marshal(signDoc)
}

func (conf *KeyPair) CommitRepositoryUpdate() error {
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

func (conf *KeyPair) SignImages() error {
	tempPrivKeyFile, err := ioutil.TempFile("", "priv-key")
	if err != nil {
		return err
	}
	defer os.Remove(tempPrivKeyFile.Name())
	_, err = io.Copy(tempPrivKeyFile, strings.NewReader(conf.PrivateKey))
	if err != nil {
		return err
	}

	so := cosignCLI.KeyOpts{
		KeyRef:   tempPrivKeyFile.Name(),
		PassFunc: cosignCLI.GetPass,
		RekorURL: REKOR_URL,
	}
	ctx := context.Background()
	for _, img := range conf.Images {
		if err := cosignCLI.SignCmd(ctx, so, nil, img, "", true, "", false, false); err != nil {
			return errors.Wrapf(err, "signing %s", img)
		}
	}

	f, err := os.Open(LEDGER_FILE_NAME)
	if err != nil {
		return err
	}

	var ledger *Ledger
	err = json.NewDecoder(f).Decode(&ledger)
	if err != nil {
		return err
	}

	err = ledger.AddEntry(nil)
	if err != nil {
		return err
	}

	return set(LEDGER_FILE_NAME, ledger)
}

func (conf *KeyPair) InitializeRepository() error {
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

	err = set(LEDGER_FILE_NAME, NewLedger())
	if err != nil {
		return err
	}

	return set(".sigrun/0.json", conf)
}

func (conf *KeyPair) Validate() error {

	return nil
}

func (conf *KeyPair) VerifyImage(image string) error {
	image, err := NormalizeImageName(image)
	if err != nil {
		return err
	}

	key := []byte(conf.PublicKey)
	pubKey, err := decodePEM(key)
	if err != nil {
		return errors.Wrapf(err, "failed to decode PEM %v", string(key))
	}

	cosignOpts := &cosign.CheckOpts{
		Annotations: map[string]interface{}{},
		SigVerifier: pubKey,
		RegistryClientOpts: []remote.Option{
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
		},
	}

	ref, err := name.ParseReference(image)
	if err != nil {
		return errors.Wrap(err, "failed to parse image")
	}

	_, err = cosign.Verify(context.Background(), ref, cosignOpts)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "NAME_UNKNOWN: repository name not known to registry") {
			return fmt.Errorf("signature not found")
		} else if strings.Contains(msg, "no matching signatures") {
			return fmt.Errorf("invalid signature")
		}
		return errors.Wrap(err, "failed to verify image")
	}

	return nil
}

func decodePEM(raw []byte) (signature.Verifier, error) {
	// PEM encoded file.
	ed, err := cosign.PemToECDSAKey(raw)
	if err != nil {
		return nil, errors.Wrap(err, "pem to ecdsa")
	}

	return signature.LoadECDSAVerifier(ed, crypto.SHA256)
}
