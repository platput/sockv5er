package utils

import log "github.com/sirupsen/logrus"

type CloudHelper struct {
	awsHelper *AWSHelper
}

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
