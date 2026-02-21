// Package sns provides stub types for the AWS SNS service.
package sns

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
)

// PublishInput contains the input parameters for the Publish operation.
type PublishInput struct {
	Message  *string
	Subject  *string
	TopicArn *string
}

// PublishOutput contains the output from the Publish operation.
type PublishOutput struct{}

// SNS provides the SNS service client.
type SNS struct{}

// Publish publishes a message to an SNS topic.
func (s *SNS) Publish(input *PublishInput) (*PublishOutput, error) {
	return &PublishOutput{}, nil
}

// New creates a new SNS service client.
func New(p client.ConfigProvider, cfgs ...*aws.Config) *SNS {
	return &SNS{}
}
