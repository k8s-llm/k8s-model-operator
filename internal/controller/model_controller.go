/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"

	llmmodelv1alpha1 "host-llm.io/k8s-model-operator/api/v1alpha1"
)

// ModelReconciler reconciles a Model object
type ModelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Definitions to manage status conditions
const (
	// typeAvailableModel represents the status of the Model reconciliation
	typeAvailableModel = "Available"
	// typeProgressingModel represents the status used when the Model is being reconciled
	typeProgressingModel = "Progressing"
	// typeDegradedModel represents the status used when the Model has encountered an error
	typeDegradedModel = "Degraded"
)

// +kubebuilder:rbac:groups=llmmodel.host-llm.io,resources=models,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=llmmodel.host-llm.io,resources=models/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=llmmodel.host-llm.io,resources=models/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployment;service,verbs=create;get;list;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscaler,verbs=create;get;list;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Model object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *ModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// TODO(user): your logic here
	var model llmmodelv1alpha1.Model

	// Reading resource to see if it exists
	if err := r.Get(ctx, req.NamespacedName, &model); err != nil {
		if apierrors.IsNotFound(err) {
			// If the custom resource is not found then it usually means that it was deleted or not created
			// In this way, we will stop the reconciliation
			log.Info("Model resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Model resource")
		return ctrl.Result{}, err
	}

	// Initialize status conditions if not yet present
	if len(model.Status.Conditions) == 0 {
		meta.SetStatusCondition(&model.Status.Conditions, metav1.Condition{
			Type:    typeProgressingModel,
			Status:  metav1.ConditionUnknown,
			Reason:  "Reconciling",
			Message: "Starting reconciliation",
		})
		if err := r.Status().Update(ctx, &model); err != nil {
			log.Error(err, "Failed to update Model status")
			return ctrl.Result{}, err
		}

		/*
			After updating the status, we re-fetch the Model to ensure we are working with
			the latest version of the object from the API server.
		*/
		if err := r.Get(ctx, req.NamespacedName, &model); err != nil {
			log.Error(err, "Failed to re-fetch Model")
			return ctrl.Result{}, err
		}
	}

	// Define the desired Deployment object
	desiredDeployment := r.desiredDeploymentForModel(&model)

	// Check if the Deployment already exists
	var foundDeployment appsv1.Deployment
	err := r.Get(ctx, types.NamespacedName{Name: desiredDeployment.Name, Namespace: desiredDeployment.Namespace}, &foundDeployment)
	if err != nil && apierrors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", desiredDeployment.Namespace, "Deployment.Name", desiredDeployment.Name)
		err = r.Create(ctx, desiredDeployment)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", desiredDeployment.Namespace, "Deployment.Name", desiredDeployment.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the Deployment size is the same as the spec
	// We only update if the replica count, affinity, or tolerations have changed
	// Note: We are not updating the container image or other fields in this example.
	// In a more complete implementation, we would compare the entire spec.
	needUpdate := false
	if *foundDeployment.Spec.Replicas != *desiredDeployment.Spec.Replicas {
		needUpdate = true
		log.Info("Deployment replica count is out of sync", "Desired", *desiredDeployment.Spec.Replicas, "Current", *foundDeployment.Spec.Replicas)
	}
	if !affinityEqual(foundDeployment.Spec.Template.Spec.Affinity, desiredDeployment.Spec.Template.Spec.Affinity) {
		needUpdate = true
		log.Info("Deployment affinity is out of sync")
	}
	if !tolerationsEqual(foundDeployment.Spec.Template.Spec.Tolerations, desiredDeployment.Spec.Template.Spec.Tolerations) {
		needUpdate = true
		log.Info("Deployment tolerations are out of sync")
	}
	if needUpdate {
		log.Info("Updating Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
		// We update the replica count, affinity, and tolerations from the desired deployment
		foundDeployment.Spec.Replicas = desiredDeployment.Spec.Replicas
		foundDeployment.Spec.Template.Spec.Affinity = desiredDeployment.Spec.Template.Spec.Affinity
		foundDeployment.Spec.Template.Spec.Tolerations = desiredDeployment.Spec.Template.Spec.Tolerations
		err = r.Update(ctx, &foundDeployment)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	// Update the Model status based on the Deployment status
	// We assume the Deployment is available if it has at least one ready replica equal to the desired replicas
	if *foundDeployment.Spec.Replicas == foundDeployment.Status.ReadyReplicas {
		meta.SetStatusCondition(&model.Status.Conditions, metav1.Condition{
			Type:    typeAvailableModel,
			Status:  metav1.ConditionTrue,
			Reason:  "DeploymentAvailable",
			Message: "The Model deployment is available",
		})
	} else {
		meta.SetStatusCondition(&model.Status.Conditions, metav1.Condition{
			Type:    typeAvailableModel,
			Status:  metav1.ConditionFalse,
			Reason:  "DeploymentNotAvailable",
			Message: "The Model deployment is not yet available",
		})
		meta.SetStatusCondition(&model.Status.Conditions, metav1.Condition{
			Type:    typeProgressingModel,
			Status:  metav1.ConditionTrue,
			Reason:  "DeploymentProgressing",
			Message: "The Model deployment is progressing",
		})
	}

	// Update the Model status
	err = r.Status().Update(ctx, &model)
	if err != nil {
		log.Error(err, "Failed to update Model status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&llmmodelv1alpha1.Model{}).
		Owns(&appsv1.Deployment{}).
		Named("model").
		Complete(r)
}
