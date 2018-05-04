package pkiutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

type SigningProfile uint8

const keyPEMType = "RSA PRIVATE KEY"
const certPEMType = "CERTIFICATE"

const (
	None SigningProfile = iota
	Server
	Peer
	Client
)

type CSRParams struct {
	Subject  pkix.Name
	AltNames []string
	ValidFor time.Duration
	Profile  SigningProfile
}

type CertCombo struct {
	Cert *x509.Certificate
	Key  *rsa.PrivateKey
}

const defaultRSAKeyLength = 2048

func InitCA(params CSRParams) *CertCombo {
	priv := makePrivateKey()
	cert := makeCACertOrDie(params, priv)

	return &CertCombo{
		Cert: cert,
		Key:  priv,
	}
}

func MakeCert(params CSRParams, caCombo *CertCombo) *CertCombo {
	priv := makePrivateKey()

	csr := makeCSROrDie(params, priv)
	cert := signCertOrDie(params, csr, caCombo)

	// TODO: validate key and cert?

	return &CertCombo{
		Cert: cert,
		Key:  priv,
	}
}

func (certCombo *CertCombo) ExtractCertData() []byte {
	derBytes := certCombo.Cert.Raw
	return convertToPEM(certPEMType, derBytes)
}

func (certCombo *CertCombo) ExtractKeyData() []byte {
	derBytes := x509.MarshalPKCS1PrivateKey(certCombo.Key)
	return convertToPEM(keyPEMType, derBytes)
}

func (certCombo *CertCombo) SetCertPEMData(certPEM []byte) error {
	derBytes, err := convertFromPEM(certPEMType, certPEM)
	if err != nil {
		return fmt.Errorf("Invalid PEM data: %v", err)
	}
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return fmt.Errorf("Failed to parse certificate: %v", err)
	}
	certCombo.Cert = cert
	return nil
}

func (certCombo *CertCombo) SetKeyPEMData(keyPEM []byte) error {
	derBytes, err := convertFromPEM(keyPEMType, keyPEM)
	if err != nil {
		return fmt.Errorf("Invalid PEM data: %v", err)
	}
	key, err := x509.ParsePKCS1PrivateKey(derBytes)
	if err != nil {
		return fmt.Errorf("Failed to parse key: %v", err)
	}
	certCombo.Key = key
	return nil
}

func makePrivateKey() *rsa.PrivateKey {
	priv, err := rsa.GenerateKey(rand.Reader, defaultRSAKeyLength)
	if err != nil {
		log.Fatalf("failed to generate ECDSA private key: %v", err)
		os.Exit(2)
	}
	return priv
}

func makeSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("failed to generate serial number: %s", err)
		os.Exit(2)
	}
	return serialNumber
}

func makeCSROrDie(params CSRParams, priv *rsa.PrivateKey) *x509.CertificateRequest {
	template := x509.CertificateRequest{
		Subject:            params.Subject,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	// set DNS names and IP addresses
	for _, h := range params.AltNames {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	if err != nil {
		panic(fmt.Errorf("failed to make CSR: %v", err))
	}

	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		panic(fmt.Errorf("failed to parse CSR (this should not happen): %v", err))
	}

	return csr
}

func signCertOrDie(params CSRParams, csr *x509.CertificateRequest, caCombo *CertCombo) *x509.Certificate {
	// check csr
	if err := csr.CheckSignature(); err != nil {
		panic(fmt.Errorf("invalid CSR signature: %v", err))
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(params.ValidFor)
	serialNumber := makeSerialNumber()

	template := x509.Certificate{
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,
		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,
		Subject:            csr.Subject,
		DNSNames:           csr.DNSNames,
		IPAddresses:        csr.IPAddresses,
		SerialNumber:       serialNumber,
		Issuer:             caCombo.Cert.Subject,
		NotBefore:          notBefore,
		NotAfter:           notAfter,
	}

	switch params.Profile {
	case Server:
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	case Client:
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	case Peer:
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}
	}

	crtBytes, err := x509.CreateCertificate(rand.Reader, &template, caCombo.Cert, csr.PublicKey, caCombo.Key)
	if err != nil {
		panic(fmt.Errorf("failed to sign certificate: %v", err))
	}

	cert, err := x509.ParseCertificate(crtBytes)
	if err != nil {
		panic(fmt.Errorf("failed to parse certificate (this should not happen): %s", err))
	}

	return cert
}

func makeCACertOrDie(params CSRParams, priv *rsa.PrivateKey) *x509.Certificate {
	notBefore := time.Now()
	notAfter := notBefore.Add(params.ValidFor)
	serialNumber := makeSerialNumber()

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               params.Subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA: true,
	}

	crtBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(fmt.Errorf("failed to sign CA cert: %v", err))
	}

	cert, err := x509.ParseCertificate(crtBytes)
	if err != nil {
		panic(fmt.Errorf("failed to parse CA cert (this should not happen): %v", err))
	}

	return cert
}

func convertToPEM(pemBlockType string, derBytes []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: pemBlockType, Bytes: derBytes})
}

func convertFromPEM(pemBlockType string, data []byte) ([]byte, error) {
	p, rest := pem.Decode(data)
	if len(rest) > 0 {
		return nil, fmt.Errorf("unparsed bytes")
	}
	if p.Type != pemBlockType {
		return nil, fmt.Errorf("invalid PEM block type %s", p.Type)
	}
	return p.Bytes, nil
}
