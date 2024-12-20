package httpapi

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1 "dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
	"dcnlab.ssu.ac.kr/kt-cloud-operator/internal/cloudapi"
	"github.com/kelseyhightower/envconfig"
)

func ProcessEnvVariables() (cloudapi.Config, *zap.SugaredLogger) {
	var Config1 cloudapi.Config
	var logger2 *zap.SugaredLogger
	err := envconfig.Process("", &Config1)
	if err != nil {
		panic(err.Error())
	}
	err, logger2 = logger(Config1.LogLevel)
	if err != nil {
		panic(err.Error())
	}

	logger2.Info("Processed Env Variables...")
	return Config1, logger2
}

func logger(logLevel string) (error, *zap.SugaredLogger) {
	var level zapcore.Level
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		return err, nil
	}
	logConfig := zap.NewDevelopmentConfig()
	logConfig.Level.SetLevel(level)
	log, err := logConfig.Build()
	if err != nil {
		return err, nil
	}
	return nil, log.Sugar()
}

func getRestConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	return ctrl.GetConfig()
}

// getClient initializes a controller-runtime Manager and returns the client it uses.
func getClient(config *rest.Config, scheme *runtime.Scheme) (client.Client, error) {
	// Register your custom resource's types
	if err := v1beta1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add custom resources to scheme: %v", err)
	}

	return client.New(config, client.Options{Scheme: scheme})
}
