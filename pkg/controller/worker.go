package controller

import (
	"context"
	"fmt"

	hellov1alpha1 "github.com/ChinmayaSharma-hue/label-operator/pkg/apis/foo/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	cache "k8s.io/client-go/tools/cache"
)

// This function is called routinely by the workers to check the queue for any events.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {

	}
}

// In the processing of the items in the queue, each item is obtained from the queue and
// then it is processed. If the processing is successful, the item is removed from the queue.
// If the process is unsuccessful, the item is requeued. If the item is requeued more than
// 3 times, it is dropped from the queue. The function returns true if the item is successfully
// processed, and false if the item is not successfully processed.
func (c *Controller) processNextItem(ctx context.Context) bool {
	obj, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	defer c.queue.Done(obj)

	err := c.processItem(ctx, obj)
	if err == nil {
		c.logger.Debug("Successfully processed the item")
		c.queue.Forget(obj)
	} else if c.queue.NumRequeues(obj) < 3 {
		c.logger.Errorf("Failed to process the item, requeuing it")
		c.queue.AddRateLimited(obj)
	} else {
		c.logger.Errorf("Failed to process the item, dropping it")
		c.queue.Forget(obj)
		utilruntime.HandleError(err)
	}

	return true
}

// The processing of each item happens by adding a case for each of the events, but in this
// case there is only one event, so only one case is handled by a function.
func (c *Controller) processItem(ctx context.Context, obj interface{}) error {
	event, ok := obj.(event)

	if !ok {
		c.logger.Errorf("Error while converting the object to event type")
		return nil
	}

	switch event.eventType {
	case addHello:
		return c.processAddHello(ctx, event.newObj.(*hellov1alpha1.HelloType))
	}

	return nil
}

// The processing of the add event is done by creating a job and then checking if the job
// already exists. If the job already exists, it is skipped. If the job does not exist, it
// is created.
func (c *Controller) processAddHello(ctx context.Context, hello *hellov1alpha1.HelloType) error {
	job := createJob(hello, c.namespace)

	exists, err := resourceExists(job, c.jobinformer.GetIndexer())
	if err != nil {
		return fmt.Errorf("error while checking if the job exists: %v", err)
	}
	if exists {
		c.logger.Infof("Job already exists, skipping creation")
		return nil
	}

	_, err = c.kubeClientSet.BatchV1().Jobs(c.namespace).Create(ctx, job, metav1.CreateOptions{})
	return err
}

func resourceExists(obj interface{}, indexer cache.Indexer) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		return false, fmt.Errorf("error while getting the key for the object: %v", err)
	}

	_, exists, err := indexer.GetByKey(key)
	return exists, err
}
