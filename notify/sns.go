package notify

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNS struct {
	Region     string
	TopicArn   string
	SNSNew     func(client.ConfigProvider, ...*aws.Config) *sns.SNS
	SNSPublish func(svc *sns.SNS, params *sns.PublishInput) (*sns.PublishOutput, error)
}

// コンストラクタ
func NewSNS(region string, topicArn string) *SNS {
	inst := &SNS{
		Region:   region,
		TopicArn: topicArn,
		SNSNew:   sns.New,
		SNSPublish: func(svc *sns.SNS, params *sns.PublishInput) (*sns.PublishOutput, error) {
			return svc.Publish(params)
		},
	}
	return inst
}

// AWS SNSに通知します。
func (self *SNS) NotifySNS(subject string, message string, status int) error {

	// snsを初期化
	svc := self.SNSNew(session.New(), &aws.Config{Region: aws.String(self.Region)})

	// パラメータを準備
	params := &sns.PublishInput{
		Message:  aws.String(message),
		Subject:  aws.String(subject),
		TopicArn: aws.String(self.TopicArn),
	}

	// SNSにpublish
	_, err := self.SNSPublish(svc, params)
	if err != nil {
		return err
	}

	return nil
}
