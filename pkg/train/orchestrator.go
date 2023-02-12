package train

import (
	"context"
	"errors"
	"fmt"
	"sync"

	traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"
	"github.com/ChinmayaSharma-hue/TrainInKubes/pkg/resources"
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
	KubeClientSet kubernetes.Interface
	TrainInKube   *traininkubev1alpha1.TrainInKube
	JobInformer   cache.SharedIndexInformer

	Namespace string

	Logger log.Logger
}

func (t *TrainOrchestrator) Run(ctx context.Context, TrainInKube *traininkubev1alpha1.TrainInKube) {
	t.Logger.Infof("Starting the job orchestrator...")

	err := t.Orchestrate(ctx, TrainInKube)
	if err != nil {
		t.Logger.Errorf("Error while orchestrating the jobs: %v", err)
	}
}

func (t *TrainOrchestrator) Orchestrate(ctx context.Context, TrainInKube *traininkubev1alpha1.TrainInKube) error {
	_, err := t.KubeClientSet.CoreV1().ConfigMaps(t.Namespace).Get(ctx, TrainInKube.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Error while getting the ConfigMap: %v", err)
	}

	// Create a job that divides the data between the jobs

	volume := resources.CreateHostPathVolume(TrainInKube.Name+"volume", "/data")
	volumeMount := resources.CreateVolumeMount(TrainInKube.Name+"volume", "/data")
	envVariables := map[string]string{
		"DIVISIONS":        strconv.Itoa(6),
		"DATASET_LOCATION": "/data/PreprocessedData",
		"SPLIT_LOCATION":   "/data/Chunks",
	}
	ownerReference := resources.CreateOwnerReference(TrainInKube)

	job := resources.CreateJob(
		resources.CreateJobWithName(TrainInKube.Name+"splitdata"),
		resources.CreateJobWithImage("splitjob:latest"),
		resources.CreateJobInNamespace(t.Namespace),
		resources.CreateJobWithVolume(volume),
		resources.CreateJobWithVolumeMounts(volumeMount),
		resources.CreateJobWithEnv(envVariables),
		resources.CreateJobWithOwnerReference(ownerReference),
	)

	exists, err := resourceExists(job, t.JobInformer.GetIndexer())
	if err != nil {
		return fmt.Errorf("Error while checking if the Job already exists: %v", err)
	}
	if exists {
		t.Logger.Infof("Job already exists, skipping creation")
		return nil
	}

	created_job, err := t.KubeClientSet.BatchV1().Jobs(t.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Error while creating the Job: %v", err)
	}

	// Block the function till the job finishes execution
	errorCh := make(chan error)
	go waitForJobToFinish(created_job, t.JobInformer, errorCh)
	err = <-errorCh
	if err != nil {
		return err
	}

	for i := 0; i < 1; i++ {
		startingIndex := 0
		endingIndex := int(TrainInKube.Spec.BatchSize / 6)

		if TrainInKube.Spec.BatchSize == 0 {
			return errors.New("Batch size cannot be 0")
		}

		numberOfMiniBatches := int(TrainInKube.Spec.NumberOfSamples / TrainInKube.Spec.BatchSize)

		for j := 0; j < numberOfMiniBatches; j++ {
			created_jobs := make([]*batchv1.Job, 6)
			for k := 0; k < 6; k++ {
=				volume := resources.CreateHostPathVolume(TrainInKube.Name+"volume", "/data")
				volumeMount := resources.CreateVolumeMount(TrainInKube.Name+"volume", "/data")
				envVariables := map[string]string{
					"MODEL_LOCATION":    "/data/model.h5",
					"GRADIENT_LOCATION": "/data/Gradients",
					"FEATURES_LOCATION": "/data/Chunks/x_train_" + strconv.Itoa(k) + ".npy",
					"LABELS_LOCATION":   "/data/Chunks/y_train_" + strconv.Itoa(k) + ".npy",
					"STARTING_INDEX":    strconv.Itoa(startingIndex),
					"ENDING_INDEX":      strconv.Itoa(endingIndex),
					"JOB_INDEX":         strconv.Itoa(k),
				}
				ownerReference := resources.CreateOwnerReference(TrainInKube)

				job := resources.CreateJob(
					resources.CreateJobWithName(TrainInKube.Name+"traimodel"+strconv.Itoa(k)),
					resources.CreateJobWithImage("trainjob:latest"),
					resources.CreateJobInNamespace(t.Namespace),
					resources.CreateJobWithVolume(volume),
					resources.CreateJobWithVolumeMounts(volumeMount),
					resources.CreateJobWithEnv(envVariables),
					resources.CreateJobWithOwnerReference(ownerReference),
				)

				exists, err := resourceExists(job, t.JobInformer.GetIndexer())
				if err != nil {
					return fmt.Errorf("Error while checking if the Job already exists: %v", err)
				}
				if exists {
					t.Logger.Infof("Job already exists, skipping creation")
					return nil
				}

				created_job, err := t.KubeClientSet.BatchV1().Jobs(t.Namespace).Create(ctx, job, metav1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("Error while creating the Job: %v", err)
				}
				// Add the job to the slice of jobs
				created_jobs[k] = created_job
			}
			// Wait until the execution of all the jobs finishes using go routines
			doneCh := make(chan error, 6)
			for _, job := range created_jobs {
				go waitForJobToFinish(job, t.JobInformer, doneCh)
			}
			for l := 0; l < 6; l++ {
				err := <-doneCh
				if err != nil {
					return err
				}
			}

			t.Logger.Infof("Finished executing all the jobs for epoch %d", i)

			// Delete all the jobs that were created for the minibatch
			for _, job := range created_jobs {
				err := t.KubeClientSet.BatchV1().Jobs(t.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
				if err != nil {
					return fmt.Errorf("Error while deleting the Job: %v", err)
				}
			}
			deleteCh := make(chan error, 6)
			for _, job := range created_jobs {
				go waitForJobToBeDeleted(job, t.JobInformer, deleteCh)
			}
			for l := 0; l < 6; l++ {
				err := <-deleteCh
				if err != nil {
					return err
				}
			}

			// Create a job that averages over all the gradients
			volume := resources.CreateHostPathVolume(TrainInKube.Name+"volume", "/data")
			volumeMount := resources.CreateVolumeMount(TrainInKube.Name+"volume", "/data")
			envVariables := map[string]string{
				"MODEL_LOCATION":    "/data/model.h5",
				"GRADIENT_LOCATION": "/data/Gradients",
				"NUMBER_OF_GRADS":   strconv.Itoa(6),
			}
			ownerReference := resources.CreateOwnerReference(TrainInKube)

			job := resources.CreateJob(
				resources.CreateJobWithName(TrainInKube.Name+"updatemodel"),
				resources.CreateJobWithImage("modelupdatejob:latest"),
				resources.CreateJobInNamespace(t.Namespace),
				resources.CreateJobWithVolume(volume),
				resources.CreateJobWithVolumeMounts(volumeMount),
				resources.CreateJobWithEnv(envVariables),
				resources.CreateJobWithOwnerReference(ownerReference),
			)

			exists, err := resourceExists(job, t.JobInformer.GetIndexer())
			if err != nil {
				return fmt.Errorf("Error while checking if the Job already exists: %v", err)
			}
			if exists {
				t.Logger.Infof("Job already exists, skipping creation")
				return nil
			}
			created_job, err := t.KubeClientSet.BatchV1().Jobs(t.Namespace).Create(ctx, job, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Error while creating the Job: %v", err)
			}
			go waitForJobToFinish(created_job, t.JobInformer, errorCh)
			err = <-errorCh
			if err != nil {
				return err
			}

			err = t.KubeClientSet.BatchV1().Jobs(t.Namespace).Delete(ctx, created_job.Name, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("Error while deleting the Job: %v", err)
			}
			deletenbCh := make(chan error)
			go waitForJobToBeDeleted(created_job, t.JobInformer, deletenbCh)
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

func waitForJobToFinish(job *batchv1.Job, JobInformer cache.SharedIndexInformer, errorCh chan error) {
	key, err := cache.MetaNamespaceKeyFunc(job)
	if err != nil {
		errorCh <- err
	}

	for {
		jobObject, exists, err := JobInformer.GetIndexer().GetByKey(key)
		if err != nil {
			errorCh <- err
			break
		}
		if exists {
			job, ok := jobObject.(*batchv1.Job)
			if !ok {
				errorCh <- errors.New("Error while converting the job object to job type")
				break
			}
			if job.Status.Succeeded == 1 {
				errorCh <- nil
				break
			} else if job.Status.Failed == 1 {
				errorCh <- errors.New("Job failed")
				break
			}
		}
	}
}

func waitForJobToBeDeleted(job *batchv1.Job, JobInformer cache.SharedIndexInformer, errorCh chan error) {
	key, err := cache.MetaNamespaceKeyFunc(job)

	if err != nil {
		errorCh <- err
	}

	for {
		_, exists, err := JobInformer.GetIndexer().GetByKey(key)
		if err != nil {
			errorCh <- err
			break
		}
		if !exists {
			errorCh <- nil
			break
		}
	}
}

func resourceExists(obj interface{}, indexer cache.Indexer) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		return false, fmt.Errorf("error while getting the key for the object: %v", err)
	}

	_, exists, err := indexer.GetByKey(key)
	return exists, err
}
