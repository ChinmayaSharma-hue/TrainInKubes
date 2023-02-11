package controller

import (
	"context"
	"fmt"

	traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"
	"github.com/gotway/gotway/pkg/log"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"strconv"
)

type TrainOrchestrator struct {
	kubeClientSet kubernetes.Interface
	trainInKube   *traininkubev1alpha1.TrainInKube
	jobInformer   cache.SharedIndexInformer
	nodeInformer  cache.SharedIndexInformer

	namespace string

	logger log.Logger
}

func (t *TrainOrchestrator) Run(ctx context.Context, trainInKube *traininkubev1alpha1.TrainInKube) {
	t.logger.Infof("Starting the job orchestrator...")

	err := t.Orchestrate(ctx, trainInKube)
	if err != nil {
		t.logger.Errorf("Error while orchestrating the jobs: %v", err)
	}
}

func (t *TrainOrchestrator) Orchestrate(ctx context.Context, trainInKube *traininkubev1alpha1.TrainInKube) error {
	configmap, err := t.kubeClientSet.CoreV1().ConfigMaps(t.namespace).Get(ctx, trainInKube.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Error while getting the ConfigMap: %v", err)
	}

	// Create a job that divides the data between the jobs
	// Find a way to use the same function or something to create jobs that
	// creates different job objects based on the options passed to it
	job := createSplitJob(t.trainInKube, strconv.Itoa(5), configmap, t.namespace)

	exists, err := resourceExists(job, t.jobInformer.GetIndexer())
	if err != nil {
		return fmt.Errorf("Error while checking if the Job already exists: %v", err)
	}
	if exists {
		t.logger.Infof("Job already exists, skipping creation")
		return nil
	}

	created_job, err := t.kubeClientSet.BatchV1().Jobs(t.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Error while creating the Job: %v", err)
	}

	// Block the function till the job finishes execution
	err = waitForJobToFinish(created_job, t.jobInformer)
	if err != nil {
		return fmt.Errorf("Error while waiting for the Job to finish: %v", err)
	}

	// Use the job informer to keep track of the job

	// After the job finishes execution, keep track of each split location
	// and then create a job out of a container that takes in the data
	// from each split location, takes a parameter that tells it which samples
	// to take from the data, performs feedforward and backpropagation on the
	// data it takes, and then stores the gradient in a specified location.

	// Keep track of which jobs finished execution, and after that, create a job
	// that takes in the gradients from each of the jobs and then performs
	// averaging, finds the average gradient, and then updates the weights.

	// After the job finishes execution, do the same thing again from the start.
	return nil
}

func waitForJobToFinish(job *batchv1.Job, jobInformer cache.SharedIndexInformer) error {
	key, err := cache.MetaNamespaceKeyFunc(job)

	if err != nil {
		return err
	}

	for {
		jobObject, exists, err := jobInformer.GetIndexer().GetByKey(key)
		if err != nil {
			return err
		}
		if exists {
			job, ok := jobObject.(*batchv1.Job)
			if !ok {
				return fmt.Errorf("Error while converting the job object to job type")
			}
			if job.Status.Succeeded == 1 {
				return nil
			} else if job.Status.Failed == 1 {
				return fmt.Errorf("Job failed")
			}
		}
	}
}
