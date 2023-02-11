package controller

import (
	"context"
	"fmt"
	"sync"

	traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"
	"github.com/gotway/gotway/pkg/log"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"strconv"
)

var (
	wg sync.WaitGroup
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
	wg.Add(1)
	go waitForJobToFinish(created_job, t.jobInformer)
	wg.Wait()
	if err != nil {
		return fmt.Errorf("Error while waiting for the Job to finish: %v", err)
	}

	// Use the job informer to keep track of the job

	// After the job finishes execution, keep track of each split location
	// and then create a job out of a container that takes in the data
	// from each split location, takes a parameter that tells it which samples
	// to take from the data, performs feedforward and backpropagation on the
	// data it takes, and then stores the gradient in a specified location.

	// Create a loop that runs for the number of epochs, creating 5 jobs for
	// each epoch, and then after each epoch, create a job that takes in the
	// gradients from each of the jobs and then performs averaging, finds the
	// average gradient, and then updates the weights.
	startingIndex := 0
	endingIndex := 2
	for i := 0; i < 1; i++ {
		// Create a slice of jobs that will be created for each epoch
		created_jobs := make([]*batchv1.Job, 5)

		for j := 0; j < 5; j++ {
			job := createTrainJob(t.trainInKube, t.namespace, j, startingIndex, endingIndex)

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
			fmt.Println("Job Created!")
			// Add the job to the slice of jobs
			created_jobs[j] = created_job
		}

		// Wait until the execution of all the jobs finishes using go routines
		wg.Add(5)
		for _, job := range created_jobs {
			go waitForJobToFinish(job, t.jobInformer)
		}
		wg.Wait()

		t.logger.Infof("Finished executing all the jobs for epoch %d", i)
	}

	// After the job finishes execution, do the same thing again from the start.
	return nil
}

func waitForJobToFinish(job *batchv1.Job, jobInformer cache.SharedIndexInformer) {
	defer wg.Done()

	key, err := cache.MetaNamespaceKeyFunc(job)

	if err != nil {
		fmt.Errorf("Error while getting the key for the job: %v", err)
	}

	for {
		jobObject, exists, err := jobInformer.GetIndexer().GetByKey(key)
		if err != nil {
			fmt.Errorf("Error while getting the job object from the informer: %v", err)
		}
		if exists {
			job, ok := jobObject.(*batchv1.Job)
			if !ok {
				fmt.Errorf("Error while converting the job object to job type")
			}
			if job.Status.Succeeded == 1 {
				return
			} else if job.Status.Failed == 1 {
				fmt.Errorf("Job failed")
			}
		}
	}
}
