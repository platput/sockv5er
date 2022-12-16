package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
)

type AWSHelper struct {
	client          *ec2.Client
	region          string
	keyPairName     string
	securityGroupID string
	defaultVPCID    string
	ec2InstanceId   string
	keyPairKey      string
}

func (helper *AWSHelper) InitializeAWS(s *Settings) error {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s.AccessKeyId, s.SecretKey, "")))
	if err != nil {
		return err
	} else {
		helper.client = ec2.NewFromConfig(cfg)
	}
	return nil
}

func (helper *AWSHelper) GetRegions() []types.Region {
	regions, err := helper.client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil
	}
	return regions.Regions
}

func (helper *AWSHelper) SetRegion(region string, s *Settings) error {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s.AccessKeyId, s.SecretKey, "")),
		config.WithRegion(region))
	if err != nil {
		return err
	}
	helper.client = ec2.NewFromConfig(cfg)
	return nil
}

func (helper *AWSHelper) GetDefaultVPC() error {
	isDefaultFilter := types.Filter{
		Name:   aws.String("is-default"),
		Values: []string{"true"},
	}
	filters := []types.Filter{isDefaultFilter}
	vpcInput := &ec2.DescribeVpcsInput{
		Filters: filters,
	}
	vpcs, err := helper.client.DescribeVpcs(context.TODO(), vpcInput)
	if err != nil {
		return err
	} else {
		if len(vpcs.Vpcs) > 0 {
			helper.defaultVPCID = *vpcs.Vpcs[0].VpcId
			return nil
		}
	}
	errorMessage := fmt.Sprintf("Unknown error in getting the default VPC for the region %s", helper.region)
	return errors.New(errorMessage)
}

func (helper *AWSHelper) CreateEC2Instance() error {
	userdata := fmt.Sprintf("#!/bin/bash\nshutdown +20")
	instanceInput := &ec2.RunInstancesInput{
		ImageId:                           aws.String("resolve:ssm:/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"),
		InstanceInitiatedShutdownBehavior: "terminate",
		InstanceType:                      "t2.micro",
		KeyName:                           aws.String(helper.keyPairName),
		SecurityGroupIds:                  []string{helper.securityGroupID},
		UserData:                          aws.String(userdata),
	}
	instance, err := helper.client.RunInstances(context.TODO(), instanceInput)
	if err != nil {
		return err
	}
	helper.ec2InstanceId = *instance.Instances[0].InstanceId
	return nil
}

//func (helper *AWSHelper) CheckIfInstanceIsActive(instanceId string) (bool, error) {
//	return false, nil
//}

func (helper *AWSHelper) CreateSecurityGroup() error {
	groupName := fmt.Sprintf("sockv5er-sg-group-%s", helper.region)
	description := fmt.Sprintf("Security group created by sockv5er for the region %s with just ssh enabled.", helper.region)
	sgInput := &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(groupName),
		Description: aws.String(description),
		VpcId:       aws.String(helper.defaultVPCID),
	}
	group, err := helper.client.CreateSecurityGroup(context.TODO(), sgInput)
	if err != nil {
		return err
	}
	helper.securityGroupID = *group.GroupId
	return nil
}

func (helper *AWSHelper) CreateKeyPair() error {
	keyName := fmt.Sprintf("sockv5er-keypair-region-%s", helper.region)
	keypairInput := &ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	}
	keypair, err := helper.client.CreateKeyPair(context.TODO(), keypairInput)
	if err != nil {
		return err
	}
	helper.keyPairName = *keypair.KeyPairId
	helper.keyPairKey = *keypair.KeyMaterial
	return nil
}

func (helper *AWSHelper) DeleteKeyPair() error {
	keypairInput := &ec2.DeleteKeyPairInput{KeyName: aws.String(helper.keyPairName)}
	_, err := helper.client.DeleteKeyPair(context.TODO(), keypairInput)
	if err != nil {
		return err
	}
	return nil
}

func (helper *AWSHelper) DeleteSecurityGroup() error {
	sgInput := &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(helper.securityGroupID),
	}
	_, err := helper.client.DeleteSecurityGroup(context.TODO(), sgInput)
	if err != nil {
		return err
	}
	return nil
}

func (helper *AWSHelper) TerminateEC2Instance() error {
	instanceInput := &ec2.TerminateInstancesInput{
		InstanceIds: []string{helper.ec2InstanceId},
	}
	_, err := helper.client.TerminateInstances(context.TODO(), instanceInput)
	if err != nil {
		return err
	}
	return nil
}
