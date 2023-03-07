package utils

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type ResourceTracker struct {
	filepath  string
	resources *SockV5erResources
}

type SockV5erResources struct {
	Version      string        `yaml:"version"`
	Generator    string        `yaml:"generator"`
	AWSResources []AWSResource `yaml:"awsResources"`
}

type AWSResource struct {
	Region          string `yaml:"region"`
	InstanceId      string `yaml:"instanceId"`
	SecurityGroupId string `yaml:"securityGroupId"`
	KeyPairId       string `yaml:"keyPairId"`
}

func GetNewTracker(trackerFilepath string) *ResourceTracker {
	sockv5Resources := SockV5erResources{
		Version:      "1.0",
		Generator:    "SockV5er",
		AWSResources: make([]AWSResource, 0),
	}
	tracker := &ResourceTracker{
		filepath:  trackerFilepath,
		resources: &sockv5Resources,
	}
	return tracker
}

func (rt *ResourceTracker) WriteResourcesFile() error {
	yamlContent, err := yaml.Marshal(&rt.resources)
	if err != nil {
		return err
	}
	err = WriteFileContent(rt.filepath, yamlContent)
	if err != nil {
		return err
	}
	return nil
}

func (rt *ResourceTracker) ReadResourcesFile() error {
	fmt.Println("rt.filepath", rt)
	content, err := ReadFileContent(rt.filepath)
	if err != nil {
		return err
	}
	yamlContent := &SockV5erResources{}
	err = yaml.Unmarshal(content, yamlContent)
	if err != nil {
		return err
	}
	rt.resources = yamlContent
	return nil
}

func (rt *ResourceTracker) AddAWSResource(resource *AWSResource) {
	rt.resources.AWSResources = append(rt.resources.AWSResources, *resource)
}

func (rt *ResourceTracker) RemoveAWSResource(resource *AWSResource) {
	for i := range rt.resources.AWSResources {
		sliceLength := len(rt.resources.AWSResources)
		if rt.resources.AWSResources[i].InstanceId == resource.InstanceId {
			if sliceLength <= 1 {
				rt.resources.AWSResources = []AWSResource{}
				return
			} else {
				rt.resources.AWSResources[i] = rt.resources.AWSResources[sliceLength-1]
				rt.resources.AWSResources[sliceLength-1] = AWSResource{}
				rt.resources.AWSResources = rt.resources.AWSResources[:sliceLength-1]
				return
			}
		}
	}
}

func (rt *ResourceTracker) GetResources() *[]AWSResource {
	return &rt.resources.AWSResources
}

func ToMap(a *AWSResource) map[string]string {
	resourceMap := make(map[string]string)
	resourceMap["region"] = a.Region
	resourceMap["instanceId"] = a.InstanceId
	resourceMap["securityGroupId"] = a.SecurityGroupId
	resourceMap["keyPairId"] = a.KeyPairId
	return resourceMap
}

func FromMap(resourceMap map[string]string) *AWSResource {
	return &AWSResource{
		Region:          resourceMap["region"],
		InstanceId:      resourceMap["instanceId"],
		SecurityGroupId: resourceMap["securityGroupId"],
		KeyPairId:       resourceMap["keyPairId"],
	}
}

func FromAWSRepository(a *AWSRepository) *AWSResource {
	return &AWSResource{
		Region:          a.Region,
		InstanceId:      a.Ec2InstanceId,
		SecurityGroupId: a.SecurityGroupID,
		KeyPairId:       a.KeyPairId,
	}
}
