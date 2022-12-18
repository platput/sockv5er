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
	log "github.com/sirupsen/logrus"
)

type AWSHelper struct {
	Client          *ec2.Client
	Region          string
	KeyPairName     string
	SecurityGroupID string
	defaultVPCID    string
	Ec2InstanceId   string
	InstanceEP      string
	KeyPairKey      string
	Settings        *Settings
}

func (helper *AWSHelper) InitializeAWS(s *Settings) error {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s.AccessKeyId, s.SecretKey, "")))
	if err != nil {
		return err
	} else {
		helper.Client = ec2.NewFromConfig(cfg)
	}
	return nil
}

func (helper *AWSHelper) GetRegions() []types.Region {
	regions, err := helper.Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil
	}
	return regions.Regions
}

func (helper *AWSHelper) CreateAWSResources(region string, s *Settings) error {
	err := helper.SetRegion(region, s)
	if err != nil {
		return err
	}
	log.Infof("Region set as: `%s`.\n", helper.Region)
	err = helper.CreateSecurityGroup()
	if err != nil {
		return err
	}
	log.Infof("Security Group with ID: `%s` created.\n", helper.SecurityGroupID)
	err = helper.CreateKeyPair()
	if err != nil {
		return err
	}
	log.Infof("Key Pair: `%s` created.\n", helper.KeyPairName)
	err = helper.CreateEC2Instance()
	if err != nil {
		return err
	}
	log.Infof("Instance with id: `%s` created.\n", helper.Ec2InstanceId)
	return nil
}

func (helper *AWSHelper) DeleteAWSResources(region string, s *Settings) error {
	err := helper.SetRegion(region, s)
	if err != nil {
		return err
	}
	err = helper.TerminateEC2Instance(helper.Ec2InstanceId)
	if err != nil {
		return err
	}
	log.Infof("EC2 instance with ID: `%s` terminated.\n", helper.Ec2InstanceId)
	err = helper.DeleteKeyPair(helper.KeyPairName)
	if err != nil {
		return err
	}
	log.Infof("Key Pair: `%s` deleted.\n", helper.KeyPairName)
	err = helper.DeleteSecurityGroup(helper.SecurityGroupID)
	if err != nil {
		return err
	}
	log.Infof("Security group with id: `%s` deleted.\n", helper.SecurityGroupID)
	return nil
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
	helper.Client = ec2.NewFromConfig(cfg)
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
	vpcs, err := helper.Client.DescribeVpcs(context.TODO(), vpcInput)
	if err != nil {
		return err
	} else {
		if len(vpcs.Vpcs) > 0 {
			helper.defaultVPCID = *vpcs.Vpcs[0].VpcId
			return nil
		}
	}
	errorMessage := fmt.Sprintf("Unknown error in getting the default VPC for the Region %s\n", helper.Region)
	return errors.New(errorMessage)
}

func (helper *AWSHelper) CreateEC2Instance() error {
	userdata := fmt.Sprintf("#!/bin/bash\nshutdown +20")
	instanceInput := &ec2.RunInstancesInput{
		ImageId:                           aws.String("resolve:ssm:/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"),
		InstanceInitiatedShutdownBehavior: "terminate",
		InstanceType:                      "t2.micro",
		KeyName:                           aws.String(helper.KeyPairName),
		SecurityGroupIds:                  []string{helper.SecurityGroupID},
		UserData:                          aws.String(userdata),
	}
	instance, err := helper.Client.RunInstances(context.TODO(), instanceInput)
	if err != nil {
		return err
	}
	helper.Ec2InstanceId = *instance.Instances[0].InstanceId
	return nil
}

//func (helper *AWSHelper) CheckIfInstanceIsActive(instanceId string) (bool, error) {
//	return false, nil
//}

func (helper *AWSHelper) CreateSecurityGroup() error {
	groupName := fmt.Sprintf("sockv5er-sg-group-%s", helper.Region)
	description := fmt.Sprintf("Security group created by sockv5er for the Region %s with just ssh enabled.", helper.Region)
	sgInput := &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(groupName),
		Description: aws.String(description),
		VpcId:       aws.String(helper.defaultVPCID),
	}
	group, err := helper.Client.CreateSecurityGroup(context.TODO(), sgInput)
	if err != nil {
		return err
	}
	helper.SecurityGroupID = *group.GroupId
	return nil
}

func (helper *AWSHelper) CreateKeyPair() error {
	keyName := fmt.Sprintf("sockv5er-keypair-Region-%s", helper.Region)
	keypairInput := &ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	}
	keypair, err := helper.Client.CreateKeyPair(context.TODO(), keypairInput)
	if err != nil {
		return err
	}
	helper.KeyPairName = *keypair.KeyPairId
	helper.KeyPairKey = *keypair.KeyMaterial
	return nil
}

func (helper *AWSHelper) DeleteKeyPair(keyPairName string) error {
	keypairInput := &ec2.DeleteKeyPairInput{KeyName: aws.String(keyPairName)}
	_, err := helper.Client.DeleteKeyPair(context.TODO(), keypairInput)
	if err != nil {
		return err
	}
	return nil
}

func (helper *AWSHelper) DeleteSecurityGroup(securityGroupId string) error {
	sgInput := &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(securityGroupId),
	}
	_, err := helper.Client.DeleteSecurityGroup(context.TODO(), sgInput)
	if err != nil {
		return err
	}
	return nil
}

func (helper *AWSHelper) TerminateEC2Instance(instanceId string) error {
	instanceInput := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceId},
	}
	_, err := helper.Client.TerminateInstances(context.TODO(), instanceInput)
	if err != nil {
		return err
	}
	return nil
}
