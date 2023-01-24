package controller

import (
	"context"
	"fmt"

	hellov1alpha1 "github.com/ChinmayaSharma-hue/label-operator/pkg/apis/foo/v1alpha1"

	batchv1 "k8s.io/api/batch/v1"
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
		c.logger.Debug("Processing the item Hello.")
		return c.processAddHello(ctx, event.newObj.(*hellov1alpha1.Hello))
	case addHelloJob:
		c.logger.Debug("Processing the item Job.")
		return c.processAddHelloJob(ctx, event.job_resource.(*batchv1.Job), event.custom_resource.(*hellov1alpha1.Hello))
	}

	return nil
}

// The processing of the add event is done by creating a job and then checking if the job
// already exists. If the job already exists, it is skipped. If the job does not exist, it
// is created.
func (c *Controller) processAddHello(ctx context.Context, hello *hellov1alpha1.Hello) error {
	job := createJob(hello, c.namespace)

	exists, err := resourceExists(job, c.jobinformer.GetIndexer())
	if err != nil {
		return fmt.Errorf("error while checking if the job exists: %v", err)
	}
	if exists {
		c.logger.Infof("Job already exists, skipping creation")
		return nil
	}

	created_job, err := c.kubeClientSet.BatchV1().Jobs(c.namespace).Create(ctx, job, metav1.CreateOptions{})

	// I need to add the deepcopy of the job to the queue, so that I can process the completion of the job
	// with the same name and create five more jobs in response to the completion of this job. The thing is,
	// though, I have to have another informer somewhere in this package (here?) listening to the job with the same
	// name as the job I have in the event queue, so the deepcopy of the job in the event queue is just so that
	// I have some reference to the job whose completion would trigger the creation of five more jobs.
	c.queue.Add(event{
		eventType:       addHelloJob,
		custom_resource: hello,
		job_resource:    created_job,
	})

	return err
}

func (c *Controller) processAddHelloJob(ctx context.Context, hellojob *batchv1.Job, hello *hellov1alpha1.Hello) error {
	key, err := cache.MetaNamespaceKeyFunc(hellojob)
	if err != nil {
		return fmt.Errorf("Error while getting the key from the job: %v", err)
	}

	for {
		jobObject, exists, err := c.jobinformer.GetIndexer().GetByKey(key)
		if err != nil {
			return fmt.Errorf("Error while checking if the job exists: %v", err)
		}

		if exists {
			job, ok := jobObject.(*batchv1.Job)
			if !ok {
				return fmt.Errorf("Error while converting the job object to job type")
			}
			if job.Status.Succeeded == 1 {
				// Create five more jobs that do the same thing as this job, execute them in parallel? Find out how later.
				for i := 0; i < 5; i++ {
					new_job := createJobFromJob(job, hello, c.namespace, fmt.Sprintf("dependent-job-%d", i))

					exists, err := resourceExists(new_job, c.jobinformer.GetIndexer())
					if err != nil {
						return fmt.Errorf("error while checking if the job exists: %v", err)
					}

					if exists {
						c.logger.Infof("Job already exists, skipping creation.")
						return nil
					}

					_, err = c.kubeClientSet.BatchV1().Jobs(c.namespace).Create(ctx, new_job, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Error creating the job: %v", err)
					}
				}
				return nil
			} else if job.Status.Failed == 1 {
				return fmt.Errorf("Failed to create five more jobs as the first job seems to have failed.")
			}
		}
	}
}

func resourceExists(obj interface{}, indexer cache.Indexer) (bool, error) {
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		return false, fmt.Errorf("error while getting the key for the object: %v", err)
	}

	_, exists, err := indexer.GetByKey(key)
	return exists, err
}
