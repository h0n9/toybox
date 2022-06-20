package main

import (
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func init() {
	log.SetLogger(zap.New())
}

func main() {
	// init logger
	logger := log.Log.WithName("cloud-secrets-controller")

	_, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Logger: logger,
	})
	if err != nil {
		logger.Error(err, "unable to se up overall controller manager")
		os.Exit(1)
	}
}
