package amazonseshandler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	abi "github.com/mailio/go-mailio-smtp-abi"
	helpers "github.com/mailio/go-mailio-smtp-helpers"
)

const MaxNumberOfRecipients = 20

type AmazonSESHandler struct {
	s3Client  *s3.Client
	sesClient *ses.Client
}

func NewAmazonSESHandler(config aws.Config) *AmazonSESHandler {
	s3Client := s3.NewFromConfig(config)
	sesClient := ses.NewFromConfig(config)
	return &AmazonSESHandler{
		s3Client:  s3Client,
		sesClient: sesClient,
	}
}

// ReceiveMail - receive mail from Amazon SES
func (m *AmazonSESHandler) ReceiveMail(request http.Request) (*abi.Mail, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	defer request.Body.Close()

	// request.Body = io.NopCloser(bytes.NewBuffer(body)) // reset the body to the original body

	var payload Payload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		return nil, err
	}

	if err := payload.VerifyPayload(); err != nil {
		return nil, err
	}

	switch payload.Type {
	case "SubscriptionConfirmation":
		_, err := payload.Subscribe()
		if err != nil {
			return nil, err
		}
		return nil, nil
	case "Notification":
		message := payload.Message
		var messageJSON MessageJSON
		err := json.Unmarshal([]byte(message), &messageJSON)
		if err != nil {
			return nil, err
		}
		switch messageJSON.NotificationType {
		case "Received":
			// mail := messageJSON.Mail
			receipt := messageJSON.Receipt
			mimeContent := messageJSON.Content

			var mime []byte
			var parsed *abi.Mail
			if mimeContent != "" {
				mime = []byte(mimeContent)
				parsed, err = helpers.ParseMime(mime)
				if err != nil {
					return nil, err
				}
			} else {
				bucket, key := ExtractBucketAndKey(receipt)
				if bucket == "" || key == "" {
					return nil, errors.New("bucket and key or mime content are required")
				}
				mime, err = DownloadS3Object(m.s3Client, bucket, key)
				if err != nil {
					return nil, err
				}
				parsed, err = helpers.ParseMime(mime)
				if err != nil {
					return nil, err
				}
			}
			if receipt.SpamVerdict != nil {
				spamMailio := "PASS"
				isSpam, err := CheckSpam(receipt) // ingore harmful emails
				if err != nil {
					return nil, err
				}
				if isSpam {
					spamMailio = "FAIL"
				}
				parsed.SpamVerdict = &abi.VerdictStatus{
					Status: spamMailio,
				}
			}
			if receipt.SpfVerdict != nil {
				parsed.SpfVerdict = &abi.VerdictStatus{
					Status: receipt.SpfVerdict.Status,
				}
			}
			if receipt.DkimVerdict != nil {
				parsed.DkimVerdict = &abi.VerdictStatus{
					Status: receipt.DkimVerdict.Status,
				}
			}
			if receipt.DmarcVerdict != nil {
				parsed.DmarcVerdict = &abi.VerdictStatus{
					Status: receipt.DmarcVerdict.Status,
				}
			}
			parsed.RawMime = mime

			//TODO!: move the mime object? Or delete it maybe? Or just re-configure in the SNS/SES topic to upload it to different bucket
			//TODO! mailio-user-received-eml-production (then i can remove it from the server code)
			//TODO! and also maybe transfer raw mime here?
			return parsed, nil
		case "Bounce", "Complaint", "Delivery", "Reject", "Send":
			return nil, nil
		}
	}

	return nil, errors.New("unknown payload type")
}
