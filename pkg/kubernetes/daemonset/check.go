// Copyright 2016-2021, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package daemonset

import (
	"fmt"

	"github.com/pulumi/cloud-ready-checks/pkg/checker"
	"github.com/pulumi/cloud-ready-checks/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
)

func NewDaemonSetChecker() *checker.StateChecker {
	return checker.NewStateChecker(&checker.StateCheckerArgs{
		Conditions: []checker.Condition{daemonSetScheduled, daemonSetReady},
	})
}

//
// Conditions
//

func daemonSetScheduled(obj interface{}) checker.Result {
	daemonSet := obj.(*appsv1.DaemonSet)
	result := checker.Result{Description: fmt.Sprintf(
		"Waiting for daemonSet %q to be scheduled", kubernetes.FullyQualifiedName(daemonSet))}

	if daemonSet.Status.CurrentNumberScheduled >= 0 {
		result.Ok = true
	}

	return result
}

func daemonSetReady(obj interface{}) checker.Result {
	daemonSet := obj.(*appsv1.DaemonSet)
	result := checker.Result{Description: fmt.Sprintf(
		"Waiting for daemonSet %q to be ready", kubernetes.FullyQualifiedName(daemonSet))}
	if daemonSet.Status.DesiredNumberScheduled > 0 {
		if daemonSet.Status.NumberAvailable == daemonSet.Status.DesiredNumberScheduled {
			result.Ok = true
		}
	}
	return result
}

//
// Helpers
//
