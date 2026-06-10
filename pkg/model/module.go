// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// Module is the root inspection result for a single Terraform/OpenTofu
// module directory.
type Module struct {
	Path               string                         `json:"path"`                          // filesystem path to the module directory
	RequiredCore       []string                       `json:"required_core,omitempty"`       // required_version constraints from terraform blocks, in encounter order
	RequiredProviders  map[string]ProviderRequirement `json:"required_providers,omitempty"`  // required_providers entries keyed by provider name
	Variables          []Variable                     `json:"variables,omitempty"`           // variable blocks declared by the module
	Outputs            []Output                       `json:"outputs,omitempty"`             // output blocks declared by the module
	Locals             []Local                        `json:"locals,omitempty"`              // named values from locals blocks
	ManagedResources   []Resource                     `json:"managed_resources,omitempty"`   // resource blocks (managed resources)
	DataResources      []Resource                     `json:"data_resources,omitempty"`      // data blocks (data sources)
	ModuleCalls        []ModuleCall                   `json:"module_calls,omitempty"`        // module blocks invoking child modules
	Providers          []ProviderConfig               `json:"providers,omitempty"`           // provider configuration blocks
	Moved              []MovedBlock                   `json:"moved,omitempty"`               // moved blocks
	Imports            []ImportBlock                  `json:"imports,omitempty"`             // import blocks
	Removed            []RemovedBlock                 `json:"removed,omitempty"`             // removed blocks
	Checks             []CheckBlock                   `json:"checks,omitempty"`              // check blocks
	EphemeralResources []EphemeralResource            `json:"ephemeral_resources,omitempty"` // ephemeral blocks (ephemeral resources)
	Encryption         *Encryption                    `json:"encryption,omitempty"`          // OpenTofu encryption block; nil for Terraform modules
	Children           map[string]*ChildModule        `json:"children,omitempty"`            // resolved child modules keyed by call name; populated only with WithModuleGraph
	Diagnostics        Diagnostics                    `json:"diagnostics,omitempty"`         // problems found while loading the module
}

// ChildModule represents one resolved (or attempted) module call.
type ChildModule struct {
	CallName string          `json:"call_name"`          // name of the module call that produced this child
	Source   string          `json:"source"`             // raw source argument from the module block
	Version  string          `json:"version,omitempty"`  // version constraint, for registry modules
	Resolved *ResolvedSource `json:"resolved,omitempty"` // what was fetched and where, on success
	Module   *Module         `json:"module,omitempty"`   // populated on success
	Error    *Diagnostic     `json:"error,omitempty"`    // populated on failure
}

// ResolvedSource records what we fetched and where it lives.
type ResolvedSource struct {
	Kind      string `json:"kind"`                 // "local" | "registry" | "git" | "http"
	Address   string `json:"address"`              // canonicalized source string
	CachePath string `json:"cache_path,omitempty"` // absolute path to extracted module
	Ref       string `json:"ref,omitempty"`        // git ref (commit SHA)
	Version   string `json:"version,omitempty"`    // registry version
}

// ProviderRequirement describes a single entry inside a
// terraform.required_providers block.
type ProviderRequirement struct {
	Source               string   `json:"source,omitempty"`                // provider source address, e.g. "hashicorp/aws"
	VersionConstraints   []string `json:"version_constraints,omitempty"`   // version constraint strings, e.g. ["~> 4.0"]
	ConfigurationAliases []string `json:"configuration_aliases,omitempty"` // e.g. ["aws.east", "aws.west"]
	Position             Position `json:"position"`                        // source position of the entry
}

// Variable describes a single variable {} block declared by the module.
type Variable struct {
	Name        string       `json:"name"`                  // variable name
	Type        string       `json:"type,omitempty"`        // typeexpr.TypeString
	Default     *Expression  `json:"default,omitempty"`     // default value expression, if any
	Description string       `json:"description,omitempty"` // human-readable description
	Sensitive   bool         `json:"sensitive,omitempty"`   // whether the value is marked sensitive
	Nullable    *bool        `json:"nullable,omitempty"`    // pointer: distinguish unset vs false
	Ephemeral   bool         `json:"ephemeral,omitempty"`   // whether the variable is ephemeral (TF/OpenTofu 1.10+)
	Validations []Validation `json:"validations,omitempty"` // validation blocks attached to the variable
	Position    Position     `json:"position"`              // source position of the variable block
}

// Validation describes a validation {} block attached to a Variable or
// to a resource lifecycle precondition/postcondition.
type Validation struct {
	Condition    Expression `json:"condition"`     // boolean condition expression
	ErrorMessage Expression `json:"error_message"` // message shown when the condition fails
	Position     Position   `json:"position"`      // source position of the validation block
}

