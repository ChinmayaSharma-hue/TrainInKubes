package controller

type eventType string

const (
	addTrainInKube eventType = "addTrainInKube"
	addConfigMap   eventType = "addConfigMap"
	addBuildModel  eventType = "addBuildModel"
)

type event struct {
	eventType      eventType
	oldObj, newObj interface{}
}
