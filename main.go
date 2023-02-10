package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gotway/gotway/pkg/log"

	traininkubev1alpha1clientset "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/client/clientset/versioned"
	"github.com/ChinmayaSharma-hue/TrainInKubes/pkg/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
)

func main() {
	var restConfig *rest.Config
	var errKubeConfig error

	// Create a logger
	// Find out what each of the arguments in the log.NewLogger() function does, and what the log.Fields{} does
	logger := log.NewLogger(
		log.Fields{
			"service": "Train-Operator",
		}, "local",
		"debug",
		os.Stdout,
	)

	logger.Debugf("Starting the controller...")

	// Get the kubeConfig file and then build the restConfig
	// Find out what flag.String() does
	kubeConfig := flag.String("kubeconfig", "/home/chinmay/.kube/config", "kubeconfig file")
	flag.Parse()

	restConfig, errKubeConfig = clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if errKubeConfig != nil {
		logger.Errorf("Error while building the kubeconfig: %v", errKubeConfig)
	}

	// restConfig, errKubeConfig = rest.InClusterConfig()
	// if errKubeConfig != nil {
	// 	logger.Errorf("Error while building the kubeconfig: %v", errKubeConfig)
	// }

	kubeClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.Errorf("Error while creating the kubernetes clientset: %v", err)
	}

	traininkubev1alpha1ClientSet, err := traininkubev1alpha1clientset.NewForConfig(restConfig)
	if err != nil {
		logger.Errorf("Error while creating the TrainInKube clientset: %v", err)
	}

	// First, create a ConfigMap containing all the paths for preprocessed data, split data and the models.
	// Then, create the job which creates the model and saves it in the path as specified in the ConfigMap.

	// Below is the code for the controller
	// Then, find a way to check resources available and determine the number of jobs to be created.
	// Then, create the job with the number of splits as the environment variable in the Job.
	// Then, create the training jobs.
	// Keep monitoring the resources available and trigger an event and add it to workqueue in case of a resource shortage
	// or a resource increase.
	// Based on the event, create or delete jobs.
	// Optional - Keep monitoring the jobs and trigger an event in case of a job failure, update the status of the TrainInKube object.

	// Creating a new controller
	ctrl := controller.New(
		kubeClientSet,
		traininkubev1alpha1ClientSet,
		"default",
		logger.WithField("type", "controller"),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), []os.Signal{
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	}...)
	defer cancel()
}
