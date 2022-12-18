package utils

import (
	"errors"
	log "github.com/sirupsen/logrus"
)

type CloudHelper struct {
	awsHelper *AWSHelper
}

type Provider string

const (
	AWS Provider = "aws"
)

func (h *CloudHelper) DeleteResource(resource *Resource) error {
	if resource.ProviderName == "aws" {
		err := h.awsHelper.SetRegion(resource.Region, h.awsHelper.Settings)
		if err != nil {
			return err
		}
		// Doesn't matter if this fails or not as we are creating instance with userdata
		// set to shut down the system in 10 minutes
		_ = h.awsHelper.TerminateEC2Instance(resource.InstanceId)
		err = h.awsHelper.DeleteKeyPair(resource.KeyPairName)
		if err != nil {
			return err
		}
		err = h.awsHelper.DeleteSecurityGroup(resource.SecurityGroupId)
		if err != nil {
			return err
		}
	}
	log.Info("All related resources under the instance id %s have been deleted successfully.\n", resource.InstanceId)
	return nil
}

func (h *CloudHelper) CreateResource(provider Provider) (*Resource, error) {
	if provider == AWS {
		err := h.awsHelper.CreateKeyPair()
		if err != nil {
			log.Fatalf("Creating key pair failed with error: %s. Program can not continue.\n", err)
		}
		err = h.awsHelper.CreateSecurityGroup()
		if err != nil {
			log.Fatalf("Creating security group failed with error: %s. Program can not continue.\n", err)
		}
		err = h.awsHelper.CreateEC2Instance()
		if err != nil {
			log.Fatalf("Creating ec2 instance failed with error: %s. Program can not continue.\n", err)
		}
		return &Resource{
			Region:          h.awsHelper.Region,
			InstanceId:      h.awsHelper.Ec2InstanceId,
			SecurityGroupId: h.awsHelper.SecurityGroupID,
			KeyPairName:     h.awsHelper.KeyPairName,
			ProviderName:    string(provider),
		}, nil
	}
	return nil, errors.New("unsupported cloud provider, this shouldn't have happened")
}
