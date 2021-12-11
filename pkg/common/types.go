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

// TODO: probably refactor package names
package common

import (
	"fmt"

	"github.com/pulumi/cloud-ready-checks/pkg/common/logging"
)

// Result specifies the result of a Condition applied to an input object.
type Result struct {
	Ok          bool            // True if the Condition is true, false otherwise.
	Description string          // A human-readable description of the associated Condition.
	Message     logging.Message // The message to be logged after evaluating the Condition.
}

func (r Result) String() string {
	s := fmt.Sprintf("[%t] %s", r.Ok, r.Description)

	if !r.Message.Empty() {
		s = fmt.Sprintf("%s -- %s", s, r.Message)
	}

	return s
}

// Condition is a function that checks a state and returns a Result.
type Condition func(state interface{}) Result

// StateChecker holds the data required to generically implement await logic.
type StateChecker struct {
	conditions []Condition // Conditions that must be true for the state to be Ready.
}

type StateCheckerArgs struct {
	Conditions []Condition // Conditions that must be true for the state to be Ready.
}

func NewStateChecker(args *StateCheckerArgs) *StateChecker {
	return &StateChecker{
		conditions: args.Conditions,
	}
}

func (s *StateChecker) Ready(state interface{}) bool {
	ok, _ := s.readyDetails(state)
	return ok
}

func (s *StateChecker) ReadyStatus(state interface{}) (bool, Result) {
	ok, results := s.readyDetails(state)
	return ok, results[len(results)-1]
}

func (s *StateChecker) ReadyDetails(state interface{}) (bool, []Result) {
	return s.readyDetails(state)
}

func (s *StateChecker) readyDetails(state interface{}) (bool, []Result) {
	var results []Result

	for _, condition := range s.conditions {
		result := condition(state)
		results = append(results, result)
		if !result.Ok {
			return false, results
		}
	}

	return true, results
}
