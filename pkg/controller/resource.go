package controller

import (
	hellov1alpha1 "github.com/ChinmayaSharma-hue/label-operator/pkg/apis/foo/v1alpha1"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createJob(newHello *hellov1alpha1.Hello, namespace string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newHello.ObjectMeta.Name,
			Namespace: namespace,
			Labels:    make(map[string]string),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(newHello, hellov1alpha1.SchemeGroupVersion.WithKind("HelloType")),
			},
		},
		Spec: createJobSpec(newHello.Name, namespace, newHello.Spec.Message),
	}
}

func createJobSpec(name, namespace, message string) batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: name + "-",
				Namespace:    namespace,
				Labels:       make(map[string]string),
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            name,
						Image:           "busybox:1.33.1",
						Command:         []string{"echo", message},
						ImagePullPolicy: "IfNotPresent",
					},
				},
				RestartPolicy: corev1.RestartPolicyNever,
			},
		},
	}
}
