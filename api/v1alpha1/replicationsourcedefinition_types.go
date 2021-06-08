/*
Copyright 2020 The Scribe authors.

This file may be used, at your option, according to either the GNU AGPL 3.0 or
the Apache V2 license.

---
This program is free software: you can redistribute it and/or modify it under
the terms of the GNU Affero General Public License as published by the Free
Software Foundation, either version 3 of the License, or (at your option) any
later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE.  See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along
with this program.  If not, see <https://www.gnu.org/licenses/>.

---
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

//+kubebuilder:validation:Required
package v1alpha1

import (
	"github.com/operator-framework/operator-lib/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReplicationSourceDefinitionSpec struct {
	ReplicationMethod string `json:"replicationMethod"`
	//RcloneConfigSection is the section in rclone_config file to use for the current job.
	RcloneConfigSection *string `json:"rcloneConfigSection,omitempty"`
	// RcloneDestPath is the remote path to sync to.
	RcloneDestPath *string `json:"rcloneDestPath,omitempty"`
	// RcloneConfig is the rclone secret name
	RcloneConfig *string `json:"rcloneConfig,omitempty"`
	// copyMethod describes how a point-in-time (PiT) image of the destination
	// volume should be created.
	CopyMethod CopyMethodType `json:"copyMethod,omitempty"`
}

// ReplicationSourceDefinitionStatus defines the observed state of ReplicationSourceDefinition
type ReplicationSourceDefinitionStatus struct {
	// conditions represent the latest available observations of the
	// destination's state.
	Conditions status.Conditions `json:"conditions,omitempty"`
}

// ReplicationSourceDefinition defines the destination for a replicated volume
//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Namespaced
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Last sync",type="string",format="date-time",JSONPath=`.status.lastSyncTime`
//+kubebuilder:printcolumn:name="Next sync",type="string",format="date-time",JSONPath=`.status.nextSyncTime`
type ReplicationSourceDefinition struct {
	metav1.TypeMeta `json:",inline"`
	//+optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// spec is the desired state of the ReplicationSourceDefinition, including the
	// replication method to use and its configuration.
	Spec ReplicationSourceDefinitionSpec `json:"spec,omitempty"`
	// status is the observed state of the ReplicationSourceDefinition as determined
	// by the controller.
	//+optional
	Status *ReplicationSourceDefinitionStatus `json:"status,omitempty"`
}

// ReplicationSourceDefinitionList contains a list of ReplicationSourceDefinition
//+kubebuilder:object:root=true
type ReplicationSourceDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicationSourceDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReplicationSourceDefinition{}, &ReplicationSourceDefinitionList{})
}
