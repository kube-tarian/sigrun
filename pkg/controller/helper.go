package controller

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// GenerateCACert creates the self-signed CA cert and private key
// it will be used to sign the webhook server certificate
func GenerateCACert(certValidityDuration time.Duration) ([]byte, *rsa.PrivateKey, error) {
	now := time.Now()
	begin := now.Add(-1 * time.Hour)
	end := now.Add(certValidityDuration)

	templ := &x509.Certificate{
		SerialNumber: big.NewInt(0),
		Subject: pkix.Name{
			CommonName: "*.sigrun.svc",
		},
		NotBefore:             begin,
		NotAfter:              end,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating key: %v", err)
	}

	der, err := x509.CreateCertificate(rand.Reader, templ, templ, key.Public(), key)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating certificate: %v", err)
	}

	return der, key, nil
}

// GenerateCertPem takes the results of GenerateCACert and uses it to create the
// PEM-encoded public certificate and private key, respectively
func GenerateCertPem(caCertRaw []byte, caKey *rsa.PrivateKey, certValidityDuration time.Duration) ([]byte, *rsa.PrivateKey, error) {
	caCert, err := x509.ParseCertificate(caCertRaw)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	begin := now.Add(-1 * time.Hour)
	end := now.Add(certValidityDuration)

	templ := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "sigrun.default.svc",
		},
		DNSNames:              []string{"sigrun.default.svc"},
		NotBefore:             begin,
		NotAfter:              end,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating key for webhook %v", err)
	}
	der, err := x509.CreateCertificate(rand.Reader, templ, caCert, key.Public(), caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating certificate for webhook %v", err)
	}

	return der, key, nil
}

// PrivateKeyToPem Creates PEM block from private key object
func PrivateKeyToPem(rsaKey *rsa.PrivateKey) []byte {
	privateKey := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
	}

	return pem.EncodeToMemory(privateKey)
}

// CertificateToPem ...
func CertificateToPem(certificateDER []byte) []byte {
	certificate := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificateDER,
	}

	return pem.EncodeToMemory(certificate)
}
