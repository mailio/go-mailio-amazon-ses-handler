package amazonseshandler

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"
)

// getTestCert creates a self-signed X.509 certificate and returns:
//   - certPEM: PEM-encoded certificate bytes (what your HTTP endpoint should return)
//   - privKey: private key to sign your payload with
func getTestCert() (*x509.Certificate, *rsa.PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// 2. Create certificate template
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "Test Cert",
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		// This matches what you're using in CheckSignature
		SignatureAlgorithm: x509.SHA1WithRSA,
	}

	// Self-signed (template is both cert + parent)
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, err
	}

	xcert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, nil, err
	}

	return xcert, priv, nil
}

func signPayload(privKey *rsa.PrivateKey, payload Payload) ([]byte, error) {
	payloadToSign := payload.BuildSignature()
	h := sha1.Sum(payloadToSign) // [20]byte
	return rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA1, h[:])
}

func getNotificationReceivedMessage(notificationReceivedMessagePath string) (*Payload, error) {
	notificationReceivedMessage, err := os.ReadFile("test_data/" + notificationReceivedMessagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read notification received json: %v", err)
	}

	tmpl, err := os.ReadFile("test_data/notification_received.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %v", err)
	}

	var templatePayload Payload
	err = json.Unmarshal(tmpl, &templatePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification received message: %v", err)
	}

	templatePayload.Message = string(notificationReceivedMessage)

	return &templatePayload, nil
}
