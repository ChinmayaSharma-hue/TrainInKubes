package controller

import (
	"context"
	"fmt"

	traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	cache "k8s.io/client-go/tools/cache"
)

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {

	}
}

func (c *Controller) processNextItem(ctx context.Context) {
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
		trainInKube, ok := event.newObj.(*traininkubev1alpha1.TrainInKube)
		return c.processAddTrainInKube(ctx, trainInKube)
	case addConfigMap:
		c.logger.Debugf("Processing the addConfigMap event")
	}

	return nil
}

func (c *Controller) processAddTrainInKube(ctx context.Context, trainInKube *traininkubev1alpha1.TrainInKube) error {
	// Create a ConfigMap for the TrainInKube
	configmap := createConfigMap(trainInKube, c.Namespace)

	exists, err := resourceExists(configmap, c.confgmapInformer.GetIndexer())
	if err != nil {
		return fmt.Errorf("error while checking if the ConfigMap already exists: %v", err)
	}
	if exists {
		c.logger.Infof("ConfigMap already exists, skipping creation")
		return nil
	}

	createConfigMap, err := c.kubeClient.CoreV1().ConfigMaps(c.Namespace).Create(ctx, configmap, metav1.CreateOptions{})

	// Add an event to the TrainInKube signalling the end of creation of CongigMap
	c.queue.Add(event{
		eventType: addConfigMap,
		newObj:    createConfigMap,
	})

	return err
}

func resourceExists(obj interface{}, indexer cache.Indexer) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		return false, fmt.Errorf("error while getting the key for the object: %v", err)
	}

	_, exists, err := indexer.GetByKey(key)
	return exists, err
}