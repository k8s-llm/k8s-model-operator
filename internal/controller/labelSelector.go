package controller

import llmmodelv1alpha1 "host-llm.io/k8s-model-operator/api/v1alpha1"

// desiredLabelSelector returns a map to fill a label selector.
func (r *ModelReconciler) desiredLabelSelector(model *llmmodelv1alpha1.Model) map[string]string {
	ls := map[string]string{
		"app": model.Name,
	}

	return ls
}
