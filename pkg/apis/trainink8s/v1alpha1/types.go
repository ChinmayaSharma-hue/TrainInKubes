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
	ModelImage               string `json:"modelimage, omitempty"`
	ModelImagePullPolicy     string `json:"modelimagepullpolicy, omitempty`
	TrainingImage            string `json:"trainingimage, omitempty"`
	TrainingImagePullPolicy  string `json:"trainingimagepullpolicy, omitempty"`
	Epochs                   int    `json:"epochs, omitempty"`
	PreprocessedDataLocation string `json:"preprocesseddatalocation, omitempty"`
	SplitDatasetLocation     string `json:"splitdatasetlocation, omitempty"`
	ModelsLocation           string `json:"modelslocation, omitempty"`
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
