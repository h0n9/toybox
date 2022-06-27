package controller

import (
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	csiWebhook "github.com/h0n9/toybox/cloud-secrets-injector/webhook"
)

func init() {
	log.SetLogger(zap.New())
}

var Cmd = &cobra.Command{
	Use:   "controller",
	Short: "run a server serving as a admission webhook controller manager",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.Log.WithName("cloud-secrets-injector-controller")

		mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{Logger: logger})
		if err != nil {
			logger.Error(err, "faild to setup manager")
			os.Exit(1)
		}

		hookServer := mgr.GetWebhookServer()
		hookServer.Register("/mutate", &webhook.Admission{Handler: &csiWebhook.Mutator{
			Client: mgr.GetClient(),
		}})
		hookServer.Register("/validate", &webhook.Admission{Handler: &csiWebhook.Validator{
			Client: mgr.GetClient(),
		}})
		logger.Info("registered mutate, validator handlers to /mutate, /validate webhook uris")

		logger.Info("starting manager")
		err = mgr.Start(signals.SetupSignalHandler())
		if err != nil {
			logger.Error(err, "failed to run manager")
			os.Exit(1)
		}
	},
}
