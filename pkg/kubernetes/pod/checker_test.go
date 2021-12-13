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
	"testing"

	"github.com/pulumi/cloud-ready-checks/internal"
	"github.com/pulumi/cloud-ready-checks/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
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
			jsonBytes, err := internal.TestStates.ReadFile(tt.testStatePath)
			assert.NoError(t, err)

			state := kubernetes.MustLoadState(jsonBytes)
			pod := corev1.Pod{}
			err = kubernetes.BuiltInScheme.Convert(state, &pod, nil)
			assert.NoError(t, err)
			if got := podInitialized(&pod); got.Ok != tt.want {
				t.Errorf("podInitialized() = %v, want %v", got.Ok, tt.want)
			}
		})
	}
}

//
// Test Pod State Checker using recorded events.
//

//func Test_Pod_Checker(t *testing.T) {
//	podInit, err := internal.TestStates.ReadFile("states/kubernetes/pod/initialized.json")
//	assert.NoError(t, err)
//	fmt.Println(string(podInit))
//	//workflow := func(name string) string {
//	//	return workflowPath("pod", name)
//	//}
//	//const (
//	//	added                                  = "added"
//	//	containerTerminatedError               = "containerTerminatedError"
//	//	containerTerminatedSuccess             = "containerTerminatedSuccess"
//	//	containerTerminatedSuccessRestartNever = "containerTerminatedSuccessRestartNever"
//	//	createSuccess                          = "createSuccess"
//	//	imagePullError                         = "imagePullError"
//	//	imagePullErrorResolved                 = "imagePullErrorResolved"
//	//	scheduled                              = "scheduled"
//	//	unready                                = "unready"
//	//	unscheduled                            = "unscheduled"
//	//)
//	//
//	tests := []struct {
//		name           string
//		recordingPaths []string
//		expectReady bool
//	}{
//		{
//			name:           "Pod added but not ready",
//			recordingPaths: []string{workflow(added)},
//			expectReady:    false,
//		},
//		{
//			name:           "Pod scheduled but not ready",
//			recordingPaths: []string{workflow(scheduled)},
//			expectReady:    false,
//		},
//		{
//			name:           "Pod create success",
//			recordingPaths: []string{workflow(createSuccess)},
//			expectReady:    true,
//		},
//		{
//			name:           "Pod image pull error",
//			recordingPaths: []string{workflow(imagePullError)},
//			expectReady:    false,
//		},
//		{
//			name:           "Pod create success after image pull failure resolved",
//			recordingPaths: []string{workflow(imagePullError), workflow(imagePullErrorResolved)},
//			expectReady:    true,
//		},
//		{
//			name:           "Pod unscheduled",
//			recordingPaths: []string{workflow(unscheduled)},
//			expectReady:    false,
//		},
//		{
//			name:           "Pod unready",
//			recordingPaths: []string{workflow(unready)},
//			expectReady:    false,
//		},
//		{
//			name:           "Pod container terminated with error",
//			recordingPaths: []string{workflow(containerTerminatedError)},
//			expectReady:    false,
//		},
//		{
//			name:           "Pod container terminated successfully",
//			recordingPaths: []string{workflow(containerTerminatedSuccess)},
//			expectReady:    false,
//		},
//		{
//			name:           "Pod container terminated successfully with restartPolicy: Never",
//			recordingPaths: []string{workflow(containerTerminatedSuccessRestartNever)},
//			expectReady:    true,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			checker := NewPodChecker()
//
//			ready, messages := mustCheckIfRecordingsReady(tt.recordingPaths, checker)
//			if ready != tt.expectReady {
//				t.Errorf("Ready() = %t, want %t\nMessages: %s", ready, tt.expectReady, messages)
//			}
//		})
//	}
//}
