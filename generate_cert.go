package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"
)

func main() {
	priv, err := generateKey()
	if err != nil {
		panic(err)
	}

	cert, err := generateCertificate(priv)
	if err != nil {
		panic(err)
	}

	saveKeyAndCert(priv, cert)
}

func generateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func generateCertificate(priv *rsa.PrivateKey) ([]byte, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"My Org"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	return x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
}

func saveKeyAndCert(priv *rsa.PrivateKey, cert []byte) {
	privFile, err := os.Create("private_key.pem")
	if err == nil {
		defer privFile.Close()
		pem.Encode(privFile, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	}

	certFile, err := os.Create("certificate.pem")
	if err == nil {
		defer certFile.Close()
		pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	}
}
