package resources

import corev1 "k8s.io/api/core/v1"

type CreateJobOption interface {
	apply(*JobOptions) error
}

type createJobOptionAdapter func(*JobOptions) error

func (c createJobOptionAdapter) apply(j *JobOptions) error {
	return c(j)
}

func createJobWithName(name string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.Name = name
		return nil
	})
}

func createJobWithImage(imageName string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.Image = imageName
		return nil
	})
}

func createJobWithImagePullPolicy(policy string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.ImagePullPolicy = corev1.PullPolicy(policy)
		return nil
	})
}

func createJobWithLabels(labels map[string]string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.Labels = labels
		return nil
	})
}

func createJobInNamespace(namespace string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		j.Namespace = namespace
		return nil
	})
}

func createJobWithVolume(volume corev1.Volume) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		append(j.Volumes, volume)
		return nil
	})
}

func createJobWithVolumeMounts(volumeMount corev1.VolumeMount) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		append(j.VolumeMounts, volumeMount)
		return nil
	})
}

func createJobWithEnv(envVariables map[string]string) CreateJobOption {
	return createJobOptionAdapter(func(j *JobOptions) error {
		for key, val := range envVariables {
			envVar := corev1.EnvVar{
				Name:  key,
				Value: val,
			}
			append(j.Env, envVar)
		}
		return nil
	})
}

func createHostPathVolume(name string, path string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: path,
			},
		},
	}
}

func createVolumeMount(name string, mountPath string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      name,
		MountPath: mountPath,
	}
}
