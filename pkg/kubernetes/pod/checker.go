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

package pod

import (
	"fmt"
	"strings"

	"github.com/pulumi/cloud-ready-checks/pkg/common"
	"github.com/pulumi/cloud-ready-checks/pkg/common/logging"
	"github.com/pulumi/cloud-ready-checks/pkg/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	corev1 "k8s.io/api/core/v1"
)

func NewPodChecker() *common.StateChecker {
	return common.NewStateChecker(&common.StateCheckerArgs{
		ReadyMessage: cmdutil.EmojiOr("✅ Pod ready", "Pod ready"),
		Conditions:   []common.Condition{podScheduled, podInitialized, podReady},
	})
}

//
// Conditions
//

func podScheduled(obj interface{}) common.Result {
	pod := toPod(obj)
	result := common.Result{Description: fmt.Sprintf(
		"Waiting for Pod %q to be scheduled", kubernetes.FullyQualifiedName(pod))}

	if condition, found := filterConditions(pod.Status.Conditions, corev1.PodScheduled); found {
		switch condition.Status {
		case corev1.ConditionTrue:
			result.Ok = true
		default:
			msg := statusFromCondition(condition)
			if len(msg) > 0 {
				result.Message = logging.StatusMessage(msg)
			}
		}
	}

	return result
}

func podInitialized(obj interface{}) common.Result {
	pod := toPod(obj)
	result := common.Result{Description: fmt.Sprintf(
		"Waiting for Pod %q to be initialized", kubernetes.FullyQualifiedName(pod))}

	if condition, found := filterConditions(pod.Status.Conditions, corev1.PodInitialized); found {
		switch condition.Status {
		case corev1.ConditionTrue:
			result.Ok = true
		default:
			var errs []string
			for _, status := range pod.Status.ContainerStatuses {
				if ok, containerErrs := hasContainerStatusErrors(status); !ok {
					errs = append(errs, containerErrs...)
				}
			}
			result.Message = logging.WarningMessage(podError(condition, errs, kubernetes.FullyQualifiedName(pod)))
		}
	}

	return result
}

func podReady(obj interface{}) common.Result {
	pod := toPod(obj)
	result := common.Result{Description: fmt.Sprintf(
		"Waiting for Pod %q to be ready", kubernetes.FullyQualifiedName(pod))}

	if condition, found := filterConditions(pod.Status.Conditions, corev1.PodReady); found {
		switch condition.Status {
		case corev1.ConditionTrue:
			result.Ok = true
		default:
			switch pod.Status.Phase {
			case corev1.PodSucceeded: // If the Pod has terminated, but .status.phase is "Succeeded", consider it Ready.
				result.Ok = true
			default:
				errs := collectContainerStatusErrors(pod.Status.ContainerStatuses)
				result.Message = logging.WarningMessage(podError(condition, errs, kubernetes.FullyQualifiedName(pod)))
			}
		}
	}

	return result
}

//
// Helpers
//

func toPod(obj interface{}) *corev1.Pod {
	// TODO: Probably need more robust logic here
	return obj.(*corev1.Pod)
}

func collectContainerStatusErrors(statuses []corev1.ContainerStatus) []string {
	var errs []string
	for _, status := range statuses {
		if hasErr, containerErrs := hasContainerStatusErrors(status); hasErr {
			errs = append(errs, containerErrs...)
		}
	}

	return errs
}

func hasContainerStatusErrors(status corev1.ContainerStatus) (bool, []string) {
	if status.Ready {
		return false, nil
	}

	var errs []string
	if hasErr, err := hasContainerWaitingError(status); hasErr {
		errs = append(errs, err)
	}
	if hasErr, err := hasContainerTerminatedError(status); hasErr {
		errs = append(errs, err)
	}

	return len(errs) > 0, errs
}

func hasContainerWaitingError(status corev1.ContainerStatus) (bool, string) {
	state := status.State.Waiting
	if state == nil {
		return false, ""
	}

	// Return false if the container is creating.
	if state.Reason == "ContainerCreating" {
		return false, ""
	}

	msg := fmt.Sprintf("[%s] %s", state.Reason, trimImagePullMsg(state.Message))
	return true, msg
}

func hasContainerTerminatedError(status corev1.ContainerStatus) (bool, string) {
	state := status.State.Terminated
	if state == nil {
		return false, ""
	}

	// Return false if no reason given.
	if len(state.Reason) == 0 {
		return false, ""
	}

	if len(state.Message) > 0 {
		msg := fmt.Sprintf("[%s] %s", state.Reason, trimImagePullMsg(state.Message))
		return true, msg
	}
	return true, fmt.Sprintf("Container %q completed with exit code %d", status.Name, state.ExitCode)
}

// trimImagePullMsg trims unhelpful error from ImagePullError status messages.
func trimImagePullMsg(msg string) string {
	msg = strings.TrimPrefix(msg, "rpc error: code = Unknown desc = Error response from daemon: ")
	msg = strings.TrimSuffix(msg, ": manifest unknown")

	return msg
}

func statusFromCondition(condition *corev1.PodCondition) string {
	if condition.Reason != "" && condition.Message != "" {
		return condition.Message
	}

	return ""
}

func filterConditions(conditions []corev1.PodCondition, desired corev1.PodConditionType) (*corev1.PodCondition, bool) {
	for _, condition := range conditions {
		if condition.Type == desired {
			return &condition, true
		}
	}

	return nil, false
}

func podError(condition *corev1.PodCondition, errs []string, name string) string {
	errMsg := fmt.Sprintf("[Pod %s]: ", name)
	if len(condition.Reason) > 0 && len(condition.Message) > 0 {
		errMsg += condition.Message
	}

	for _, err := range errs {
		errMsg += fmt.Sprintf(" -- %s", err)
	}

	return errMsg
}
