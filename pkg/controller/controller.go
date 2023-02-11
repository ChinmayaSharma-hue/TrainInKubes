package controller

import (
	"context"
	"errors"
	"time"

	"github.com/gotway/gotway/pkg/log"

	traininkubev1alpha1 "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/apis/trainink8s/v1alpha1"
	traininkubev1alpha1clientset "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/client/clientset/versioned"
	traininkubev1alpha1informers "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/client/informers/externalversions"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	kubeClientSet kubernetes.Interface

	traininkubeInformer cache.SharedIndexInformer
	configmapInformer   cache.SharedIndexInformer
	jobInformer         cache.SharedIndexInformer
	nodeInformer        cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	namespace string

	logger log.Logger
}

func (c *Controller) Run(ctx context.Context) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Infof("Starting the controller...")

	c.logger.Infof("Starting the informers...")
	for _, i := range []cache.SharedIndexInformer{
		c.traininkubeInformer,
		c.jobInformer,
		c.nodeInformer,
	} {
		go i.Run(ctx.Done())
	}

	c.logger.Infof("Waiting for the informers to sync...")
	if !cache.WaitForCacheSync(ctx.Done(), []cache.InformerSynced{
		c.traininkubeInformer.HasSynced,
		c.jobInformer.HasSynced,
		c.nodeInformer.HasSynced,
	}...) {
		err := errors.New("Failed to wait for informers to sync")
		utilruntime.HandleError(err)
		return err
	}

	c.logger.Infof("Starting 4 workers...")
	for i := 0; i < 4; i++ {
		go wait.Until(func() {
			c.runWorker(ctx)
		}, time.Second, ctx.Done())
	}

	c.logger.Infof("Controller Ready")

	<-ctx.Done()
	c.logger.Infof("Shutting down the controller")

	return nil
}

func (c *Controller) addTrainInKube(obj interface{}) {
	c.logger.Debugf("Adding TrainInKube")

	traininkube, ok := obj.(*traininkubev1alpha1.TrainInKube)

	if !ok {
		c.logger.Errorf("Error while converting the object to TrainInKube")
	}

	c.queue.Add(event{
		eventType:      addTrainInKube,
		customResource: traininkube,
	})
}

func New(
	kubeClientSet kubernetes.Interface,
	traininkubev1alpha1ClientSet traininkubev1alpha1clientset.Interface,
	namespace string,
	logger log.Logger,
) *Controller {
	traininkubeInformerFactory := traininkubev1alpha1informers.NewSharedInformerFactory(
		traininkubev1alpha1ClientSet,
		time.Second*10,
	)

	traininkubeInformer := traininkubeInformerFactory.Foo().V1alpha1().TrainInKubes().Informer()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(
		kubeClientSet,
		time.Second*10,
	)

	configmapInformer := kubeInformerFactory.Core().V1().ConfigMaps().Informer()
	jobInformer := kubeInformerFactory.Batch().V1().Jobs().Informer()
	nodeInformer := kubeInformerFactory.Core().V1().Nodes().Informer()

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	ctrl := &Controller{
		kubeClientSet:       kubeClientSet,
		traininkubeInformer: traininkubeInformer,
		configmapInformer:   configmapInformer,
		jobInformer:         jobInformer,
		nodeInformer:        nodeInformer,
		queue:               queue,
		namespace:           namespace,
		logger:              logger,
	}

	traininkubeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.addTrainInKube,
	})

	return ctrl
}
