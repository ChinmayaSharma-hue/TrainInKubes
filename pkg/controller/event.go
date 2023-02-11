package controller

import traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"

type eventType string

const (
	addTrainInKube eventType = "addTrainInKube"
	addConfigMap   eventType = "addConfigMap"
	addBuildModel  eventType = "addBuildModel"
)

type event struct {
	eventType      eventType
	newObj         interface{}
	customResource *traininkubev1alpha1.TrainInKube
}
