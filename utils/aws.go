package utils

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type AWSRepository struct {
	Client          *ec2.Client
	Region          string
	KeyPairId       string
	SecurityGroupID string
	defaultVPCID    string
	Ec2InstanceId   string
	InstanceIP      string
	KeyPairKey      string
}

func NewAWSProvider() CloudProvider {
	return &AWSRepository{}
}

func (repo *AWSRepository) UpdateTracker(resources map[string]string, op TrackingOp, tracker *ResourceTracker) {
	awsResource := FromMap(resources)
	if op == Add {
		tracker.AddAWSResource(awsResource)
	} else if op == Remove {
		tracker.RemoveAWSResource(awsResource)
	}
	err := tracker.WriteResourcesFile()
	if err != nil {
		log.Warnf("Updating resources tracker file failed with err: %s.", err)
	}
}

func (repo *AWSRepository) Initialize(s *Settings) error {
	client, err := getEC2Client(s)
	if err != nil {
		return err
	}
	repo.Client = client
	return nil
}

func getEC2Client(settings *Settings) (*ec2.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(settings.AccessKeyId, settings.SecretKey, "")))
	if err != nil {
		return nil, err
	} else {
		return ec2.NewFromConfig(cfg), nil
	}
}

func (repo *AWSRepository) GetRegions(s *Settings) []map[string]string {
	result, err := repo.Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return []map[string]string{}
	}
	regions := result.Regions
	gh := GeoHelper{Settings: s}
	cloudAgnosticRegions := make([]map[string]string, 0)
	for i := range regions {
		country, err := gh.FindCountry(*regions[i].Endpoint)
		if err != nil {
			log.Warnf("Finding country for endpoint %s failed with error %v\n", *regions[i].Endpoint, err)
		} else {
			country := gh.GetCountryShortName(country)
			countryToRegionMap := map[string]string{
				"country": country,
				"Region":  *regions[i].RegionName,
			}
			cloudAgnosticRegions = append(cloudAgnosticRegions, countryToRegionMap)
		}
	}
	return cloudAgnosticRegions
}

func (repo *AWSRepository) CreateResources(region string, s *Settings, tracker *ResourceTracker) error {
	err := repo.SetRegion(region, s)
	if err != nil {
		resource := ToMap(FromAWSRepository(repo))
		repo.UpdateTracker(resource, Add, tracker)
		return err
	}
	log.Infof("Region set as: `%s`.\n", repo.Region)
	err = repo.CreateSecurityGroup()
	if err != nil {
		resource := ToMap(FromAWSRepository(repo))
		repo.UpdateTracker(resource, Add, tracker)
		return err
	}
	log.Infof("Security Group with ID: `%s` created.\n", repo.SecurityGroupID)
	err = repo.CreateKeyPair()
	if err != nil {
		resource := ToMap(FromAWSRepository(repo))
		repo.UpdateTracker(resource, Add, tracker)
		return err
	}
	log.Infof("Key Pair: `%s` created.\n", repo.KeyPairId)
	instanceId, err := repo.CreateEC2Instance()
	if err != nil {
		resource := ToMap(FromAWSRepository(repo))
		repo.UpdateTracker(resource, Add, tracker)
		return err
	}
	repo.Ec2InstanceId = instanceId
	log.Infof("Instance with id: `%s` created.\n", repo.Ec2InstanceId)
	resource := ToMap(FromAWSRepository(repo))
	repo.UpdateTracker(resource, Add, tracker)
	repo.WaitUntilInstanceIsActive(repo.Ec2InstanceId)
	repo.InstanceIP, err = repo.getPublicIPAddress(instanceId)
	if err != nil {
		log.Fatalf("Unknown error occured while trying to create EC2 Instance. Details: %s", err)
	}
	return nil
}

