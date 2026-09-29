package controller

import (
	"cmp"

	"k8s.io/apimachinery/pkg/util/intstr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	llmmodelv1alpha1 "host-llm.io/k8s-model-operator/api/v1alpha1"
)

// nameEqual returns true if services are equal
func serviceEqual(svc1 *corev1.Service, svc2 *corev1.Service) bool {
	return (svc1.Name == svc2.Name)
}

// desiredServiceForModel returns a Service object for the Model CR.
func (r *ModelReconciler) desiredServiceForModel(model *llmmodelv1alpha1.Model) *corev1.Service {
	ls := r.desiredLabelSelector(model)
	namespace := cmp.Or(model.Spec.TargetNamespace, "default")
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      model.Name,
			Namespace: namespace,
			Labels:    ls,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "ollama",
					Protocol:   "TCP",
					TargetPort: intstr.FromInt32(11434),
					Port:       11434,
				},
			},
			Selector: ls,
		},
	}

	// Set the Model instance as the owner and controller
	err := ctrl.SetControllerReference(model, svc, r.Scheme)
	if err != nil {
		log := logf.Log
		log.Error(err, "Failed to create a service reference")
	}
	return svc
}
