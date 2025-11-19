package amazonseshandler

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
)

// https://github.com/robbiet480/go.sns/issues/2
var hostPattern = regexp.MustCompile(`^sns\.[a-zA-Z0-9\-]{3,}\.amazonaws\.com(\.cn)?$`)

// BuildSignature returns a byte array containing a signature usable for SNS verification
func (payload *Payload) BuildSignature() []byte {
	var builtSignature bytes.Buffer
	signableKeys := []string{"Message", "MessageId", "Subject", "SubscribeURL", "Timestamp", "Token", "TopicArn", "Type"}
	for _, key := range signableKeys {
		reflectedStruct := reflect.ValueOf(payload)
		field := reflect.Indirect(reflectedStruct).FieldByName(key)
		value := field.String()
		if field.IsValid() && value != "" {
			builtSignature.WriteString(key + "\n")
			builtSignature.WriteString(value + "\n")
		}
	}
	return builtSignature.Bytes()
}

// VerifyPayload will verify that a payload came from SNS
func (payload *Payload) VerifyPayload() error {
	payloadSignature, err := base64.StdEncoding.DecodeString(payload.Signature)
	if err != nil {
		return err
	}

	if len(UnitTestCertificate) > 0 {
		// works with unit tests
		// cert is in PEM format
		certBlock, _ := pem.Decode(UnitTestCertificate)
		if certBlock == nil {
			return fmt.Errorf("failed to decode certificate PEM")
		}
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %v", err)
		}
		return cert.CheckSignature(x509.SHA1WithRSA, payload.BuildSignature(), payloadSignature)
	}

	if payload.SigningCertURL == "" {
		return errors.New("payload does not have a SigningCertURL")
	}

	certURL, err := url.Parse(payload.SigningCertURL)
	if err != nil {
		return err
	}

	if certURL.Scheme != "https" {
		return fmt.Errorf("url should be using https")
	}

	if !hostPattern.Match([]byte(certURL.Host)) {
		// check if cert
		return fmt.Errorf("certificate is located on an invalid domain")
	}

	resp, err := http.Get(payload.SigningCertURL)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	decodedPem, _ := pem.Decode(body)
	if decodedPem == nil {
		return errors.New("the decoded PEM file was empty")
	}

	parsedCertificate, err := x509.ParseCertificate(decodedPem.Bytes)
	if err != nil {
		return err
	}

	return parsedCertificate.CheckSignature(x509.SHA1WithRSA, payload.BuildSignature(), payloadSignature)
}

// Subscribe will use the SubscribeURL in a payload to confirm a subscription and return a ConfirmSubscriptionResponse
func (payload *Payload) Subscribe() (ConfirmSubscriptionResponse, error) {
	var response ConfirmSubscriptionResponse
	if payload.SubscribeURL == "" {
		return response, errors.New("Payload does not have a SubscribeURL")
	}

	resp, err := http.Get(payload.SubscribeURL)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	xmlErr := xml.Unmarshal(body, &response)
	if xmlErr != nil {
		return response, xmlErr
	}
	return response, nil
}

// Unsubscribe will use the UnsubscribeURL in a payload to confirm a subscription and return a UnsubscribeResponse
func (payload *Payload) Unsubscribe() (UnsubscribeResponse, error) {
	var response UnsubscribeResponse
	resp, err := http.Get(payload.UnsubscribeURL)
	if err != nil {
		return response, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	xmlErr := xml.Unmarshal(body, &response)
	if xmlErr != nil {
		return response, xmlErr
	}
	return response, nil
}
