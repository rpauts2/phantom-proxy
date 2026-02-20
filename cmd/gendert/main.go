package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	// Создание директории certs
	if err := os.MkdirAll("certs", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create certs directory: %v\n", err)
		os.Exit(1)
	}

	// Генерация приватного ключа
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate private key: %v\n", err)
		os.Exit(1)
	}

	// Создание сертификата
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate serial number: %v\n", err)
		os.Exit(1)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"PhantomProxy Dev"},
			CommonName:   "phantom.local",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"phantom.local", "*.phantom.local"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create certificate: %v\n", err)
		os.Exit(1)
	}

	// Запись сертификата
	certOut, err := os.Create("certs/cert.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create cert.pem: %v\n", err)
		os.Exit(1)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write cert.pem: %v\n", err)
		os.Exit(1)
	}

	// Запись приватного ключа
	keyOut, err := os.Create("certs/key.pem")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create key.pem: %v\n", err)
		os.Exit(1)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal private key: %v\n", err)
		os.Exit(1)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write key.pem: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ SSL certificates generated successfully!")
	fmt.Println("  - certs/cert.pem")
	fmt.Println("  - certs/key.pem")
}