func (repo *AWSRepository) DeleteResources(region string, s *Settings, tracker *ResourceTracker) error {
	err := repo.SetRegion(region, s)
	if err != nil {
		return err
	}
	if repo.Ec2InstanceId != "" {
		err = repo.TerminateEC2Instance(repo.Ec2InstanceId)
		if err != nil {
			log.Warnf("EC2 instance termination failed with error: %s", err)
			return err
		}
		log.Infof("EC2 instance with ID: `%s` terminated.\n", repo.Ec2InstanceId)
		repo.WaitUntilInstanceIsTerminated(repo.Ec2InstanceId)
	}
	if repo.KeyPairId != "" {
		err = repo.DeleteKeyPair(repo.KeyPairId)
		if err != nil {
			log.Warnf("EC2 key pair deletion failed with error: %s", err)
			return err
		}
		log.Infof("Key Pair: `%s` deleted.\n", repo.KeyPairId)
	}
	if repo.SecurityGroupID != "" {
		err = repo.DeleteSecurityGroup(repo.SecurityGroupID)
		if err != nil {
			log.Warnf("EC2 security group deletion failed with error: %s", err)
			return err
		}
		log.Infof("Security group with id: `%s` deleted.\n", repo.SecurityGroupID)
	}
	resource := ToMap(FromAWSRepository(repo))
	repo.UpdateTracker(resource, Remove, tracker)
	return nil
}

func (repo *AWSRepository) SetRegion(region string, s *Settings) error {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s.AccessKeyId, s.SecretKey, "")),
		config.WithRegion(region))
	if err != nil {
		return err
	}
	repo.Region = region
	repo.Client = ec2.NewFromConfig(cfg)
	return nil
}

func (repo *AWSRepository) PrepareResourcesForDeletion(resource map[string]string) {
	res := FromMap(resource)
	repo.Region = res.Region
	repo.Ec2InstanceId = res.InstanceId
	repo.KeyPairId = res.KeyPairId
	repo.SecurityGroupID = res.SecurityGroupId
}

func (repo *AWSRepository) GetDefaultVPC() error {
	isDefaultFilter := types.Filter{
		Name:   aws.String("is-default"),
		Values: []string{"true"},
	}
	filters := []types.Filter{isDefaultFilter}
	vpcInput := &ec2.DescribeVpcsInput{
		Filters: filters,
	}
	vpcs, err := repo.Client.DescribeVpcs(context.TODO(), vpcInput)
	if err != nil {
		return err
	} else {
		if len(vpcs.Vpcs) > 0 {
			repo.defaultVPCID = *vpcs.Vpcs[0].VpcId
			return nil
		}
	}
	errorMessage := fmt.Sprintf("Unknown error in getting the default VPC for the Region %s\n", repo.Region)
	return errors.New(errorMessage)
}

func (repo *AWSRepository) CreateEC2Instance() (string, error) {
	userdata := fmt.Sprintf("#!/bin/bash\nshutdown +20")
	encodedUserdata := base64.StdEncoding.EncodeToString([]byte(userdata))
	var maxCount int32 = 1
	var minCount int32 = 1
	var keyName = fmt.Sprintf("sockv5er-keypair-%s", repo.Region)
	instanceInput := &ec2.RunInstancesInput{
		ImageId:                           aws.String("resolve:ssm:/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"),
		InstanceInitiatedShutdownBehavior: "terminate",
		InstanceType:                      "t2.micro",
		KeyName:                           &keyName,
		SecurityGroupIds:                  []string{repo.SecurityGroupID},
		UserData:                          aws.String(encodedUserdata),
		MaxCount:                          &maxCount,
		MinCount:                          &minCount,
	}
	instance, err := repo.Client.RunInstances(context.TODO(), instanceInput)
	if err != nil {
		return "", err
	}
	return *instance.Instances[0].InstanceId, nil
}

func (repo *AWSRepository) CreateSecurityGroup() error {
	groupName := fmt.Sprintf("sockv5er-sg-group-%s", repo.Region)
	description := fmt.Sprintf("Security group created by sockv5er for the Region %s with just ssh enabled.", repo.Region)
	sgInput := &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(groupName),
		Description: aws.String(description),
		VpcId:       aws.String(repo.defaultVPCID),
	}
	group, err := repo.Client.CreateSecurityGroup(context.TODO(), sgInput)
	if err != nil {
		return err
	}
	repo.SecurityGroupID = *group.GroupId
	sgIngressInput := &ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:     aws.String("0.0.0.0/0"),
		FromPort:   aws.Int32(22),
		GroupId:    group.GroupId,
		IpProtocol: aws.String("tcp"),
		ToPort:     aws.Int32(22),
	}
	_, err = repo.Client.AuthorizeSecurityGroupIngress(context.TODO(), sgIngressInput)
	if err != nil {
		log.Error("Error opening port 22 using security group:", err)
		return err
	}
	return nil
}

