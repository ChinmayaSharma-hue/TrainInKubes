package controller

import (
	"context"
	"errors"
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
	job := createSplitJob(t.trainInKube, strconv.Itoa(6), configmap, t.namespace)

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
	// errorCh := make(chan error)
	// go waitForJobToFinish(created_job, t.jobInformer, errorCh)
	// err = <-errorCh
	// if err != nil {
	// 	return err
	// }
	wg.Add(1)
	go dummyWaitForJobToFinish(created_job, t.jobInformer)
	wg.Wait()

	// Use the job informer to keep track of the job

	// After the job finishes execution, keep track of each split location
	// and then create a job out of a container that takes in the data
	// from each split location, takes a parameter that tells it which samples
	// to take from the data, performs feedforward and backpropagation on the
	// data it takes, and then stores the gradient in a specified location.

	// Create a loop that runs for the number of epochs, creating 6 jobs for
	// each epoch, and then after each epoch, create a job that takes in the
	// gradients from each of the jobs and then performs averaging, finds the
	// average gradient, and then updates the weights.
	for i := 0; i < 1; i++ {
		startingIndex := 0
		// endingIndex is batch size divided by total number of jobs, and is also an integer,
		// so get the division result and get the integer rounded down
		endingIndex := int(trainInKube.Spec.BatchSize / 6)

		if trainInKube.Spec.BatchSize == 0 {
			return errors.New("Batch size cannot be 0")
		}
		numberOfMiniBatches := int(trainInKube.Spec.NumberOfSamples / trainInKube.Spec.BatchSize)

		for j := 0; j < numberOfMiniBatches; j++ {
			// Create a slice of jobs that will be created for each minibatch
			created_jobs := make([]*batchv1.Job, 6)
			for k := 0; k < 6; k++ {
				job := createTrainJob(t.trainInKube, t.namespace, k, startingIndex, endingIndex)

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
				// Add the job to the slice of jobs
				created_jobs[j] = created_job
			}
			// Wait until the execution of all the jobs finishes using go routines
			// doneCh := make(chan error, 6)
			// for _, job := range created_jobs {
			// 	go waitForJobToFinish(job, t.jobInformer, doneCh)
			// }
			// for l := 0; l < 6; l++ {
			// 	err := <-doneCh
			// 	if err != nil {
			// 		return err
			// 	}
			// }
			wg.Add(6)
			for _, job := range created_jobs {
				go dummyWaitForJobToFinish(job, t.jobInformer)
			}
			wg.Wait()

			t.logger.Infof("Finished executing all the jobs for epoch %d", i)

			// Delete all the jobs that were created for the minibatch
			for _, job := range created_jobs {
				err := t.kubeClientSet.BatchV1().Jobs(t.namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
				if err != nil {
					return fmt.Errorf("Error while deleting the Job: %v", err)
				}
			}
			deleteCh := make(chan error, 6)
			for _, job := range created_jobs {
				go waitForJobToBeDeleted(job, t.jobInformer, deleteCh)
			}
			for l := 0; l < 6; l++ {
				err := <-deleteCh
				if err != nil {
					return err
				}
			}

			// Create a job that averages over all the gradients
			job := createModelUpdateJob(t.trainInKube, strconv.Itoa(6), t.namespace)
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
			// go waitForJobToFinish(created_job, t.jobInformer, errorCh)
			// err = <-errorCh
			// if err != nil {
			// 	return err
			// }

			wg.Add(1)
			go dummyWaitForJobToFinish(created_job, t.jobInformer)
			wg.Wait()

			err = t.kubeClientSet.BatchV1().Jobs(t.namespace).Delete(ctx, created_job.Name, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("Error while deleting the Job: %v", err)
			}
			deletenbCh := make(chan error)
			go waitForJobToBeDeleted(created_job, t.jobInformer, deletenbCh)
			err = <-deletenbCh
			if err != nil {
				return err
			}

			startingIndex += endingIndex
			endingIndex += endingIndex
		}
	}
	// After the job finishes execution, do the same thing again from the start.
	return nil
}

func dummyWaitForJobToFinish(job *batchv1.Job, jobInformer cache.SharedIndexInformer) {
	defer wg.Done()
	key, err := cache.MetaNamespaceKeyFunc(job)

	if err != nil {
		panic(err)
	}

	for {
		jobObject, exists, err := jobInformer.GetIndexer().GetByKey(key)
		if err != nil {
			panic(err)
		}
		if exists {
			job, ok := jobObject.(*batchv1.Job)
			if !ok {
				panic(errors.New("Error while converting the job object to job type"))
			}
			if job.Status.Succeeded == 1 {
				return
			} else if job.Status.Failed == 1 {
				panic(errors.New("Job failed"))
			}
		}
	}

}

func waitForJobToFinish(job *batchv1.Job, jobInformer cache.SharedIndexInformer, errorCh chan error) {
	key, err := cache.MetaNamespaceKeyFunc(job)

	if err != nil {
		errorCh <- err
	}

	for {
		jobObject, exists, err := jobInformer.GetIndexer().GetByKey(key)
		if err != nil {
			errorCh <- err
		}
		if exists {
			job, ok := jobObject.(*batchv1.Job)
			if !ok {
				errorCh <- errors.New("Error while converting the job object to job type")
			}
			if job.Status.Succeeded == 1 {
				errorCh <- nil
			} else if job.Status.Failed == 1 {
				errorCh <- errors.New("Job failed")
			}
		}
	}
}

func waitForJobToBeDeleted(job *batchv1.Job, jobInformer cache.SharedIndexInformer, errorCh chan error) {
	key, err := cache.MetaNamespaceKeyFunc(job)

	if err != nil {
		errorCh <- err
	}

	for {
		_, exists, err := jobInformer.GetIndexer().GetByKey(key)
		if err != nil {
			errorCh <- err
		}
		if !exists {
			errorCh <- nil
		}
	}
}
