package amazonseshandler

import "encoding/xml"

var UnitTestCertificate = []byte{}

// Payload contains a single POST from SNS
type Payload struct {
	Message          string `json:"Message"`
	MessageId        string `json:"MessageId"`
	Signature        string `json:"Signature"`
	SignatureVersion string `json:"SignatureVersion"`
	SigningCertURL   string `json:"SigningCertURL"`
	SubscribeURL     string `json:"SubscribeURL"`
	Subject          string `json:"Subject"`
	Timestamp        string `json:"Timestamp"`
	Token            string `json:"Token"`
	TopicArn         string `json:"TopicArn"`
	Type             string `json:"Type"`
	UnsubscribeURL   string `json:"UnsubscribeURL"`
}

// ConfirmSubscriptionResponse contains the XML response of accessing a SubscribeURL
type ConfirmSubscriptionResponse struct {
	XMLName         xml.Name `xml:"ConfirmSubscriptionResponse"`
	SubscriptionArn string   `xml:"ConfirmSubscriptionResult>SubscriptionArn"`
	RequestId       string   `xml:"ResponseMetadata>RequestId"`
}

// UnsubscribeResponse contains the XML response of accessing an UnsubscribeURL
type UnsubscribeResponse struct {
	XMLName   xml.Name `xml:"UnsubscribeResponse"`
	RequestId string   `xml:"ResponseMetadata>RequestId"`
}

// MessageJSON contains the JSON response of a notification
var AwsCredCtxKey = &AwsContextKey{"awsCred"}

type AwsContextKey struct {
	Name string
}

type AwsCredentials struct {
	AwsSecretKey string
	AwsAccessKey string
}

// Amazon SNS Received message parsing for email
type Mail struct {
	Timestamp        string             `json:"timestamp"`
	Source           string             `json:"source"`
	MessageID        string             `json:"messageId"`
	Destination      []string           `json:"destination"`
	HeadersTruncated bool               `json:"headersTruncated"`
	Headers          []*HeaderAttribute `json:"headers"`
	CommonHeaders    *CommonHeader      `json:"commonHeaders"`
}
type MessageJSON struct {
	NotificationType string   `json:"notificationType"`
	Mail             *Mail    `json:"mail,omitempty"`
	Receipt          *Receipt `json:"receipt,omitempty"`
	Content          string   `json:"content,omitempty"`
}

type HeaderAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type CommonHeader struct {
	ReturnPath string   `json:"returnPath"`
	From       []string `json:"from"`
	Date       string   `json:"date"`
	To         []string `json:"to"`
	MessageID  string   `json:"messageId"`
	Subject    string   `json:"subject"`
}

type Receipt struct {
	Timestamp            string         `json:"timestamp"`
	ProcessingTimeMillis int            `json:"processingTimeMillis"`
	Recipients           []string       `json:"recipients"`
	SpamVerdict          *VerdictStatus `json:"spamVerdict"`
	VirusVerdict         *VerdictStatus `json:"virusVerdict"`
	SpfVerdict           *VerdictStatus `json:"spfVerdict"`
	DkimVerdict          *VerdictStatus `json:"dkimVerdict"`
	DmarcVerdict         *VerdictStatus `json:"dmarcVerdict"`
	Action               *Action        `json:"action"`
}

type Action struct {
	Type            string `json:"type"`
	TopicArn        string `json:"topicArn"`
	BucketName      string `json:"bucketName"`
	ObjectKeyPrefix string `json:"objectKeyPrefix"`
	ObjectKey       string `json:"objectKey"`
}

type VerdictStatus struct {
	Status string `json:"status"`
}

// UserBounceReason - bounce type
type UserBounceReason struct {
	Email        string
	BounceReason string
}

type PlainTextEnvelope struct {
	AwsRef      string                 `json:"awsRef"`
	Subject     string                 `json:"subject"`
	Email       string                 `json:"email"`
	Attachments []*PlainTextAttachment `json:"attachments"`
}

type PlainTextAttachment struct {
	AwsKey      string `json:"awsKey"`
	Size        uint32 `json:"size"`
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
}
