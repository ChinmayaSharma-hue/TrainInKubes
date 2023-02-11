package controller

type eventType string

const (
	addTrainInKube eventType = "addTrainInKube"
	addConfigMap   eventType = "addConfigMap"
)

type event struct {
	eventType      eventType
	oldObj, newObj interface{}
}
