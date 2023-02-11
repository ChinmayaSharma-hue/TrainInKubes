package controller

import (
	traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createConfigMap(trainInKube *traininkubev1alpha1.TrainInKube, namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      trainInKube.ObjectMeta.Name,
			Namespace: namespace,
			Labels:    make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(trainInKube, traininkubev1alpha1.SchemeGroupVersion.WithKind("TrainInKube")),
			},
		},
		Data: map[string]string{
			"epochs":                      string(trainInKube.Spec.Epochs),
			"preprocessedDatasetLocation": trainInKube.Spec.PreprocessedDataLocation,
			"splitDatasetLocation":        trainInKube.Spec.SplitDatasetLocation,
			"modelsLocation":              trainInKube.Spec.ModelsLocation,
		},
	}
}
