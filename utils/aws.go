package utils

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type AWSHelper struct {
	ec2Client *ec2.Client
	cfg       config.Config
}

func (helper *AWSHelper) InitializeAWS(s *Settings) error {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s.AccessKeyId, s.SecretKey, "")))
	if err != nil {
		return err
	} else {
		helper.cfg = cfg
		helper.ec2Client = ec2.NewFromConfig(cfg)
	}
	return nil
}

func (helper *AWSHelper) GetRegions() []types.Region {
	regions, err := helper.ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil
	}
	return regions.Regions
}
