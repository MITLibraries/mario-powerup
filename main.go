package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"os"
	"strings"
)

var awssession *session.Session

func init() {
	region, ok := os.LookupEnv("AWS_DEFAULT_REGION")
	if !ok {
		region = "us-east-1"
	}
	awssession = session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
}

func Handler(s3Event events.S3Event) {
	event := s3Event.Records[0]
	filename := fmt.Sprintf("s3://%s/%s", event.S3.Bucket.Name, event.S3.Object.Key)
	runtask(filename)
}

func main() {
	lambda.Start(Handler)
}

func runtask(filename string) {
	securitygroup := aws.String(os.Getenv("ECS_SECURITY_GROUP"))
	cluster := aws.String(os.Getenv("ECS_CLUSTER"))
	family := aws.String(os.Getenv("ECS_FAMILY"))
	esurl := aws.String(os.Getenv("ES_URL"))
	subnets_s := strings.Split(",", os.Getenv("ECS_SUBNETS"))

	var subnets []*string
	for _, s := range subnets_s {
		subnets = append(subnets, aws.String(s))
	}

	vpc := &ecs.AwsVpcConfiguration{
		SecurityGroups: []*string{securitygroup},
		Subnets:        subnets,
	}
	override := &ecs.ContainerOverride{
		Command: []*string{
			aws.String("--v4"),
			aws.String("--url"),
			esurl,
			aws.String("ingest"),
			&filename,
		},
	}
	input := &ecs.RunTaskInput{
		Cluster:              cluster,
		Count:                aws.Int64(1),
		LaunchType:           aws.String("FARGATE"),
		NetworkConfiguration: &ecs.NetworkConfiguration{AwsvpcConfiguration: vpc},
		Overrides:            &ecs.TaskOverride{ContainerOverrides: []*ecs.ContainerOverride{override}},
		TaskDefinition:       family,
	}
	svc := ecs.New(awssession)
	result, err := svc.RunTask(input)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(result)
}
