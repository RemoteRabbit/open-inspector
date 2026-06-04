// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// Module is the root inspection result for a single Terraform/OpenTofu
// module directory.
type Module struct {
	Path string `json:"path"`

	RequiredCore      []string                       `json:"required_core,omitempty"`
	RequiredProviders map[string]ProviderRequirement `json:"required_providers,omitempty"`

	Variables []Variable `json:"variables,omitempty"`
	Outputs   []Output   `json:"outputs,omitempty"`
	Locals    []Local    `json:"locals,omitempty"`

	ManagedResources []Resource `json:"managed_resources,omitempty"`
	DataResources    []Resource `json:"data_resources,omitempty"`

	ModuleCalls []ModuleCall     `json:"module_calls,omitempty"`
	Providers   []ProviderConfig `json:"providers,omitempty"`

	Moved   []MovedBlock   `json:"moved,omitempty"`
	Imports []ImportBlock  `json:"imports,omitempty"`
	Removed []RemovedBlock `json:"removed,omitempty"`

	Checks []CheckBlock `json:"checks,omitempty"`

	Diagnostics Diagnostics `json:"diagnostics,omitempty"`
}

// ProviderRequirement describes a single entry inside a
// terraform.required_providers block.
type ProviderRequirement struct {
	Source               string   `json:"source,omitempty"`
	VersionConstraints   []string `json:"version_constraints,omitempty"`
	ConfigurationAliases []string `json:"configuration_aliases,omitempty"` // e.g. ["aws.east", "aws.west"]
	Range                Range    `json:"range"`
}

// Variable describes a single variable {} block declared by the module.
type Variable struct {
	Name        string       `json:"name"`
	Type        string       `json:"type,omitempty"` // typeexpr.TypeString
	Default     *Expression  `json:"default,omitempty"`
	Description string       `json:"description,omitempty"`
	Sensitive   bool         `json:"sensitive,omitempty"`
	Nullable    *bool        `json:"nullable,omitempty"` // pointer: distinguish unset vs false
	Ephemeral   bool         `json:"ephemeral,omitempty"`
	Validations []Validation `json:"validations,omitempty"`
	Range       Range        `json:"range"`
}

// Validation describes a validation {} block attached to a Variable or
// to a resource lifecycle precondition/postcondition.
type Validation struct {
	Condition    Expression `json:"condition"`
	ErrorMessage Expression `json:"error_message"`
	Range        Range      `json:"range"`
}

// Output describes a single output {} block declared by the module.
type Output struct {
	Name        string     `json:"name"`
	Value       Expression `json:"value"`
	Description string     `json:"description,omitempty"`
	Sensitive   bool       `json:"sensitive,omitempty"`
	Ephemeral   bool       `json:"ephemeral,omitempty"`
	DependsOn   []string   `json:"depends_on,omitempty"` // traversal source forms
	Range       Range      `json:"range"`
}

// Local describes a single name = value binding inside a locals {} block.
type Local struct {
	Name  string     `json:"name"`
	Value Expression `json:"value"`
	Range Range      `json:"range"`
}

// ResourceMode distinguishes managed resources from data resources.
type ResourceMode string

// Resource mode values.
const (
	ManagedResourceMode ResourceMode = "managed"
	DataResourceMode    ResourceMode = "data"
)

// Resource describes a single resource {} or data {} block.
type Resource struct {
	Mode     ResourceMode `json:"mode"`
	Type     string       `json:"type"`
	Name     string       `json:"name"`
	Provider string       `json:"provider,omitempty"` // from `provider =` meta-arg, if set

	Count     *Expression `json:"count,omitempty"`
	ForEach   *Expression `json:"for_each,omitempty"`
	DependsOn []string    `json:"depends_on,omitempty"`

	Lifecycle *Lifecycle `json:"lifecycle,omitempty"`
	Range     Range      `json:"range"`
}

// Lifecycle describes a resource lifecycle {} block.
type Lifecycle struct {
	CreateBeforeDestroy *bool        `json:"create_before_destroy,omitempty"`
	PreventDestroy      *bool        `json:"prevent_destroy,omitempty"`
	IgnoreChanges       []string     `json:"ignore_changes,omitempty"` // attribute names or "all"
	ReplaceTriggeredBy  []string     `json:"replace_triggered_by,omitempty"`
	Preconditions       []Validation `json:"preconditions,omitempty"`
	Postconditions      []Validation `json:"postconditions,omitempty"`
}

// ModuleCall describes a module {} block invoking another module.
type ModuleCall struct {
	Name      string            `json:"name"`
	Source    string            `json:"source"`
	Version   string            `json:"version,omitempty"`
	Count     *Expression       `json:"count,omitempty"`
	ForEach   *Expression       `json:"for_each,omitempty"`
	DependsOn []string          `json:"depends_on,omitempty"`
	Providers map[string]string `json:"providers,omitempty"` // local -> remote
	Range     Range             `json:"range"`
}

// ProviderConfig describes a single provider {} configuration block.
type ProviderConfig struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitempty"`
	Range Range  `json:"range"`
}
