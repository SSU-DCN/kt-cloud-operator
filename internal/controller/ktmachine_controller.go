/*
Copyright 2024.

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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
	infrastructurev1beta1 "dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
)

// KTMachineReconciler reconciles a KTMachine object
type KTMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.dcnlab.ssu.ac.kr,resources=ktmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.dcnlab.ssu.ac.kr,resources=ktmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.dcnlab.ssu.ac.kr,resources=ktmachines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KTMachine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *KTMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx, "LogFrom", "KTMachine")
	logger.V(1).Info("KTMachine Reconcile", "KTMachine", req)

	ktMachine := &v1beta1.KTMachine{}
	if err := r.Get(ctx, req.NamespacedName, ktMachine); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("KTMachine resource not found. Ignoring since it must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get KTMachine resource")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KTMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.KTMachine{}).
		Named("ktmachine").
		Complete(r)
}
