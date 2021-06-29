package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
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
	subnets_s := strings.Split(os.Getenv("ECS_SUBNETS"), ",")

	var subnets []*string
	for _, s := range subnets_s {
		subnets = append(subnets, aws.String(s))
	}

	vpc := &ecs.AwsVpcConfiguration{
		SecurityGroups: []*string{securitygroup},
		Subnets:        subnets,
	}

	// This block relies on certain file naming conventions to work. Daily
	// updates to Alma have the string UPDATE in the filename. If that
	// string is present we will add the records to the current production
	// Alma index instead of creating a new index.
	override := &ecs.ContainerOverride{
		Name: aws.String("dip"),
	}
	if strings.Contains(filename, "UPDATE") {
		log.Printf("Alma update file detected: %s", filename)
		command := []*string{
			aws.String("--url"),
			esurl,
			aws.String("ingest"),
			aws.String("--source"),
			aws.String("alma"),
			&filename,
		}
		override.SetCommand(command)
	} else {
		log.Printf("Alma full export detected: %s", filename)
		command := []*string{
			aws.String("--url"),
			esurl,
			aws.String("ingest"),
			aws.String("--source"),
			aws.String("alma"),
			aws.String("--new"),
			aws.String("--auto"),
			&filename,
		}
		override.SetCommand(command)
	}
	taskoverride := &ecs.TaskOverride{
		ContainerOverrides: []*ecs.ContainerOverride{override},
	}
	input := &ecs.RunTaskInput{
		Cluster:              cluster,
		Count:                aws.Int64(1),
		LaunchType:           aws.String("FARGATE"),
		NetworkConfiguration: &ecs.NetworkConfiguration{AwsvpcConfiguration: vpc},
		Overrides:            taskoverride,
		TaskDefinition:       family,
	}
	svc := ecs.New(awssession)
	result, err := svc.RunTask(input)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(result)
}
