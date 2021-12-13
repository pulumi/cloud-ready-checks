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
	"github.com/pulumi/cloud-ready-checks/pkg/checker"
	"testing"

	"github.com/pulumi/cloud-ready-checks/internal"
	"github.com/pulumi/cloud-ready-checks/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

//
// Test Conditions
//

func Test_podInitialized(t *testing.T) {
	tests := []struct {
		name          string
		testStatePath string
		want          bool
	}{
		{
			"Pod initialized",
			"states/kubernetes/pod/initialized.json",
			true,
		},
		{
			"Pod uninitialized",
			"states/kubernetes/pod/uninitialized.json",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := loadPod(t, tt.testStatePath)
			if got := podInitialized(pod); got.Ok != tt.want {
				t.Errorf("podInitialized() = %v, want %v", got.Ok, tt.want)
			}
		})
	}
}

func Test_podReady(t *testing.T) {
	tests := []struct {
		name          string
		testStatePath string
		want          bool
	}{
		{
			"Pod ready",
			"states/kubernetes/pod/ready.json",
			true,
		},
		{
			"Pod succeeded",
			"states/kubernetes/pod/succeeded.json",
			true,
		},
		{
			"Pod unready",
			"states/kubernetes/pod/initialized.json",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := loadPod(t, tt.testStatePath)
			if got := podReady(pod); got.Ok != tt.want {
				t.Errorf("podReady() = %v, want %v", got.Ok, tt.want)
			}
		})
	}
}

func Test_podScheduled(t *testing.T) {
	tests := []struct {
		name          string
		testStatePath string
		want          bool
	}{
		{
			"Pod scheduled",
			"states/kubernetes/pod/scheduled.json",
			true,
		},
		{
			"Pod unscheduled",
			"states/kubernetes/pod/unscheduled.json",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := loadPod(t, tt.testStatePath)
			if got := podScheduled(pod); got.Ok != tt.want {
				t.Errorf("podScheduled() = %v, want %v", got.Ok, tt.want)
			}
		})
	}
}

//
// Test Pod State Checker using recorded events.
//

func Test_Pod_Checker(t *testing.T) {
	workflow := func(name string) string {
		return workflowPath(name)
	}
	const (
		added                                  = "added"
		containerTerminatedError               = "containerTerminatedError"
		containerTerminatedSuccess             = "containerTerminatedSuccess"
		containerTerminatedSuccessRestartNever = "containerTerminatedSuccessRestartNever"
		createSuccess                          = "createSuccess"
		imagePullError                         = "imagePullError"
		imagePullErrorResolved                 = "imagePullErrorResolved"
		scheduled                              = "scheduled"
		unready                                = "unready"
		unscheduled                            = "unscheduled"
	)

	tests := []struct {
		name          string
		workflowPaths []string
		expectReady   bool
	}{
		{
			name:          "Pod added but not ready",
			workflowPaths: []string{workflow(added)},
			expectReady:   false,
		},
		{
			name:          "Pod scheduled but not ready",
			workflowPaths: []string{workflow(scheduled)},
			expectReady:   false,
		},
		{
			name:          "Pod create success",
			workflowPaths: []string{workflow(createSuccess)},
			expectReady:   true,
		},
		{
			name:          "Pod image pull error",
			workflowPaths: []string{workflow(imagePullError)},
			expectReady:   false,
		},
		{
			name:          "Pod create success after image pull failure resolved",
			workflowPaths: []string{workflow(imagePullError), workflow(imagePullErrorResolved)},
			expectReady:   true,
		},
		{
			name:          "Pod unscheduled",
			workflowPaths: []string{workflow(unscheduled)},
			expectReady:   false,
		},
		{
			name:          "Pod unready",
			workflowPaths: []string{workflow(unready)},
			expectReady:   false,
		},
		{
			name:          "Pod container terminated with error",
			workflowPaths: []string{workflow(containerTerminatedError)},
			expectReady:   false,
		},
		{
			name:          "Pod container terminated successfully",
			workflowPaths: []string{workflow(containerTerminatedSuccess)},
			expectReady:   false,
		},
		{
			name:          "Pod container terminated successfully with restartPolicy: Never",
			workflowPaths: []string{workflow(containerTerminatedSuccessRestartNever)},
			expectReady:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podChecker := NewPodChecker()

			ready := false
			var details checker.Results
			podStates := loadWorkflows(t, tt.workflowPaths...)
			for _, podState := range podStates {
				ready, details = podChecker.ReadyDetails(podState)
				if ready {
					break
				}
			}
			fmt.Printf("Expect Ready() = %t\n", tt.expectReady)
			fmt.Println(details)
			if ready != tt.expectReady {
				t.Errorf("Ready() = %t, want %t", ready, tt.expectReady)
			}
		})
	}
}

//
// Helpers
//

func loadPod(t *testing.T, statePath string) *corev1.Pod {
	jsonBytes, err := internal.TestStates.ReadFile(statePath)
	require.NoError(t, err)

	state := kubernetes.MustLoadState(jsonBytes)
	pod := corev1.Pod{}
	err = kubernetes.BuiltInScheme.Convert(state, &pod, nil)
	require.NoError(t, err)

	return &pod
}

func loadWorkflows(t *testing.T, workflowPaths ...string) []*corev1.Pod {
	var pods []*corev1.Pod
	for _, workflowPath := range workflowPaths {
		jsonBytes, err := internal.TestStates.ReadFile(workflowPath)
		require.NoError(t, err)

		states := kubernetes.MustLoadWorkflow(jsonBytes)
		for _, state := range states {
			pod := corev1.Pod{}
			err = kubernetes.BuiltInScheme.Convert(state, &pod, nil)
			require.NoError(t, err)
			pods = append(pods, &pod)
		}
	}

	return pods
}

func workflowPath(name string) string {
	return fmt.Sprintf("workflows/kubernetes/pod/%s.json", name)
}
