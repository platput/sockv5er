package utils

type CloudProvider interface {
	Initialize(s *Settings) error
	GetRegions(s *Settings) []map[string]string
	CreateResources(region string, s *Settings, tracker *ResourceTracker) error
	DeleteResources(region string, s *Settings, tracker *ResourceTracker) error
	PrepareResourcesForDeletion(resources map[string]string)
	UpdateTracker(resources map[string]string, op TrackingOp, tracker *ResourceTracker)
	GetHostIP() string
	GetPrivateKey() []byte
}

type TrackingOp int

const (
	Add TrackingOp = iota
	Remove
)
