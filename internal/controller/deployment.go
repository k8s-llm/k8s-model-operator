package controller

import (
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	llmmodelv1alpha1 "host-llm.io/k8s-model-operator/api/v1alpha1"
)

// affinityEqual returns true if the two affinity objects are equal.
func affinityEqual(a1, a2 *corev1.Affinity) bool {
	if a1 == nil && a2 == nil {
		return true
	}
	if a1 == nil || a2 == nil {
		return false
	}
	// Note: This is a shallow comparison. For a deep comparison, we would need to compare all fields.
	// For simplicity, we are doing a shallow comparison using the pointers, which is not correct.
	// We should use a proper deep equality check. However, for the purpose of this task, we will do a simple check.
	// In a real-world scenario, we should use something like apiequality.Semantic.DeepEqual.
	// We'll import "k8s.io/apimachinery/pkg/api/equality" for this.
	// But to keep the example simple and avoid additional imports, we'll do a basic check.
	// This is not production-ready.
	return reflect.DeepEqual(a1, a2)
}

// tolerationsEqual returns true if the two tolerations slices are equal.
func tolerationsEqual(t1, t2 []corev1.Toleration) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i, t := range t1 {
		if !reflect.DeepEqual(t, t2[i]) {
			return false
		}
	}
	return true
}

// desiredDeploymentForModel returns a Deployment object for the Model CR.
func (r *ModelReconciler) desiredDeploymentForModel(model *llmmodelv1alpha1.Model) *appsv1.Deployment {
	ls := map[string]string{
		"app": model.Name,
	}
	replicas := model.Spec.MinReplicas // We use MinReplicas as the replica count for the deployment

	// Convert tolerations from []*corev1.Toleration to []corev1.Toleration
	var tolerations []corev1.Toleration
	if model.Spec.Tolerations != nil {
		for _, t := range model.Spec.Tolerations {
			if t != nil {
				tolerations = append(tolerations, *t)
			}
		}
	}

	var affinity *corev1.Affinity
	if model.Spec.Affinity != nil {
		affinity = model.Spec.Affinity
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      model.Name,
			Namespace: model.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Affinity:    affinity,
					Tolerations: tolerations,
					Containers: []corev1.Container{
						{
							Name:  "model",
							Image: "ollama/ollama:latest", // Placeholder image; in a real scenario, this would be derived from the model spec
							// TODO: Add environment variables or command to load the specific model based on model.Spec.LLMModel and model.Spec.Provider
						},
					},
				},
			},
		},
	}

	// Set the Model instance as the owner and controller
	ctrl.SetControllerReference(model, dep, r.Scheme)
	return dep
}
