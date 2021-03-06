package config

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sigstore/cosign/pkg/cosign"

	"github.com/sigstore/cosign/cmd/cosign/cli/fulcio"

	"github.com/pkg/errors"
	cosignCLI "github.com/sigstore/cosign/cmd/cosign/cli"
	fulcioClient "github.com/sigstore/fulcio/pkg/client"
)

type Keyless struct {
	Name        string
	Mode        string
	Maintainers []string
	Images      []string
}

func (conf *Keyless) VerifyImage(image string) error {
	ctx := context.Background()

	image, err := NormalizeImageName(image)
	if err != nil {
		return err
	}

	ref, err := name.ParseReference(image)
	if err != nil {
		return errors.Wrap(err, "failed to parse image")
	}

	signatureRepo, err := cosignCLI.TargetRepositoryForImage(ref)
	if err != nil {
		return err
	}

	cosignOpts := &cosign.CheckOpts{
		RootCerts:          fulcio.Roots,
		RegistryClientOpts: cosignCLI.DefaultRegistryClientOpts(ctx),
		ClaimVerifier:      cosign.SimpleClaimVerifier,
		RekorURL:           REKOR_URL,
		SignatureRepo:      signatureRepo,
		VerifyBundle:       false,
	}

	payload, err := cosign.Verify(ctx, ref, cosignOpts)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "NAME_UNKNOWN: repository name not known to registry") {
			return fmt.Errorf("signature not found")
		} else if strings.Contains(msg, "no matching signatures") {
			return fmt.Errorf("invalid signature")
		}
		return errors.Wrap(err, "failed to verify image")
	}

	var verified bool
	for _, pl := range payload {
		for _, memail := range conf.Maintainers {
			for _, email := range pl.Cert.EmailAddresses {
				if memail == email {
					verified = true
					break
				}
			}
		}
	}

	if !verified {
		return fmt.Errorf("image was not signed by any of the maintainers")
	}

	return nil
}

func (conf *Keyless) GetVerificationInfo() *VerificationInfo {
	return &VerificationInfo{
		Name:        conf.Name,
		Mode:        conf.Mode,
		PublicKey:   "",
		Maintainers: conf.Maintainers,
		Images:      conf.Images,
	}
}

func (conf *Keyless) InitializeRepository(repoPath string) error {

	err := os.MkdirAll(repoPath, 0755)
	if err != nil {
		return err
	}

	err = os.Chdir(repoPath)
	if err != nil {
		return err
	}

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

	return nil
}

const (
	OICD_ISSUER = "https://oauth2.sigstore.dev/auth"
	REKOR_URL   = "https://rekor.sigstore.dev"
)

func (conf *Keyless) SignImages(repoPath string, annotations map[string]string) error {
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

	so := cosignCLI.KeyOpts{
		KeyRef:           "",
		PassFunc:         cosignCLI.GetPass,
		Sk:               false,
		Slot:             "",
		FulcioURL:        fulcioClient.SigstorePublicServerURL,
		RekorURL:         REKOR_URL,
		IDToken:          "",
		OIDCIssuer:       OICD_ISSUER,
		OIDCClientID:     "sigstore",
		OIDCClientSecret: "",
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

func (conf *Keyless) Sign(msg []byte) (string, error) {
	fulcioServer, err := url.Parse(fulcioClient.SigstorePublicServerURL)
	if err != nil {
		return "", errors.Wrap(err, "parsing Fulcio URL")
	}
	fClient := fulcioClient.New(fulcioServer)
	signer, err := fulcio.NewSigner(context.Background(), "", OICD_ISSUER, "sigstore", fClient)
	if err != nil {
		return "", errors.Wrap(err, "getting key from Fulcio")
	}

	sig, err := signer.SignMessage(bytes.NewReader(msg))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sig), nil
}
