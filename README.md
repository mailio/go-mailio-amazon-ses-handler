# Go-mailio-amazon-ses-handler

A Go library for handling Amazon SES incoming email notifications via Amazon SNS. This handler processes SNS notifications, verifies payload signatures, downloads email content from S3, and parses MIME messages with spam and authentication verdicts.

## Features

- ✅ **SNS Payload Verification**: Automatically verifies SNS message signatures to ensure authenticity
- ✅ **Subscription Confirmation**: Handles SNS subscription confirmation requests
- ✅ **Email Processing**: Processes incoming email notifications from Amazon SES
- ✅ **S3 Integration**: Downloads email content from S3 buckets when configured
- ✅ **MIME Parsing**: Parses MIME email content with support for attachments
- ✅ **Security Verdicts**: Extracts spam, SPF, DKIM, and DMARC verdicts from SES receipts
- ✅ **Spam Detection**: Implements spam filtering logic based on SES verdicts

## Installation

```bash
go get github.com/mailio/go-mailio-amazon-ses-handler
```

## Usage

### Basic Example

```go
package main

import (
    "net/http"
    "github.com/aws/aws-sdk-go-v2/aws"
    amazonseshandler "github.com/mailio/go-mailio-amazon-ses-handler"
)

func main() {
    // Initialize AWS config
    cfg := aws.Config{
        Region: "us-east-1",
        // Add your AWS credentials here
    }

    // Create handler
    handler := amazonseshandler.NewAmazonSESHandler(cfg)

    // HTTP handler function
    http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
        mail, err := handler.ReceiveMail(*r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        if mail == nil {
            // Subscription confirmation or other notification type
            w.WriteHeader(http.StatusOK)
            return
        }

        // Process the email
        // mail contains parsed MIME content, verdicts, and raw MIME data
        w.WriteHeader(http.StatusOK)
    })

    http.ListenAndServe(":8080", nil)
}
```

### Processing Email Data

The handler returns a `*abi.Mail` object that contains:

- **Parsed MIME content**: Headers, body, attachments
- **Security verdicts**: Spam, SPF, DKIM, DMARC status
- **Raw MIME**: Original email content as bytes

```go
mail, err := handler.ReceiveMail(*r)
if err != nil {
    // Handle error
}

// Access parsed email data
subject := mail.Subject
from := mail.From
body := mail.Body

// Check security verdicts
if mail.SpamVerdict != nil {
    spamStatus := mail.SpamVerdict.Status // "PASS" or "FAIL"
}

// Access raw MIME content
rawMime := mail.RawMime
```

## AWS Setup Instructions

This section provides step-by-step instructions for configuring Amazon SES to receive emails, store them in S3, send notifications via SNS, and handle failures with a Dead Letter Queue.

### Prerequisites

- An active AWS account
- AWS CLI configured with appropriate permissions
- Permissions to create and manage:
  - Amazon SES (receiving email)
  - Amazon S3 (bucket for email storage)
  - Amazon SNS (notifications)
  - Amazon SQS (Dead Letter Queue)

### Step 1: Create an S3 Bucket for Email Storage

