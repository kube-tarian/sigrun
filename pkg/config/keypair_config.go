package config

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/pkg/errors"
	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"

	"github.com/sigstore/cosign/pkg/cosign"
)

type KeyPair struct {
	Name       string
	Mode       string
	PublicKey  string
	PrivateKey string
	Images     []string
	Signature  string
}

func (conf *KeyPair) GetSignature() string {
	return conf.Signature
}

func (conf *KeyPair) GetVerificationInfo() *VerificationInfo {
	return &VerificationInfo{
		Name:        conf.Name,
		Mode:        conf.Mode,
		PublicKey:   conf.PublicKey,
		Maintainers: nil,
		Images:      conf.Images,
	}
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

func (conf *KeyPair) SignImages(repoPath string, annotations map[string]string) error {
	repoPath = filepath.Clean(repoPath)

	err := os.Chdir(repoPath)
	if err != nil {
		return err
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

	err = ledger.AddEntry(annotations)
	if err != nil {
		return err
	}

	var compatibleAnnotations = make(map[string]interface{})
	for k, v := range annotations {
		compatibleAnnotations[k] = v
	}

	jsonEncodedLedgerEntry, _ := json.Marshal(ledger.Entries[len(ledger.Entries)-1])
	compatibleAnnotations["sigrun-ledger-entry"] = jsonEncodedLedgerEntry

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
		if err := cosignCLI.SignCmd(ctx, so, compatibleAnnotations, img, "", true, "", false, false); err != nil {
			return errors.Wrapf(err, "signing %s", img)
		}
	}

	fmt.Println("Please input password again for ledger signature")
	encodedLedger, _ := json.Marshal(ledger)
	ledgerSig, err := conf.Sign(encodedLedger)
	if err != nil {
		return err
	}

	f, err = os.Create(".sigrun/ledger.sig")
	if err != nil {
		return err
	}

	_, err = io.Copy(f, strings.NewReader(ledgerSig))
	if err != nil {
		return err
	}

	return set(LEDGER_FILE_NAME, ledger)
}

func (conf *KeyPair) InitializeRepository(repoPath string) error {
	repoPath = filepath.Clean(repoPath)

	err := os.MkdirAll(repoPath, 0755)
	if err != nil {
		return err
	}

	err = os.Chdir(repoPath)
	if err != nil {
		return err
	}

	conf.Signature = ""
	err = set(CONFIG_FILE_NAME, conf)
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