// Output describes a single output {} block declared by the module.
type Output struct {
	Name        string     `json:"name"`                  // output name
	Value       Expression `json:"value"`                 // output value expression
	Description string     `json:"description,omitempty"` // human-readable description
	Sensitive   bool       `json:"sensitive,omitempty"`   // whether the value is marked sensitive
	Ephemeral   bool       `json:"ephemeral,omitempty"`   // whether the output is ephemeral (TF/OpenTofu 1.10+)
	DependsOn   []string   `json:"depends_on,omitempty"`  // traversal source forms
	Position    Position   `json:"position"`              // source position of the output block
}

// Local describes a single name = value binding inside a locals {} block.
type Local struct {
	Name     string     `json:"name"`     // local value name
	Value    Expression `json:"value"`    // value expression
	Position Position   `json:"position"` // source position of the binding
}

// ResourceMode distinguishes managed resources from data resources.
type ResourceMode string

// Resource mode values.
const (
	ManagedResourceMode   ResourceMode = "managed"
	DataResourceMode      ResourceMode = "data"
	EphemeralResourceMode ResourceMode = "ephemeral"
)

// Resource describes a single resource {} or data {} block.
type Resource struct {
	Mode     ResourceMode `json:"mode"`               // managed, data, or ephemeral
	Type     string       `json:"type"`               // resource type, e.g. "aws_instance"
	Name     string       `json:"name"`               // local name, e.g. "web"
	Provider string       `json:"provider,omitempty"` // from `provider =` meta-arg, if set

	Count     *Expression `json:"count,omitempty"`      // count meta-argument expression, if set
	ForEach   *Expression `json:"for_each,omitempty"`   // for_each meta-argument expression, if set
	DependsOn []string    `json:"depends_on,omitempty"` // explicit dependency references

	// AttrNames lists the user-set top-level attribute names that are not
	// meta-arguments, captured at load time and sorted. It excludes
	// count, for_each, provider, depends_on, and the lifecycle block. It
	// is best-effort: nested blocks (for example a versioning {} block)
	// do not appear here.
	AttrNames []string `json:"attr_names,omitempty"`

	// SchemaFindings holds schema-derived annotations. It is populated
	// only when inspection runs with a provider schema (WithSchema or
	// WithSchemaAuto); otherwise it is nil and omitted from JSON.
	SchemaFindings *SchemaFindings `json:"schema_findings,omitempty"`

	Lifecycle *Lifecycle `json:"lifecycle,omitempty"` // lifecycle block, if present
	Position  Position   `json:"position"`            // source position of the resource block
}

// Lifecycle describes a resource lifecycle {} block.
type Lifecycle struct {
	CreateBeforeDestroy *bool        `json:"create_before_destroy,omitempty"` // create_before_destroy setting, if set
	PreventDestroy      *bool        `json:"prevent_destroy,omitempty"`       // prevent_destroy setting, if set
	IgnoreChanges       []string     `json:"ignore_changes,omitempty"`        // attribute names or "all"
	ReplaceTriggeredBy  []string     `json:"replace_triggered_by,omitempty"`  // replace_triggered_by references
	Preconditions       []Validation `json:"preconditions,omitempty"`         // precondition blocks
	Postconditions      []Validation `json:"postconditions,omitempty"`        // postcondition blocks
}

// ModuleCall describes a module {} block invoking another module.
type ModuleCall struct {
	Name      string            `json:"name"`                 // module call name
	Source    string            `json:"source"`               // raw source argument
	Version   string            `json:"version,omitempty"`    // version constraint, for registry modules
	Count     *Expression       `json:"count,omitempty"`      // count meta-argument expression, if set
	ForEach   *Expression       `json:"for_each,omitempty"`   // for_each meta-argument expression, if set
	DependsOn []string          `json:"depends_on,omitempty"` // explicit dependency references
	Providers map[string]string `json:"providers,omitempty"`  // local -> remote
	Position  Position          `json:"position"`             // source position of the module block
}

// ProviderConfig describes a single provider {} configuration block.
type ProviderConfig struct {
	Name     string      `json:"name"`               // provider name, e.g. "aws"
	Alias    string      `json:"alias,omitempty"`    // provider alias, if set
	ForEach  *Expression `json:"for_each,omitempty"` // for_each expression for multi-instance providers (OpenTofu 1.9+)
	Position Position    `json:"position"`           // source position of the provider block
}
