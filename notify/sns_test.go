package notify

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/sns"
	"testing"
)

func TestNewSNS(t *testing.T) {
	expecteds := []struct {
		region   string
		topicarn string
	}{
		{"us-east-1", "arn:aws:sns:ap-northeast-1:9999999999:goron_test"},
		{"ap-northeast-1", "arn:aws:sns:ap-northeast-1:9999999999:goron_test2"},
	}

	for _, s := range expecteds {
		inst := NewSNS(s.region, s.topicarn)

		if s.region != inst.Region {
			t.Errorf("(expected) '%s' != '%s'", s.region, inst.Region)
		}
		if s.topicarn != inst.TopicArn {
			t.Errorf("(expected) '%s' != '%s'", s.topicarn, inst.TopicArn)
		}
	}
}

func TestNotifySNS(t *testing.T) {
	expecteds := []struct {
		region   string
		topicarn string
		subject  string
		message  string
		status   int
	}{
		{"us-east-1", "arn:aws:sns:ap-northeast-1:9999999999:goron_test", "subject", "message", 0},
		{"ap-northeast-1", "arn:aws:sns:ap-northeast-1:9999999999:goron_test2", "サブジェクト", "メッセージ", 1},
	}

	for _, s := range expecteds {
		inst := NewSNS(s.region, s.topicarn)

		// モックを設定
		inst.SNSNew = func(client.ConfigProvider, ...*aws.Config) *sns.SNS {
			return nil
		}
		inst.SNSPublish = func(svc *sns.SNS, params *sns.PublishInput) (*sns.PublishOutput, error) {
			if s.subject != *params.Subject {
				t.Errorf("(expected) '%s' != '%s'", s.subject, *params.Subject)
			}
			if s.message != *params.Message {
				t.Errorf("(expected) '%s' != '%s'", s.message, *params.Message)
			}
			return nil, nil
		}

		err := inst.NotifySNS(s.subject, s.message, s.status)
		if err != nil {
			t.Error(err)
		}
	}
}
