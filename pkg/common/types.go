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
	ready      bool        // True if all the conditions evaluated to true on the most recent Update.
	conditions []Condition // Conditions that must be true for the state to be Ready.
	readyMsg   string      // Status message to show if the state is Ready.
}

type StateCheckerArgs struct {
	Conditions   []Condition // Conditions that must be true for the state to be Ready.
	ReadyMessage string      // Status message to show if the state is Ready.
}

func NewStateChecker(args *StateCheckerArgs) *StateChecker {
	return &StateChecker{
		ready:      false,
		conditions: args.Conditions,
		readyMsg:   args.ReadyMessage,
	}
}

// Ready is true if all the Conditions associated with this checker are true. Ready will always return false prior
// to running Update.
func (s *StateChecker) Ready() bool {
	return s.ready
}

// Update runs the conditions associated with the StateChecker against the provided object. Each condition produces
// a status message that is appended to the returned list of Messages. Iff all of the Conditions are true, the ready
// status is set to true, otherwise, the ready condition is set to false.
func (s *StateChecker) Update(state interface{}) logging.Messages {
	s.ready = false

	var messages logging.Messages
	for i, condition := range s.conditions {
		prefix := fmt.Sprintf("[%d/%d]", i, len(s.conditions))

		result := condition(state)
		messages = append(messages, logging.StatusMessage(fmt.Sprintf("%s %s", prefix, result.Description)))

		if !result.Ok {
			if !result.Message.Empty() {
				messages = append(messages, result.Message)
			}
			return messages
		}
	}

	s.ready = true
	messages = append(messages, logging.StatusMessage(s.readyMsg))
	return messages
}
