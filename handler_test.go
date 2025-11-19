package amazonseshandler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-playground/assert/v2"
)

func TestSubscriptionConfirmation(t *testing.T) {
	handler := NewAmazonSESHandler(aws.Config{
		Region: "us-east-1",
	})

	// load subscription confirmation json
	subscriptionConfirmation, err := os.ReadFile("test_data/subscription_confirmation.json")
	if err != nil {
		t.Fatalf("failed to read subscription confirmation json: %v", err)
	}

	// Create HTTP request mimicking AWS SNS format
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(subscriptionConfirmation))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Set headers as AWS SNS sends them
	req.Header.Set("x-amz-sns-message-type", "SubscriptionConfirmation")
	req.Header.Set("x-amz-sns-message-id", "165545c9-2a5c-472c-8df2-7ff2be2b3b1b")
	req.Header.Set("x-amz-sns-topic-arn", "arn:aws:sns:us-west-2:123456789012:MyTopic")
	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")
	req.Header.Set("Host", "myhost.example.com")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("User-Agent", "Amazon Simple Notification Service Agent")
	req.ContentLength = int64(len(subscriptionConfirmation))

	_, err = handler.ReceiveMail(*req)
	if err != nil && err.Error() != "unknown payload type" {
		assert.MatchRegex(t, err.Error(), "certificate is located on an invalid domain")
	}
}

func TestNotificationReceived(t *testing.T) {
	handler := NewAmazonSESHandler(aws.Config{
		Region: "us-east-1",
	})

	// load notification received json
	p, err := getNotificationReceivedMessage("notification_received_contains_mime.json")
	if err != nil {
		t.Fatalf("failed to get notification received message: %v", err)
	}
	payload := *p
	cert, privKey, err := getTestCert()
	if err != nil {
		t.Fatalf("failed to get test cert: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	UnitTestCertificate = certPEM[:]
	signature, err := signPayload(privKey, payload)
	if err != nil {
		t.Fatalf("failed to sign payload: %v", err)
	}
	payload.Signature = base64.StdEncoding.EncodeToString(signature)
	payload.SigningCertURL = "https://mailio.io/SimpleNotificationServiceMailioTest.pem"

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Set headers as AWS SNS sends them
	req.Header.Set("x-amz-sns-message-type", "Notification")
	req.Header.Set("x-amz-sns-message-id", "165545c9-2a5c-472c-8df2-7ff2be2b3b1b")
	req.Header.Set("x-amz-sns-topic-arn", "arn:aws:sns:us-west-2:123456789012:MyTopic")
	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")
	req.Header.Set("Host", "myhost.example.com")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("User-Agent", "Amazon Simple Notification Service Agent")
	req.ContentLength = int64(len(payloadBytes))

	abi, err := handler.ReceiveMail(*req)
	if err != nil && err.Error() != "unknown payload type" {
		t.Fatalf("failed to receive mail: %v", err)
	}
	assert.Equal(t, abi.SpamVerdict.Status, "PASS")
	assert.Equal(t, abi.SpfVerdict.Status, "PASS")
	assert.Equal(t, abi.DkimVerdict.Status, "PASS")
}
