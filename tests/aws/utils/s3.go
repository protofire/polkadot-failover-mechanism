package utils

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// EnsureTFBucket ebsure TF bucket exists
func EnsureTFBucket(name, region string) (bool, error) {

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
				return false, fmt.Errorf("Bucket %q already exists and not belongs to you: %w", name, aerr)
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return false, nil
			default:
				return false, aerr
			}
		}
		return false, err
	}
	return true, nil
}

func deleteBucketObjects(client *s3.S3, name string) error {

	result, err := client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("Cannot read bucket %q objects: %w", name, err)
	}

	for _, obj := range result.Contents {
		_, err := client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(name),
			Key:    obj.Key,
		})
		if err != nil {
			return fmt.Errorf("Cannot delete key %q. Bucket %q: %w", *obj.Key, name, err)
		}
		log.Printf("Deleted key %q from bucket %q", *obj.Key, name)
	}
	return err

}

// DeleteTFBucket clear and deleted bucket
func DeleteTFBucket(name, region string) error {

	log.Printf("Deleting bucket %q...", name)

	ses := session.Must(session.NewSession())
	client := s3.New(ses, aws.NewConfig().WithRegion(region))

	if err := deleteBucketObjects(client, name); err != nil {
		return err
	}

	deleteInput := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}

	_, err := client.DeleteBucket(deleteInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				return fmt.Errorf("Bucket %q absent: %w", name, aerr)
			default:
				return aerr
			}
		}
		return err
	}

	log.Printf("Deleted bucket %q", name)

	return nil
}

// ClearTFBucket ebsure TF bucket exists
func ClearTFBucket(name, region string) error {

	log.Printf("Clearing bucket %q...", name)

	ses := session.Must(session.NewSession())
	client := s3.New(ses, aws.NewConfig().WithRegion(region))

	if err := deleteBucketObjects(client, name); err != nil {
		return err
	}

	return deleteBucketObjects(client, name)

}
