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
	"testing"

	"github.com/pulumi/cloud-ready-checks/internal"
	"github.com/pulumi/cloud-ready-checks/pkg/checker"
	"github.com/pulumi/cloud-ready-checks/pkg/kubernetes/test"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
)

//
// Test Conditions
//

func Test_daemonSetReady(t *testing.T) {
	tests := []struct {
		name          string
		testStatePath string
		want          bool
	}{
		{
			"daemonSet ready",
			"states/kubernetes/daemonSet/ready.json",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSet := loaddaemonSet(t, tt.testStatePath)
			if got := daemonSetReady(daemonSet); got.Ok != tt.want {
				t.Errorf("daemonSetReady() = %v, want %v", got.Ok, tt.want)
			}
		})
	}
}

func Test_daemonSetScheduled(t *testing.T) {
	tests := []struct {
		name          string
		testStatePath string
		want          bool
	}{
		{
			"daemonSet scheduled",
			"states/kubernetes/daemonSet/scheduled.json",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSet := loaddaemonSet(t, tt.testStatePath)
			if got := daemonSetScheduled(daemonSet); got.Ok != tt.want {
				t.Errorf("daemonSetScheduled() = %v, want %v", got.Ok, tt.want)
			}
		})
	}
}

//
// Test daemonSet State Checker using recorded events.
//

func Test_daemonSet_Checker(t *testing.T) {
	workflow := func(name string) string {
		return workflowPath(name)
	}
	const (
		scheduled = "scheduled"
	)

	tests := []struct {
		name          string
		workflowPaths []string
		expectReady   bool
	}{
		{
			name:          "daemonSet scheduled but not ready",
			workflowPaths: []string{workflow(scheduled)},
			expectReady:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daemonSetChecker := NewDaemonSetChecker()

			ready := false
			var details checker.Results
			daemonSetStates := loadWorkflows(t, tt.workflowPaths...)
			for _, daemonSetState := range daemonSetStates {
				ready, details = daemonSetChecker.ReadyDetails(daemonSetState)
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

func loaddaemonSet(t *testing.T, statePath string) *appsv1.DaemonSet {
	jsonBytes, err := internal.TestStates.ReadFile(statePath)
	require.NoError(t, err)

	state := test.MustLoadState(jsonBytes)
	daemonSet := appsv1.DaemonSet{}
	err = test.BuiltInScheme.Convert(state, &daemonSet, nil)
	require.NoError(t, err)

	return &daemonSet
}

func loadWorkflows(t *testing.T, workflowPaths ...string) []*appsv1.DaemonSet {
	var daemonSets []*appsv1.DaemonSet
	for _, workflowPath := range workflowPaths {
		jsonBytes, err := internal.TestStates.ReadFile(workflowPath)
		require.NoError(t, err)

		states := test.MustLoadWorkflow(jsonBytes)
		for _, state := range states {
			daemonSet := appsv1.DaemonSet{}
			err = test.BuiltInScheme.Convert(state, &daemonSet, nil)
			require.NoError(t, err)
			daemonSets = append(daemonSets, &daemonSet)
		}
	}

	return daemonSets
}

func workflowPath(name string) string {
	return fmt.Sprintf("workflows/kubernetes/daemonSet/%s.json", name)
}
