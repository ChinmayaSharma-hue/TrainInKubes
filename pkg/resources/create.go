package resources

import (
	corev1 "k8s.io/api/core/v1"
)

func CreateJob(options ...CreateOption) err {
	jopts := &JobOptions{
		Name: "defaultjobname",
		ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
		Labels: make(map[string]string),
		Namespace: "default",
		Volumes: make([]corev1.Volume),
		Env: make([]corev1.EnvVar)
	}

	for _, o := range options {
		if err := o.apply(jopts); err != nil {
			return err
		}
	}

	return CreateJobWithOptions(jopts)
}

func CreateJobWithOptions(jopts *JobOptions) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			// Concatenate the trainInKube name and the string "_build_model" to create the job name
			Name:      jopts.Name,
			Namespace: jopts.Namespace,
			Labels:    jopts.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(trainInKube, traininkubev1alpha1.SchemeGroupVersion.WithKind("TrainInKube")),
			},
		},
		Spec: CreateJobSpecWithOptions(jopts),
	}
}

func CreateJobSpecWithOptions(jopts *JobOptions) batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: jopts.Name + "-",
				Namespace:    jopts.Namespace,
				Labels:       make(map[string]string),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  jopts.Name,
						Image: jopts.Image,
						ImagePullPolicy: jopts.ImagePullPolicy,
						VolumeMounts: jopts.VolumeMounts,
						Env: jopts.Env,
					},
				},
				Volumes: jopts.Volumes,
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
	}
}

func CreateConfigMap(options ...CreateOption) err {
	cmopts := &ConfigMapOptions{
		Name: "defaultcmname",
		Data: make(map[string]string),
		Namespace: "default",
	}

	return CreateConfigMapWithOptions(cmopts)
}

func CreateConfigMapWithOptions(cmopts *ConfigMapOptions) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmopts.Name,
			Namespace: cmopts.Namespace,
			Labels:    cmopts.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(trainInKube, traininkubev1alpha1.SchemeGroupVersion.WithKind("TrainInKube")),
			},
		},
		Data: cmopts.Data,
	}
}

//map[string]string{
// 	"epochs":                      string(trainInKube.Spec.Epochs),
// 	"batchSize":                   string(trainInKube.Spec.BatchSize),
// 	"numberOfSamples":             strconv.Itoa(trainInKube.Spec.NumberOfSamples),
// 	"preprocessedDatasetLocation": trainInKube.Spec.PreprocessedDataLocation,
// 	"splitDatasetLocation":        trainInKube.Spec.SplitDatasetLocation,
// 	"modelsLocation":              trainInKube.Spec.ModelsLocation,
// }