1. **Navigate to Amazon S3 Console**
   - Go to [Amazon S3 console](https://console.aws.amazon.com/s3/)

2. **Create a New Bucket**
   - Click "Create bucket"
   - Enter a unique bucket name (e.g., `my-email-storage-bucket`)
   - Select your preferred AWS region
   - Click "Create bucket"

3. **Configure Bucket Policy for SES**
   - Go to your bucket → Permissions → Bucket policy
   - Add the following policy (replace `YOUR_ACCOUNT_ID` and `BUCKET_NAME`):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowSESPuts",
      "Effect": "Allow",
      "Principal": {
        "Service": "ses.amazonaws.com"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:aws:s3:::BUCKET_NAME/*",
      "Condition": {
        "StringEquals": {
          "aws:Referer": "YOUR_ACCOUNT_ID"
        }
      }
    }
  ]
}
```

For more details, see: [Amazon SES Receiving Email Permissions](https://docs.aws.amazon.com/ses/latest/dg/receiving-email-permissions.html)

### Step 2: Verify Your Domain in Amazon SES

1. **Navigate to Amazon SES Console**
   - Go to [Amazon SES console](https://console.aws.amazon.com/ses/)
   - Select "Verified identities" from the left menu

2. **Verify Your Domain**
   - Click "Create identity"
   - Select "Domain"
   - Enter your domain name
   - Follow the DNS verification steps to add the required DNS records

3. **Request Production Access** (if in SES Sandbox)
   - If you're in the SES sandbox, request production access to receive emails from any address
   - Go to "Account dashboard" → "Request production access"

### Step 3: Create an SNS Topic for Notifications

1. **Navigate to Amazon SNS Console**
   - Go to [Amazon SNS console](https://console.aws.amazon.com/sns/)

2. **Create a Topic**
   - Click "Create topic"
   - Choose "Standard" topic type
   - Enter a topic name (e.g., `ses-email-notifications`)
   - Click "Create topic"

3. **Note the Topic ARN**
   - Copy the Topic ARN (e.g., `arn:aws:sns:us-east-1:123456789012:ses-email-notifications`)
   - You'll need this when configuring SES receipt rules

### Step 4: Create a Dead Letter Queue (DLQ)

1. **Navigate to Amazon SQS Console**
   - Go to [Amazon SQS console](https://console.aws.amazon.com/sqs/)

2. **Create a Queue for DLQ**
   - Click "Create queue"
   - Choose "Standard" queue type
   - Enter a queue name (e.g., `ses-notifications-dlq`)
   - Configure retention period (default: 4 days, max: 14 days)
   - Click "Create queue"

3. **Note the Queue ARN**
   - Copy the Queue ARN for later use

### Step 5: Subscribe Your Endpoint to SNS Topic

1. **Create an HTTP/HTTPS Subscription**
   - In your SNS topic, click "Create subscription"
   - Protocol: "HTTPS" (or "HTTP" for development)
   - Endpoint: Your webhook URL (e.g., `https://your-domain.com/webhook`)
   - Click "Create subscription"

2. **Confirm the Subscription**
   - AWS will send a subscription confirmation request to your endpoint
   - The handler automatically confirms subscriptions when it receives a `SubscriptionConfirmation` message
   - Verify the subscription status shows "Confirmed" in the SNS console

3. **Configure Dead Letter Queue for Subscription**
   - In your SNS topic, select the subscription
   - Click "Edit" → "Redrive policy"
   - Enable "Dead-letter queue"
   - Select your DLQ queue
   - Set "Maximum receives" (recommended: 3-5)
   - Click "Save changes"

For more details, see: [Amazon SNS Dead-Letter Queues](https://docs.aws.amazon.com/sns/latest/dg/sns-dead-letter-queues.html)

### Step 6: Configure SES Receipt Rules

1. **Navigate to Amazon SES Console**
   - Go to [Amazon SES console](https://console.aws.amazon.com/ses/)
   - Select "Email receiving" → "Rule sets" from the left menu

2. **Create or Edit a Rule Set**
   - If you don't have a rule set, create one and set it as active
   - Click on your rule set to edit it

3. **Create a Receipt Rule**
   - Click "Create rule"
   - **Rule name**: Enter a name (e.g., `store-and-notify`)
   
   - **Recipients**: 
     - Add email addresses or domains that should trigger this rule
     - Example: `@yourdomain.com` or `incoming@yourdomain.com`
   
   - **Actions** (add multiple actions):
     
     **Action 1: Store in S3**
     - Click "Add action" → "S3"
     - Select your S3 bucket
     - Optionally set an object key prefix (e.g., `emails/`)
     - Click "Save"
     
     **Action 2: Publish to SNS**
     - Click "Add action" → "SNS"
     - Select your SNS topic ARN
     - Click "Save"
   
   - Click "Create rule"

4. **Verify Rule Order**
   - Rules are processed in order
   - Ensure your rule is positioned correctly in the rule set

### Step 7: Configure MX Records

1. **Get Your SES Receiving Endpoint**
   - In SES console, go to "Email receiving" → "Rule sets"
   - Note the receiving endpoint (e.g., `inbound-smtp.us-east-1.amazonaws.com`)

2. **Update DNS MX Records**
   - In your domain's DNS settings, add an MX record:
     - **Name**: `@` or your subdomain
     - **Priority**: `10`
     - **Value**: Your SES receiving endpoint

3. **Wait for DNS Propagation**
   - DNS changes can take up to 48 hours to propagate
   - Use `dig` or `nslookup` to verify the MX record

For more details, see: [Amazon SES Receiving Email Concepts](https://docs.aws.amazon.com/ses/latest/dg/receiving-email-concepts.html)

### Step 8: Test Your Setup

1. **Send a Test Email**
   - Send an email to an address configured in your receipt rule
   - Example: `test@yourdomain.com`

2. **Verify Email Storage**
   - Check your S3 bucket for the stored email
   - The email should appear with a timestamp-based object key

3. **Check SNS Notifications**
   - Verify your webhook endpoint receives the SNS notification
   - Check CloudWatch Logs for any errors

4. **Test Dead Letter Queue**
   - Temporarily disable your webhook endpoint
   - Send a test email
   - After the maximum retry attempts, check your DLQ for the message

## Notification Types

The handler processes the following notification types:

- **SubscriptionConfirmation**: Automatically confirms SNS subscriptions
- **Notification** with type:
  - **Received**: Processes incoming emails (downloads from S3, parses MIME, extracts verdicts)
  - **Bounce**: Email bounce notifications (returns nil)
  - **Complaint**: Email complaint notifications (returns nil)
  - **Delivery**: Email delivery notifications (returns nil)
  - **Reject**: Email rejection notifications (returns nil)
  - **Send**: Email send notifications (returns nil)

## Security Verdicts

The handler extracts and processes the following security verdicts from SES:

- **SpamVerdict**: Spam detection status (`PASS` or `FAIL`)
- **VirusVerdict**: Virus detection status (`PASS` or `FAIL`)
- **SPFVerdict**: SPF authentication status (`PASS`, `FAIL`, or `GRAY`)
- **DKIMVerdict**: DKIM authentication status (`PASS` or `FAIL`)
- **DMARCVerdict**: DMARC authentication status (`PASS` or `FAIL`)

The handler implements spam filtering logic:
- Emails are considered spam if: `SpamVerdict = PASS` AND `VirusVerdict = PASS` AND `SPFVerdict = PASS or GRAY` → Not spam
- Emails with viruses are always flagged
- Other combinations default to not spam

## Error Handling

The handler returns errors in the following cases:

- Invalid JSON payload
- Failed SNS signature verification
- Missing S3 bucket/key when MIME content is not included
- S3 download failures
- MIME parsing errors

Always check for errors when calling `ReceiveMail()`:

```go
mail, err := handler.ReceiveMail(*r)
if err != nil {
    log.Printf("Error processing email: %v", err)
    // Handle error appropriately
    return
}
```

## Testing

Run the test suite:

```bash
go test ./...
```

The tests include:
- Subscription confirmation handling
- Notification processing with MIME content
- Signature verification

## Additional Resources

- [Amazon SES Receiving Email Concepts](https://docs.aws.amazon.com/ses/latest/dg/receiving-email-concepts.html)
- [Amazon SES SNS Notifications](https://docs.aws.amazon.com/ses/latest/dg/monitor-sending-activity-using-notifications-sns.html)
- [Amazon SNS Dead-Letter Queues](https://docs.aws.amazon.com/sns/latest/dg/sns-dead-letter-queues.html)
- [Mail Manager – Amazon SES Email Routing and Archiving](https://aws.amazon.com/blogs/messaging-and-targeting/mail-manager-amazon-ses-introduces-new-email-routing-and-archiving-features/)

## License

See [LICENSE](LICENSE) file for details.
