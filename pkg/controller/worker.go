package controller

import (
	"context"
	"fmt"

	traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	cache "k8s.io/client-go/tools/cache"
)

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {

	}
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	obj, shutdown := c.queue.Get()

	if shutdown {
		return false
	}

	defer c.queue.Done(obj)

	err := c.processItem(ctx, obj)
	if err == nil {
		c.logger.Debugf("Successfully processed the item")
		c.queue.Forget(obj)
	} else if c.queue.NumRequeues(obj) < 3 {
		c.logger.Errorf("Failed to process the item, requeuing it: %v", err)
	} else {
		c.logger.Errorf("Failed to process the item, dropping it: %v", err)
		c.queue.Forget(obj)
		utilruntime.HandleError(err)
	}

	return true
}

func (c *Controller) processItem(ctx context.Context, obj interface{}) error {
	event, ok := obj.(event)

	if !ok {
		c.logger.Errorf("Failed to cast the item to event")
		return fmt.Errorf("Failed to cast the item to event")
	}

	switch event.eventType {
	case addTrainInKube:
		c.logger.Debugf("Processing the addTrainInKube event")
		return c.processAddTrainInKube(ctx, event.newObj.(*traininkubev1alpha1.TrainInKube))
	case addConfigMap:
		c.logger.Debugf("Processing the addConfigMap event")
		return c.processAddConfigMap(ctx, event.oldObj.(*traininkubev1alpha1.TrainInKube), event.newObj.(*corev1.ConfigMap))
	case addBuildModel:
		c.logger.Debugf("Processing the addBuildModel event")
	}

	return nil
}

func (c *Controller) processAddTrainInKube(ctx context.Context, trainInKube *traininkubev1alpha1.TrainInKube) error {
	// Create a ConfigMap for the TrainInKube
	configmap := createConfigMap(trainInKube, c.namespace)

	exists, err := resourceExists(configmap, c.confgmapInformer.GetIndexer())
	if err != nil {
		return fmt.Errorf("error while checking if the ConfigMap already exists: %v", err)
	}
	if exists {
		c.logger.Infof("ConfigMap already exists, skipping creation")
		return nil
	}

	createdConfigMap, err := c.kubeClientSet.CoreV1().ConfigMaps(c.namespace).Create(ctx, configmap, metav1.CreateOptions{})

	// Add an event to the TrainInKube signalling the end of creation of CongigMap
	c.queue.Add(event{
		eventType: addConfigMap,
		newObj:    createdConfigMap,
		oldObj:    trainInKube,
	})

	return err
}

func (c *Controller) processAddConfigMap(
		ctx context.Context, 
		trainInKube *traininkubev1alpha1.TrainInKube, 
		configmap *corev1.ConfigMap
	) error {
	// Create a Job to build the model
	job := createJob(trainInKube, configmap, c.namespace)

	exists, err := resourceExists(job, c.jobInformer.GetIndexer())
	if err != nil {
		return fmt.Errorf("Error while checking if the Job already exists: %v", err)
	}
	if exists {
		c.logger.Infof("Job already exists, skipping creation")
		return nil
	}

	created_job, err := c.kubeClientSet.BatchV1().Jobs(c.namespace).Create(ctx, job, metav1.CreateOptions{})

	// Add an event to the TrainInKube signalling the end of creation of Job
	c.queue.Add(event{
		eventType: addBuildModel,
		newObj:    created_job,
	})

	return nil
}

func resourceExists(obj interface{}, indexer cache.Indexer) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		return false, fmt.Errorf("error while getting the key for the object: %v", err)
	}

	_, exists, err := indexer.GetByKey(key)
	return exists, err
}
