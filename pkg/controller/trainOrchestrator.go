package controller

import (
	"fmt"

	"github.com/gotway/gotway/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type TrainOrchestrator struct {
	kubeClientSet kubernetes.Interface

	jobInformer  cache.SharedIndexInformer
	nodeInformer cache.SharedIndexInformer

	logger log.Logger
}

func (t *TrainOrchestrator) Run(stopCh <-chan struct{}) {
	t.logger.Infof("Starting the job orchestrator...")

	select {
	case <-done:
		fmt.Println("Orchestrator stopped.")
	}
}
