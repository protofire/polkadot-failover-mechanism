package utils

// This file contains all the supplementary functions that are required to query EC2's Elastic Block Storage API

import (
	"fmt"
	"testing"

	saws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/require"
)

// getVolumeDescribe This function list all prefixed volumes that does not attached to any instance
func getVolumeDescribe(t *testing.T, region string, tag string, value string) []*ec2.Volume {

	ses, err := session.NewSession(&saws.Config{
		Region: saws.String(region),
	})
	require.NoError(t, err)
	svc := ec2.New(ses)
	input := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: saws.String("status"),
				Values: []*string{
					saws.String("creating"),
					saws.String("available"),
					saws.String("deleting"),
					saws.String("error"),
				},
			},
			{
				Name: saws.String("tag:prefix"),
				Values: []*string{
					saws.String(value),
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

// VolumesCheck checks cvolumes
func VolumesCheck(t *testing.T, awsRegions []string, prefix string) bool {

	count := 0
	// Go through each region. Select unattached labeled disks. If no disks found, then the test passes successfully
	for _, region := range awsRegions {

		check := getVolumeDescribe(t, region, "prefix", prefix)

		if len(check) == 0 {
			t.Log("No unnatached disks were found in region " + region)
			continue
		} else {
			t.Error("Unattached disks were found in region " + region)
			count = count + 1
		}

	}

	return count == 0

}
