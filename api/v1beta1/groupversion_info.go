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

// Package v1beta1 contains API Schema definitions for the infrastructure v1beta1 API group.
// +kubebuilder:object:generate=true
// +groupName=infrastructure.dcnlab.ssu.ac.kr
package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// GroupName is the group name use in this package.
const GroupName = "infrastructure.dcnlab.ssu.ac.kr"

// // SchemeGroupVersion is group version used to register these objects.
// var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1beta1"}

// // Resource takes an unqualified resource and returns a Group qualified GroupResource.
// func Resource(resource string) schema.GroupResource {
// 	return SchemeGroupVersion.WithResource(resource).GroupResource()
// }

// var (
// 	// schemeBuilder is used to add go types to the GroupVersionKind scheme.
// 	schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// 	// AddToScheme adds the types in this group-version to the given scheme.
// 	AddToScheme = schemeBuilder.AddToScheme

// 	objectTypes = []runtime.Object{}
// )

//	func addKnownTypes(scheme *runtime.Scheme) error {
//		scheme.AddKnownTypes(SchemeGroupVersion, objectTypes...)
//		metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
//		return nil
//	}
var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "infrastructure.dcnlab.ssu.ac.kr", Version: "v1beta1"}

	SchemeGroupVersion = GroupVersion

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}
