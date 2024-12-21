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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
	infrastructurev1beta1 "dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
	"dcnlab.ssu.ac.kr/kt-cloud-operator/cmd/httpapi"
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

	//trigger to create machine on KTCloud by calling API
	if ktMachine.Status.ID == "" {
		logger.Info("Machine has no ID in the status field, create it")
		//first get the token associated for the cluster and find token
		subjectToken, err := r.getSubjectToken(ctx, ktMachine, req)
		if err != nil {
			logger.Error(err, "Failed to find KTSubject token matching cluster")
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}

		if subjectToken == "" {
			logger.Error(err, "We have to reconcile again to check the Subject token")
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}

		err = httpapi.CreateVM(ktMachine, subjectToken)
		if err != nil {
			logger.Error(err, "Failed to create VM on KT Cloud during API Call")
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}

		//use the response to from the api and update the machine
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else {
		logger.Info("Machine already created and has ID")
	}

	return ctrl.Result{}, nil
}

func (r *KTMachineReconciler) getSubjectToken(ctx context.Context, ktMachine *infrastructurev1beta1.KTMachine, req ctrl.Request) (string, error) {

	logger := log.FromContext(ctx, "LogFrom", "Machine")

	ktMachineDeploymentList := &v1beta1.MachineDeploymentList{}
	err := r.List(ctx, ktMachineDeploymentList, client.InNamespace(ktMachine.Namespace))
	if err != nil {
		logger.Error(err, "failed to list MachineDeployments for this machine")
		return "", err
	}

	// Filter by ownerReferences
	var ownerMachineDeployment v1beta1.MachineDeployment
	for _, machineDeployment := range ktMachineDeploymentList.Items {
		for _, ref := range ktMachine.OwnerReferences {
			if ref.UID == machineDeployment.UID {
				ownerMachineDeployment = machineDeployment
				logger.Info("Found owned MachineDeployment", "name", machineDeployment.Name)
				break
			}
		}
	}

	// we found matching machine deployment owner
	// we have to find the cluster from this
	// KTMachineTemplate is owned by KTCluster and KTMachineTemplate.Name = MachineDeployment.Name, KTMachineTemplate.NameSpace = MachineDeployment.NameSpace
	// therefore, use MachineDeployment to MachineTemplate to find Cluster then token for the cluster
	if ownerMachineDeployment.UID != "" {
		ktMachineTemplate := &v1beta1.KTMachineTemplate{}
		err := r.Get(ctx, types.NamespacedName{Name: ownerMachineDeployment.Name, Namespace: ownerMachineDeployment.Namespace}, ktMachineTemplate)

		if err != nil {
			if apierrors.IsNotFound(err) {
				logger.Error(err, "KTMachineTemplate not found no need to proceed for finding SubjectToken To Auth API", "Name", ownerMachineDeployment.Name, "Namespace", ownerMachineDeployment.Namespace)
				return "", err
			}
			return "", err
		}

		var clusterUID types.UID

		for _, ref := range ktMachineTemplate.OwnerReferences {
			if ref.Kind == "KTCluster" {
				clusterUID = ref.UID
				break

			}
		}

		ktSubjectTokenList := &v1beta1.KTSubjectTokenList{}
		if err := r.List(ctx, ktSubjectTokenList, client.InNamespace(ownerMachineDeployment.Namespace)); err != nil {
			logger.Error(err, "Failed to list KTSubjectTokens")
			return "", err
		}

		for _, token := range ktSubjectTokenList.Items {
			for _, ref := range token.OwnerReferences {
				if ref.Kind == "KTCluster" && ref.UID == clusterUID {
					logger.Info("Found SubjectToken for KTCluster", "ClusterUID", clusterUID)
					return token.Spec.SubjectToken, nil
				}
			}
		}

		logger.Info("No SubjectToken found for KTCluster", "ClusterUID", clusterUID)
		return "", nil

	}

	return "", nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KTMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.KTMachine{}).
		Named("ktmachine").
		Complete(r)
}
