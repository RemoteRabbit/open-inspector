// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package graph

// loadState tracks the recursion stack for cycle detection and depth limiting during a single Build call.
type loadState struct {
	visited map[string]bool
	depth   int
}

func newLoadState() *loadState {
	return &loadState{
		visited: map[string]bool{},
	}
}
