// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// EphemeralResource corresponds to `ephemeral "<type>" "<name>" {}`
// Same meta-args as a managed resource.
type EphemeralResource struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Provider string `json:"provider,omitempty"`

	Count     *Expression `json:"count,omitempty"`
	ForEach   *Expression `json:"for_each,omitempty"`
	DependsOn []string    `json:"depends_on,omitempty"`

	Lifecycle *Lifecycle `json:"lifecycle,omitempty"`
	Range     Range      `json:"range"`
}
