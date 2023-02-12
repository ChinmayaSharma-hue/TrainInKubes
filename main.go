package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gotway/gotway/pkg/log"

	traininkubev1alpha1clientset "github.com/ChinmayaSharma-hue/TrainInKubes/pkg/client/clientset/versioned"
	"github.com/ChinmayaSharma-hue/TrainInKubes/pkg/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	var restConfig *rest.Config
	var errKubeConfig error

	// Create a logger
	logger := log.NewLogger(
		log.Fields{
			"service": "Train-Operator",
		}, "local",
		"debug",
		os.Stdout,
	)

	logger.Debugf("Starting the controller...")

	restConfig, errKubeConfig = rest.InClusterConfig()
	if errKubeConfig != nil {
		logger.Errorf("Error while building the kubeconfig: %v", errKubeConfig)
	}

	kubeClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.Errorf("Error while creating the kubernetes clientset: %v", err)
	}

	traininkubev1alpha1ClientSet, err := traininkubev1alpha1clientset.NewForConfig(restConfig)
	if err != nil {
		logger.Errorf("Error while creating the TrainInKube clientset: %v", err)
	}

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

	err = ctrl.Run(ctx)
	if err != nil {
		logger.Fatal("Error running controller ", err)
	}
}
