package controller

import (
	"context"
	"errors"
	"time"

	"github.com/gotway/gotway/pkg/log"

	hellov1alpha1 "github.com/ChinmayaSharma-hue/label-operator/pkg/apis/foo/v1alpha1"
	hellov1alpha1clientset "github.com/ChinmayaSharma-hue/label-operator/pkg/client/clientset/versioned"
	helloinformers "github.com/ChinmayaSharma-hue/label-operator/pkg/client/informers/externalversions"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	kubeClientSet kubernetes.Interface

	helloinformer cache.SharedIndexInformer
	jobinformer   cache.SharedIndexInformer

	queue workqueue.RateLimitingInterface

	namespace string

	logger log.Logger
}

// First, we define the logic of the controller when it is running. What it does is,
// it first uses the informers to listen to the events, and then it uses the queue to
// add events to the queue. Then it starts 4 workers that each start  a go routine that
// routinely checks the queue for any events and do the necessary processing on the events.
// Finally, a blocking call is made which is receiving the value of a channel that finishes
// execution when it is cancelled, so the function can finally stop running. Therefore, until
// the "context" is cancelled, the controller will keep running.
func (c *Controller) Run(ctx context.Context) error {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Infof("Starting the controller")

	c.logger.Infof("Starting the informers")
	for _, i := range []cache.SharedIndexInformer{
		c.helloinformer,
		c.jobinformer,
	} {
		go i.Run(ctx.Done())
	}

	c.logger.Infof("Waiting for the informers to sync")
	if !cache.WaitForCacheSync(ctx.Done(), []cache.InformerSynced{
		c.helloinformer.HasSynced,
		c.jobinformer.HasSynced,
	}...) {
		err := errors.New("failed to wait for informers to sync")
		utilruntime.HandleError(err)
		return err
	}

	c.logger.Infof("Starting 4 workers")
	for i := 0; i < 4; i++ {
		go wait.Until(func() {
			c.runWorker(ctx)
		}, time.Second, ctx.Done())
	}

	c.logger.Infof("Controller Ready")

	<-ctx.Done()
	c.logger.Infof("Shutting down the controller")

	return nil
}

// This function is the event handler that is run when the hello resource is added to the cluster.
// It is called by the informer when it receives an event that a hello resource is added to the cluster.
func (c *Controller) addHello(obj interface{}) {
	c.logger.Debug("Adding hello")

	hello, ok := obj.(*hellov1alpha1.Hello)

	if !ok {
		c.logger.Errorf("Error while converting the object to hello type")
		return
	}
	c.queue.Add(event{
		eventType: addHello,
		newObj:    hello,
	})
}

// This function creates a Controller object that has all the fields that can be used when the controller is running.
// First, we create the informer factories of both the hello and the kubernetes clientsets. Then we create the informers
// for the same. Then we create the queue that is used to add events to the queue. The event handler for when the hello
// resource is added is specified here. Finally, we return the controller object.
func New(
	kubeClientSet kubernetes.Interface,
	helloclientset hellov1alpha1clientset.Interface,
	namespace string,
	logger log.Logger,
) *Controller {
	helloInformerFactory := helloinformers.NewSharedInformerFactory(
		helloclientset,
		time.Second*10,
	)

	helloinformer := helloInformerFactory.Foo().V1alpha1().Hellos().Informer()

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(
		kubeClientSet,
		time.Second*10,
	)
	jobinformer := kubeInformerFactory.Batch().V1().Jobs().Informer()

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	ctrl := &Controller{
		kubeClientSet: kubeClientSet,
		helloinformer: helloinformer,
		jobinformer:   jobinformer,
		queue:         queue,
		namespace:     namespace,
		logger:        logger,
	}

	helloinformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.addHello,
	})

	return ctrl
}
