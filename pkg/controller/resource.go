package controller

import (
	traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"

	batchv1 "k8s.io/api/batch/v1"
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

func createJob(trainInKube *traininkubev1alpha1.TrainInKube, configmap *corev1.ConfigMap, namespace string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			// Concatenate the trainInKube name and the string "_build_model" to create the job name
			Name:      trainInKube.ObjectMeta.Name + "buildmodel",
			Namespace: namespace,
			Labels:    make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(trainInKube, traininkubev1alpha1.SchemeGroupVersion.WithKind("TrainInKube")),
			},
		},
		Spec: createJobSpec(trainInKube, configmap, namespace),
	}
}

// Maybe experiment with interfaces and mutable behaviors observed in Kind here to change behavior based on option like
// create job spec with host volume or with s3 volume or other cloud storage volumes
func createJobSpec(trainInKube *traininkubev1alpha1.TrainInKube, configmap *corev1.ConfigMap, namespace string) batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: trainInKube.Name + "-",
				Namespace:    namespace,
				Labels:       make(map[string]string),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            trainInKube.Name,
						Image:           "trainjob:latest",
						ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      trainInKube.Name + "volume",
								MountPath: "/data",
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "MODEL_STORAGE_LOCATION",
								Value: "/data",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: trainInKube.Name + "volume",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: configmap.Data["modelsLocation"],
							},
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
	}
}
