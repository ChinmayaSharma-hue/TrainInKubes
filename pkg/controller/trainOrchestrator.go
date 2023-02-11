package controller

import (
	"context"

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

func (t *TrainOrchestrator) Run(ctx context.Context) {
	t.logger.Infof("Starting the job orchestrator...")
	// Will think about the ctx.Done() later
	<-ctx.Done()
}
