package aws

// This file contains all the supplementary functions that are required to query EC2's Elastic Block Storage API

import (
	"testing"

	saws "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/require"
)

// getVolumeDescribe This function list all prefixed volumes that does not attached to any instance
func getVolumeDescribe(t *testing.T, region string, value string) ([]*ec2.Volume, error) {

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
		if awsErr, ok := err.(awserr.Error); ok {
			err = awsErr
		}
		return nil, err
	}
	return result.Volumes, err
}

// VolumesCheck checks volumes
func VolumesCheck(t *testing.T, awsRegions []string, prefix string) bool {

	count := 0
	// Go through each region. Select unattached labeled disks. If no disks found, then the test passes successfully
	for _, region := range awsRegions {

		volumes, err := getVolumeDescribe(t, region, prefix)

		if err != nil {
			t.Error(err)
		}

		if len(volumes) == 0 {
			t.Log("No unattached disks were found in region " + region)
			continue
		} else {
			t.Error("Unattached disks were found in region " + region)
			count++
		}

	}

	return count == 0

}
