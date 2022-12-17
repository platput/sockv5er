package utils

import (
	"gopkg.in/yaml.v2"
)

type ResourcesTracker interface {
	WriteResourcesFile() error
	ReadResourcesFile() error
	AddResource(resource *Resource)
	RemoveResource(resource *Resource)
	GetResources() *[]Resource
}

type Resource struct {
	Region          string `yaml:"region"`
	InstanceId      string `yaml:"instanceId"`
	SecurityGroupId string `yaml:"securityGroupId"`
	KeyPairName     string `yaml:"keyPairName"`
	ProviderName    string `yaml:"providerName"`
}

type SockV5erResources struct {
	Version   string     `yaml:"version"`
	Generator string     `yaml:"generator"`
	Resources []Resource `yaml:"resources"`
}

type YAMLHelper struct {
	filepath  string
	resources *SockV5erResources
}

func (y *YAMLHelper) WriteResourcesFile() error {
	yamlContent, err := yaml.Marshal(&y.resources)
	if err != nil {
		return err
	}
	err = WriteFileContent(y.filepath, yamlContent)
	if err != nil {
		return err
	}
	return nil
}

func (y *YAMLHelper) ReadResourcesFile() error {
	content, err := ReadFileContent(y.filepath)
	if err != nil {
		return err
	}
	yamlContent := SockV5erResources{}
	err = yaml.Unmarshal(content, yamlContent)
	if err != nil {
		return err
	}
	y.resources = &yamlContent
	return nil
}

func (y *YAMLHelper) AddResource(resource *Resource) {
	y.resources.Resources = append(y.resources.Resources, *resource)
}

func (y *YAMLHelper) RemoveResource(resource *Resource) {
	for i := range y.resources.Resources {
		sliceLength := len(y.resources.Resources)
		if y.resources.Resources[i].InstanceId == resource.InstanceId {
			if sliceLength <= 1 {
				y.resources.Resources = []Resource{}
				return
			} else {
				y.resources.Resources[i] = y.resources.Resources[sliceLength-1]
				y.resources.Resources[sliceLength-1] = Resource{}
				y.resources.Resources = y.resources.Resources[:sliceLength-1]
				return
			}
		}
	}
}

func (y *YAMLHelper) GetResources() *[]Resource {
	return &y.resources.Resources
}
