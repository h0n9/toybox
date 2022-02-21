package main

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	ctx       context.Context
	cfg       *rest.Config
	clientSet *kubernetes.Clientset
}

func NewClient(kubeconfigPath string) (*Client, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	return &Client{
		ctx:       context.Background(),
		cfg:       cfg,
		clientSet: clientSet,
	}, nil
}
