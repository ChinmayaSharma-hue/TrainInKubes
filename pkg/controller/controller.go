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

	trainInKubeInformer cache.SharedIndexInformer
	jobInformer 	   cache.SharedIndexInformer
	nodeInformer 	   cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	namespace string 

	logger log.Logger
}

func New(
	kubeClientSet kubernetes.Interface,
	traininkubev1alpha1ClientSet traininkubev1alpha1clientset.Interface(),
	namespace string,
	logger log.Logger,
) *Controller {
	traininkubeInformerFactory := traininkubev1alpha1informers.NewSharedInformerFactory(
		traininkubev1alpha1ClientSet,
		time.Second*10,
	)

	traininkubeInformer := traininkubeInformerFactory.Trainink8s().V1Alpha1().TrainInKubes().Informer()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(
		kubeClientSet,
		time.Second*10,
	)

	jobInformer := kubeInformerFactory.Batch().V1().Jobs().Informer()
	nodeInformer := kubeInformerFactory.V1().Nodes().Informer()

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	ctrl := &Controller{
		kubeClientSet: kubeClientSet,
		traininkubeInformer: traininkubeInformer,
		jobInformer: jobInformer,
		nodeInformer: nodeInformer,
		queue: queue,
		namespace: namespace,
		logger: logger,
	}

	return ctrl
}