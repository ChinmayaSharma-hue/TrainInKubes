package resources

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobOptions struct {
	Name            string
	Image           string
	ImagePullPolicy corev1.PullPolicy
	Labels          map[string]string
	OwnerReferences []metav1.OwnerReference
	Namespace       string
	Volumes         []corev1.Volume
	VolumeMounts    []corev1.VolumeMount
	Env             []corev1.EnvVar
}

type ConfigMapOptions struct {
	Name            string
	Data            map[string]string
	Namespace       string
	OwnerReferences []metav1.OwnerReference
}
