package config

import "time"

const (
	// AWSRegionUSEast1 is a helper constant for AWS configs.
	AWSRegionUSEast1 = "us-east-1"
	// AWSRegionUSWest is a helper constant for AWS configs.
	AWSRegionUSWest = "us-west-1"
)

type (
	// AWS holds common AWS credentials and keys.
	AWS struct {
		SecretKey string `envconfig:"AWS_SECRET_KEY"`
		AccessKey string `envconfig:"AWS_ACCESS_KEY"`

		Region string `envconfig:"AWS_REGION"`
	}

	// SQS holds the info required to work with Amazon SQS
	SQS struct {
		AWS
		QueueName string `envconfig:"AWS_SQS_NAME"`
		// MaxMessages will override the DefaultSQSMaxMessages.
		MaxMessages *int64 `envconfig:"AWS_SQS_MAX_MESSAGES"`
		// TimeoutSeconds will override the DefaultSQSTimeoutSeconds.
		TimeoutSeconds *int64 `envconfig:"AWS_SQS_TIMEOUT_SECONDS"`
		// SleepInterval will override the DefaultSQSSleepInterval.
		SleepInterval *time.Duration `envconfig:"AWS_SQS_SLEEP_INTERVAL"`
		// DeleteBufferSize will override the DefaultSQSDeleteBufferSize.
		DeleteBufferSize *int `envconfig:"AWS_SQS_DELETE_BUFFER_SIZE"`
		// ConsumeBase64 is a flag to signal the subscriber to base64 decode the payload
		// before returning it. If it is not set in the config, the flag will default
		// to 'true'.
		ConsumeBase64 *bool `envconfig:"AWS_SQS_CONSUME_BASE64"`
	}

	// SNS holds the info required to work with Amazon SNS.
	SNS struct {
		AWS
		Topic string `envconfig:"AWS_SNS_TOPIC"`
	}

	// S3 holds the info required to work with Amazon S3.
	S3 struct {
		AWS
		Bucket string `envconfig:"AWS_S3_BUCKET_NAME"`
	}

	// DynamoDB holds some basic info required to work with Amazon DynamoDB.
	DynamoDB struct {
		AWS
		TableName string `envconfig:"AWS_DYNAMODB_TABLE_NAME"`
	}
)

// LoadAWSFromEnv will attempt to load the AWS structs
// from environment variables. If not populated, nil
// is returned.
func LoadAWSFromEnv() (*AWS, *SNS, *SQS, *S3, *DynamoDB) {
	var (
		aws = &AWS{}
		sns = &SNS{}
		sqs = &SQS{}
		s3  = &S3{}
		ddb = &DynamoDB{}
	)
	LoadEnvConfig(aws)
	if aws.AccessKey == "" {
		aws = nil
	}
	LoadEnvConfig(&sns)
	if sns.Topic == "" {
		sns = nil
	}
	LoadEnvConfig(&sqs)
	if sqs.QueueName == "" {
		sqs = nil
	}
	LoadEnvConfig(&s3)
	if s3.Bucket == "" {
		s3 = nil
	}
	LoadEnvConfig(&ddb)
	if ddb.TableName == "" {
		ddb = nil
	}
	return aws, sns, sqs, s3, ddb
}
