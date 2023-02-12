package resources

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateJob(options ...CreateJobOption) *batchv1.Job {
	jopts := &JobOptions{
		Name:            "defaultjobname",
		ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
		Labels:          make(map[string]string),
		OwnerReferences: make([]metav1.OwnerReference, 5),
		Namespace:       "default",
		Volumes:         make([]corev1.Volume, 5),
		Env:             make([]corev1.EnvVar, 10),
	}

	for _, o := range options {
		o.apply(jopts)
	}

	return CreateJobWithOptions(jopts)
}

func CreateJobWithOptions(jopts *JobOptions) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			// Concatenate the trainInKube name and the string "_build_model" to create the job name
			Name:            jopts.Name,
			Namespace:       jopts.Namespace,
			Labels:          jopts.Labels,
			OwnerReferences: jopts.OwnerReferences,
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
						Name:            jopts.Name,
						Image:           jopts.Image,
						ImagePullPolicy: jopts.ImagePullPolicy,
						VolumeMounts:    jopts.VolumeMounts,
						Env:             jopts.Env,
					},
				},
				Volumes:       jopts.Volumes,
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
	}
}

func CreateConfigMap(options ...CreateConfigMapOption) *corev1.ConfigMap {
	cmopts := &ConfigMapOptions{
		Name:            "defaultcmname",
		Data:            make(map[string]string),
		Namespace:       "default",
		OwnerReferences: make([]metav1.OwnerReference, 5),
	}

	for _, o := range options {
		o.apply(cmopts)
	}

	return CreateConfigMapWithOptions(cmopts)
}

func CreateConfigMapWithOptions(cmopts *ConfigMapOptions) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cmopts.Name,
			Namespace:       cmopts.Namespace,
			Labels:          make(map[string]string),
			OwnerReferences: cmopts.OwnerReferences,
		},
		Data: cmopts.Data,
	}
}