func (repo *AWSRepository) CreateKeyPair() error {
	keyName := fmt.Sprintf("sockv5er-keypair-%s", repo.Region)
	keypairInput := &ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	}
	keypair, err := repo.Client.CreateKeyPair(context.TODO(), keypairInput)
	if err != nil {
		return err
	}
	repo.KeyPairId = *keypair.KeyPairId
	repo.KeyPairKey = *keypair.KeyMaterial
	return nil
}

func (repo *AWSRepository) DeleteKeyPair(keyPairId string) error {
	keypairInput := &ec2.DeleteKeyPairInput{KeyPairId: aws.String(keyPairId)}
	_, err := repo.Client.DeleteKeyPair(context.TODO(), keypairInput)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
			return nil
		}
		return err
	}
	return nil
}

func (repo *AWSRepository) DeleteSecurityGroup(securityGroupId string) error {
	sgInput := &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(securityGroupId),
	}
	_, err := repo.Client.DeleteSecurityGroup(context.TODO(), sgInput)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidGroup.NotFound") {
			return nil
		}
		return err
	}
	return nil
}

func (repo *AWSRepository) TerminateEC2Instance(instanceId string) error {
	instanceInput := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceId},
	}
	_, err := repo.Client.TerminateInstances(context.TODO(), instanceInput)
	if err != nil {
		if err != nil {
			if strings.Contains(err.Error(), " InvalidInstanceID.NotFound") {
				return nil
			}
			return err
		}
		return err
	}
	return nil
}

type InstanceState string

const (
	Running    InstanceState = "running"
	Stopped                  = "stopped"
	Terminated               = "terminated"
)

func (repo *AWSRepository) CheckIfInstanceIsInState(instanceId string, state InstanceState) bool {
	// Create a "describe instances" input with the specified instance ID
	instanceInput := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	}
	// Use the EC2 client to describe the instances
	result, err := repo.Client.DescribeInstances(context.TODO(), instanceInput)
	if err != nil {
		return false
	}
	// Check if the instance is running
	if len(result.Reservations) > 0 && len(result.Reservations[0].Instances) > 0 {
		currentState := result.Reservations[0].Instances[0].State.Name
		if InstanceState(currentState) == state {
			return true
		}
	}
	return false
}

func (repo *AWSRepository) WaitUntilInstanceIsActive(instanceId string) bool {
	fn := func() bool {
		return repo.CheckIfInstanceIsInState(instanceId, Running)
	}
	status := callWithTimeout(5*time.Minute, fn)
	if status {
		log.Infoln("Instance is ready.... \nWaiting 10 seconds to finish up the public ip assignment ")
		time.Sleep(10 * time.Second)
		return true
	}
	return false
}

func (repo *AWSRepository) WaitUntilInstanceIsTerminated(instanceId string) bool {
	// Check if the instance exists
	status, err := repo.CheckIfInstanceExists(instanceId)
	if err != nil {
		log.Errorf("Unknown error while try to check if the instance exists: %s", err)
		return true
	}
	if status == false {
		return true
	}
	fn := func() bool {
		return repo.CheckIfInstanceIsInState(instanceId, Terminated)
	}
	status = callWithTimeout(5*time.Minute, fn)
	if status {
		log.Infoln("Instance has been terminated.... \nWaiting 10 seconds to finish up!")
		time.Sleep(10 * time.Second)
		return true
	}
	return false
}

func callWithTimeout(duration time.Duration, fn func() bool) bool {
	timeout := time.After(duration)
	done := make(chan bool)
	defer close(done)

	go func() {
		for {
			select {
			case <-timeout:
				done <- false
				return
			default:
				if fn() {
					done <- true
					return
				}
			}
		}
	}()
	return <-done
}

func (repo *AWSRepository) GetHostIP() string {
	return repo.InstanceIP
}

func (repo *AWSRepository) GetPrivateKey() []byte {
	return []byte(repo.KeyPairKey)
}

func (repo *AWSRepository) getPublicIPAddress(instanceID string) (string, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	resp, err := repo.Client.DescribeInstances(context.TODO(), input)
	if err != nil {
		return "", err
	}
	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found: %s", instanceID)
	}
	return *resp.Reservations[0].Instances[0].PublicIpAddress, nil
}

func (repo *AWSRepository) CheckIfInstanceExists(instanceID string) (bool, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}
	resp, err := repo.Client.DescribeInstances(context.TODO(), input)
	if err != nil {
		log.Errorf("Getting instance status failed with error: %s", err)
		return false, err
	}
	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return false, nil
	}
	return true, nil
}
