package utils

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// EnsureTFBucket ebsure TF bucket exists
func EnsureTFBucket(name, region string) error {

	ses := session.Must(session.NewSession())
	client := s3.New(ses, aws.NewConfig().WithRegion(region))

	createInput := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}

	if len(region) > 0 && region != "us-east-1" {
		createInput.CreateBucketConfiguration = &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(region),
		}
	}
	_, err := client.CreateBucket(createInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				return fmt.Errorf("Bucket %q already exists and not belongs to you", name)
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return nil
			default:
				return aerr
			}
		}
		return err
	}
	return nil
}
