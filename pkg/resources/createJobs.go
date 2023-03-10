package resources

import (
	traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateJobOption interface {
	apply(*JobOptions) error
}

type createJobOptionAdapter func(*JobOptions) error

func (c createJobOptionAdapter) apply(j *JobOptions) error {
	return c(j)
}

func CreateJobWithName(name string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.Name = name
		return nil
	})
}

func CreateJobWithImage(imageName string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.Image = imageName
		return nil
	})
}

func CreateJobWithImagePullPolicy(policy string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.ImagePullPolicy = corev1.PullPolicy(policy)
		return nil
	})
}

func CreateJobWithLabels(labels map[string]string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.Labels = labels
		return nil
	})
}

func CreateJobInNamespace(namespace string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.Namespace = namespace
		return nil
	})
}

func CreateJobWithVolume(volume corev1.Volume) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.Volumes = append(j.Volumes, volume)
		return nil
	})
}

func CreateJobWithVolumeMounts(volumeMount corev1.VolumeMount) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.VolumeMounts = append(j.VolumeMounts, volumeMount)
		return nil
	})
}

func CreateJobWithEnv(envVariables map[string]string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		for key, val := range envVariables {
			envVar := corev1.EnvVar{
				Name:  key,
				Value: val,
			}
			j.Env = append(j.Env, envVar)
		}
		return nil
	})
}

func CreateJobWithOwnerReference(ownerReference metav1.OwnerReference) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.OwnerReferences = append(j.OwnerReferences, ownerReference)
		return nil
	})
}

func CreateOwnerReference(trainInKube *traininkubev1alpha1.TrainInKube) metav1.OwnerReference {
	return *metav1.NewControllerRef(trainInKube, traininkubev1alpha1.SchemeGroupVersion.WithKind("TrainInKube"))
}

func CreateHostPathVolume(name string, path string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: path,
			},
		},
	}
}

func CreateVolumeMount(name string, mountPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	}
}
