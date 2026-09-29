package controller

import (
	"cmp"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	llmmodelv1alpha1 "host-llm.io/k8s-model-operator/api/v1alpha1"
)

// nameEqual returns true if services are equal
func hpaEqual(hpa1 *autoscalingv2.HorizontalPodAutoscaler, hpa2 *autoscalingv2.HorizontalPodAutoscaler) bool {
	return (hpa1.Spec.MinReplicas == hpa2.Spec.MinReplicas && hpa1.Spec.MaxReplicas == hpa2.Spec.MaxReplicas)
}

// desiredServiceForModel returns a Service object for the Model CR.
func (r *ModelReconciler) desiredHPAForModel(model *llmmodelv1alpha1.Model) *autoscalingv2.HorizontalPodAutoscaler {
	ls := r.desiredLabelSelector(model)
	maxReplicas := cmp.Or(model.Spec.MaxReplicas, 1)
	minReplicas := cmp.Or(model.Spec.MinReplicas, 1)
	minReplicas = max(1, minReplicas)
	maxReplicas = max(maxReplicas, minReplicas)

	targetMemory := int32(80)
	namespace := cmp.Or(model.Spec.TargetNamespace, "default")
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      model.Name,
			Namespace: namespace,
			Labels:    ls,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MaxReplicas: maxReplicas,
			MinReplicas: &minReplicas,
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       model.Name,
			},
			Metrics: []autoscalingv2.MetricSpec{
				{
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: corev1.ResourceMemory,
						Target: autoscalingv2.MetricTarget{
							AverageUtilization: &targetMemory,
							Type:               autoscalingv2.UtilizationMetricType,
						},
					},
					Type: "Resource",
				},
			},
		},
	}

	// Set the Model instance as the owner and controller
	err := ctrl.SetControllerReference(model, hpa, r.Scheme)
	if err != nil {
		log := logf.Log
		log.Error(err, "Failed to create a service reference")
	}
	return hpa
}
