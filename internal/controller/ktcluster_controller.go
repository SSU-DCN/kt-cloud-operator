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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
	infrastructurev1beta1 "dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
)

// KTClusterReconciler reconciles a KTCluster object
type KTClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.dcnlab.ssu.ac.kr,resources=ktclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.dcnlab.ssu.ac.kr,resources=ktclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.dcnlab.ssu.ac.kr,resources=ktclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KTCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *KTClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.V(1).Info("KTCluster Reconcile", "ktCluster", req)

	// Fetch the KTCluster instance
	ktcluster := &v1beta1.KTCluster{}
	if err := r.Get(ctx, req.NamespacedName, ktcluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("KTCluster resource not found. Ignoring since it must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get KTCluster resource")
		return ctrl.Result{}, err
	}

	// Fetch child resources
	foundKTMachineTemplateCP, err := r.fetchMachineTemplate(ctx, ktcluster, "-control-plane", req)
	if err != nil {
		logger.Error(err, "Failed to find control-plane machine template")
		return ctrl.Result{}, nil // Or return an error if this is critical
	}

	foundKTMachineTemplateMD, err := r.fetchMachineTemplate(ctx, ktcluster, "-md-0", req)
	if err != nil {
		logger.Error(err, "Failed to find -md-0 machine template")
		return ctrl.Result{}, nil // Or return an error if this is critical
	}

	// Check if any required machine template is missing
	if foundKTMachineTemplateCP == nil || foundKTMachineTemplateMD == nil {
		logger.Info("One or more machine templates are missing. Requeuing...")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	logger.Info("Successfully added owner references", "KTCluster.Name", ktcluster.Name)
	return ctrl.Result{}, nil
}

func (r *KTClusterReconciler) fetchMachineTemplate(ctx context.Context, ktcluster *v1beta1.KTCluster, suffix string, req ctrl.Request) (*v1beta1.KTMachineTemplate, error) {
	logger := log.FromContext(ctx)
	templateName := string(ktcluster.Name + suffix)
	machineTemplate := &v1beta1.KTMachineTemplate{}
	err := r.Get(ctx, types.NamespacedName{Name: templateName, Namespace: ktcluster.Namespace}, machineTemplate)

	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("MachineTemplate not found for "+templateName, "Name", templateName, "Namespace", ktcluster.Namespace)
			return nil, err
		}
		return nil, err
	}

	// Add owner references
	if err := r.ktClusterForMachineTemplate(ktcluster, machineTemplate, ctx, req); err != nil {
		logger.Error(err, "Failed to add owner reference to control-plane machine template")
	}

	return machineTemplate, nil
}

func (r *KTClusterReconciler) ktClusterForMachineTemplate(ktCluster *v1beta1.KTCluster, ktMachineTemplate *v1beta1.KTMachineTemplate, ctx context.Context, req ctrl.Request) error {
	logger := log.FromContext(ctx)
	logger.Info("adding owner ref for machine ", "KTMachineTemplate.Namespace ", ktMachineTemplate.Namespace, " KTMachineTemplate.Name ", ktMachineTemplate.Name)

	// Set the ownerRef for the KTCluster
	// will be deleted when the Cluster CR is deleted.
	// controllerutil.SetControllerReference(ktCluster, ktMachineTemplate, r.Scheme)
	if err := controllerutil.SetControllerReference(ktCluster, ktMachineTemplate, r.Scheme); err != nil {
		logger.Error(err, "Failed to set ktmachine template owner reference")
		return err
	}

	if err := r.Client.Update(ctx, ktMachineTemplate); err != nil {
		logger.Error(err, "Can't update for ktmachine template owner reference")
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KTClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.KTCluster{}).
		Owns(&v1beta1.KTMachineTemplate{}).
		Named("ktcluster").
		Complete(r)
}
