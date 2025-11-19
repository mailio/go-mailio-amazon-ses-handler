package amazonseshandler

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ExtractBucketAndKey(receipt *Receipt) (string, string) {
	bucket := ""
	key := ""
	if receipt != nil {
		action := receipt.Action
		if action != nil {
			bucket = action.BucketName
			key = action.ObjectKey
			if action.ObjectKeyPrefix != "" {
				key = action.ObjectKeyPrefix + "/" + action.ObjectKey
			}
		}
	}
	return bucket, key
}

func CheckSpam(receipt *Receipt) (bool, error) {
	spamV := receipt.SpamVerdict
	virusV := receipt.VirusVerdict
	spfV := receipt.SpfVerdict
	//dkimVerdict - not checked
	//dmarcVerdict - not checked
	if spamV.Status == "PASS" && virusV.Status == "PASS" && (spfV.Status == "PASS" || spfV.Status == "GRAY") {
		// not spam
		return false, nil
	} else if virusV.Status == "PASS" {
		// as long as there is no virtus, it will go to spam folders
		return true, nil
	}
	// not spam by default
	return false, nil
}

func DownloadS3Object(s3Client *s3.Client, bucket string, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	result, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
