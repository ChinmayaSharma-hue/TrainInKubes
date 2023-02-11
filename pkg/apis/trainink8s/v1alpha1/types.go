package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TrainInKube struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TrainInKubeSpec `json:"spec"`

	Status TrainInKubeStatus `json:"status,omitempty"`
}

type TrainInKubeSpec struct {
	ModelImage               string `json:"modelImage, omitempty"`
	ModelImagePullPolicy     string `json:"modelImagepullpolicy, omitempty`
	Epochs                   int    `json:"epochs, omitempty"`
	BatchSize                int    `json:"batchSize, omitempty"`
	PreprocessedDataLocation string `json:"preprocessedDataLocation, omitempty"`
	SplitDatasetLocation     string `json:"splitDatasetLocation, omitempty"`
	ModelsLocation           string `json:"modelsLocation, omitempty"`
}

type TrainInKubeStatus struct {
	NumberOfJobs   int
	Phase          string
	Succeeded      int
	CompletionTime string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TrainInKubeList struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ListMeta `json:"metadata,omitempty"`

	Items []TrainInKube `json:"items"`
}
