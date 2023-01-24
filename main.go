package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gotway/gotway/pkg/log"

	helloalpha1clientset "github.com/ChinmayaSharma-hue/label-operator/pkg/client/clientset/versioned"
	"github.com/ChinmayaSharma-hue/label-operator/pkg/controller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Later, when this controller is being run inside the kubernetes cluster,
	// the kubeconfig file will be provided by the kubernetes cluster itself.
	var restConfig *rest.Config
	var errKubeConfig error

	// Create a logger
	logger := log.NewLogger(
		log.Fields{
			"service": "Hello-Operator",
		}, "local",
		"debug",
		os.Stdout,
	)
	logger.Debugf("Starting the controller")

	// Get the kubeconfig file, and then build the rest config
	kubeconfig := flag.String("kubeconfig", "/home/chinmay/.kube/config", "kubeconfig file")
	flag.Parse()

	restConfig, errKubeConfig = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if errKubeConfig != nil {
		logger.Errorf("Error while building the kubeconfig: %v", errKubeConfig)
	}

	// restConfig, errKubeConfig = rest.InClusterConfig()
	// if errKubeConfig != nil {
	// 	logger.Errorf("Error while building the kubeconfig: %v", errKubeConfig)
	// }

	// Create a clientset for the kubernetes cluster
	kubeClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		logger.Errorf("Error while creating the kubernetes clientset: %v", err)
	}

	helloalpha1ClientSet, err := helloalpha1clientset.NewForConfig(restConfig)
	if err != nil {
		logger.Errorf("Error while creating the helloalpha1 clientset: %v", err)
	}

	ctrl := controller.New(
		kubeClientSet,
		helloalpha1ClientSet,
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
		logger.Fatal("error running controller ", err)
	}

}
