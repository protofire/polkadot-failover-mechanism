package test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
)

func GetVolumeDescribe(t *testing.T, region string, tag string, value string) ([]*ec2.Volume) {

	svc := ec2.New(session.New(&aws.Config{
		Region: aws.String(region),
        }))
    input := &ec2.DescribeVolumesInput{
	   Filters: []*ec2.Filter{
			{
			  Name: aws.String("status"),
			  Values: []*string{
				aws.String("creating"),
				aws.String("available"),
				aws.String("deleting"),
				aws.String("error"),
			   },
			 },
			{   
			   Name: aws.String("tag:prefix"),
			   Values: []*string{
				 aws.String(prefix),
				 },
			 },
		 },
	   }
    result, err := svc.DescribeVolumes(input)
     if err != nil {
     if aerr, ok := err.(awserr.Error); ok {
	   switch aerr.Code() {
	   default:
	    fmt.Println(aerr.Error())
		}
	 } else {
	  // Print the error, cast err to awserr.Error to get the Code and
	  // Message from an error.
	  fmt.Println(err.Error())
	 }
	 //return
     }
   return result.Volumes
}