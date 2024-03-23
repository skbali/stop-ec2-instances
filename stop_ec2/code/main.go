package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"os"
	"strconv"
	"strings"
	"time"
)

var client *ec2.Client
var snsClient *sns.Client
var topicArn string

var maxHours int

const maxResults = 100

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	client = ec2.NewFromConfig(cfg)
	snsClient = sns.NewFromConfig(cfg)

	topicArn = os.Getenv("TOPIC_ARN")
	maxHours, err = strconv.Atoi(os.Getenv("MAX_HOURS"))
	if err != nil {
		panic("configuration error, " + err.Error())
	}
}

func checkInstance(ins types.Instance) (string, error) {
	var threshold, warnUser int
	var msg string
	id := *ins.InstanceId
	now := time.Now().UTC()

	for _, tag := range ins.Tags {
		if *tag.Key == "AllowStop" {
			hours, err := strconv.Atoi(*tag.Value)
			if err != nil {
				hours = maxHours
			}
			if hours > maxHours {
				hours = maxHours
			}

			if hours > 0 {
				threshold = hours * 3600
			} else {
				threshold = 600
			}

			warnUser = threshold - 420
			ld := now.Sub(*ins.LaunchTime)
			if ld.Seconds() > float64(warnUser) {
				msg = fmt.Sprintf("InstanceID: %v will be shutting down soon! Uptime %v, Allowed %v", *ins.InstanceId, ld, time.Duration(threshold)*time.Second)
				if ld.Seconds() > float64(threshold) {
					_, err := client.StopInstances(context.TODO(), &ec2.StopInstancesInput{
						InstanceIds: []string{id},
					})
					if err != nil {
						return "", err
					}
					msg = fmt.Sprintf("InstanceID: %v has been shut down!", *ins.InstanceId)
				}
			}
			break
		}
	}
	fmt.Println(msg)
	return msg, nil
}
func HandleRequest() ([]string, error) {
	var nextToken *string
	snsMessage := make(map[string]string)
	for {
		result, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
			Filters: []types.Filter{
				{
					Name: aws.String("instance-state-name"),
					Values: []string{
						"running",
					},
				},
				{
					Name: aws.String("tag-key"),
					Values: []string{
						"AllowStop",
					},
				},
			},
			MaxResults: aws.Int32(maxResults),
			NextToken:  nextToken,
		})
		if err != nil {
			return []string{}, err
		}

		for _, r := range result.Reservations {
			for _, ins := range r.Instances {
				msg, err := checkInstance(ins)
				if err != nil {
					fmt.Println(err)
					return []string{}, err
				}
				if msg != "" {
					snsMessage[*ins.InstanceId] = msg
				}
			}
		}
		if result.NextToken == nil {
			break
		}
		nextToken = result.NextToken
	}

	if len(snsMessage) > 0 {
		var out strings.Builder
		for _, v := range snsMessage {
			out.WriteString(fmt.Sprintf("Message: %v\n", v))
		}
		_, err := snsClient.Publish(context.TODO(), &sns.PublishInput{
			Message:  aws.String(out.String()),
			TopicArn: aws.String(topicArn),
			Subject:  aws.String("EC2 Instance Shutdown Warning"),
		})
		if err != nil {
			fmt.Println(err)
			return []string{}, err
		}
	}
	return []string{}, nil
}

func main() {
	lambda.Start(HandleRequest)
	//_, _ = HandleRequest()
}
