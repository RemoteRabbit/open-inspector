// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schema

import (
	"sort"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// Enrich annotates each managed resource, data source, and ephemeral
// resource on module with a model.SchemaFindings derived from s. It is
// best-effort and static-only:
//
//   - It sees attribute names the user wrote, not their values; an
//     attribute set to null is treated as set.
//   - Nested blocks (for example aws_s3_bucket's versioning {}) are not in
//     AttrNames, so their attributes are not checked.
//   - dynamic {} blocks are not expanded; only the static declaration is
//     visible.
//   - count = 0 and for_each = {} resources are still annotated; the
//     expressions are not evaluated.
//
// Resources whose type the schema does not cover are left untouched (no
// findings, no diagnostic): the caller chose the schema and presumably
// knows which providers it covers.
func Enrich(module *model.Module, s *Schema) {
	if module == nil || s == nil {
		return
	}
	sources := providerSourcesByName(module.RequiredProviders)

	for index := range module.ManagedResources {
		resource := &module.ManagedResources[index]
		block, _ := s.LookupResource(providerSource(sources, resource.Provider, resource.Type), resource.Type)
		resource.SchemaFindings = findingsFor(block, resource.AttrNames, resource.Range)
	}
	for index := range module.DataResources {
		resource := &module.DataResources[index]
		block, _ := s.LookupDataSource(providerSource(sources, resource.Provider, resource.Type), resource.Type)
		resource.SchemaFindings = findingsFor(block, resource.AttrNames, resource.Range)
	}
	for index := range module.EphemeralResources {
		resource := &module.EphemeralResources[index]
		block, _ := s.LookupEphemeralResource(providerSource(sources, resource.Provider, resource.Type), resource.Type)
		resource.SchemaFindings = findingsFor(block, resource.AttrNames, resource.Range)
	}
}

// findingsFor compares the user-set attribute names against the schema
// block and returns the resulting findings, or nil when the schema does
// not cover the block or there is nothing to report.
func findingsFor(block *tfjson.Schema, attrNames []string, blockRange model.Range) *model.SchemaFindings {
	if block == nil || block.Block == nil {
		return nil
	}

	schemaAttrs := block.Block.Attributes
	userSet := make(map[string]bool, len(attrNames))
	for _, name := range attrNames {
		userSet[name] = true
	}

	findings := &model.SchemaFindings{}
	for _, name := range attrNames {
		attribute, ok := schemaAttrs[name]
		if !ok {
			findings.UnknownAttrs = append(findings.UnknownAttrs, model.AttrFinding{
				Name:  name,
				Range: blockRange,
			})
			continue
		}
		if attribute.Deprecated {
			findings.DeprecatedAttrs = append(findings.DeprecatedAttrs, model.DeprecatedAttr{
				Name:    name,
				Message: attribute.Description,
				Range:   blockRange,
			})
		}
	}

	for name, attribute := range schemaAttrs {
		// Computed-only attributes are set by the provider, never the
		// user, so they are never "missing required".
		if attribute.Required && !userSet[name] {
			findings.MissingRequired = append(findings.MissingRequired, name)
		}
	}
	sort.Strings(findings.MissingRequired)

	if len(findings.UnknownAttrs) == 0 &&
		len(findings.DeprecatedAttrs) == 0 &&
		len(findings.MissingRequired) == 0 {
		return nil
	}
	return findings
}

// providerSourcesByName maps each declared provider local name to its full
// source address. It normalizes the shorthand forms:
//
//	"hashicorp/null"                         -> "registry.terraform.io/hashicorp/null"
//	"registry.opentofu.org/hashicorp/null"   -> unchanged
//	legacy entry with no source              -> "registry.terraform.io/hashicorp/<name>"
func providerSourcesByName(requirements map[string]model.ProviderRequirement) map[string]string {
	out := make(map[string]string, len(requirements))
	for name, requirement := range requirements {
		source := requirement.Source
		switch {
		case source == "":
			source = "registry.terraform.io/hashicorp/" + name
		case strings.Count(source, "/") == 1:
			source = "registry.terraform.io/" + source
		}
		out[name] = source
	}
	return out
}

// providerSource resolves the full provider source address for a resource
// given its explicit provider meta-arg (which may be aliased, for example
// "aws.east") or, failing that, the provider-name prefix of its type.
func providerSource(sources map[string]string, providerMeta, resourceType string) string {
	name := providerName(providerMeta, resourceType)
	return sources[name]
}

// providerName derives the provider local name from a resource's provider
// meta-arg or type. The alias suffix on a meta-arg ("aws.east") is
// stripped; otherwise the prefix before the first underscore of the type
// ("aws_instance" -> "aws") is used.
func providerName(providerMeta, resourceType string) string {
	if providerMeta != "" {
		return strings.SplitN(providerMeta, ".", 2)[0]
	}
	if index := strings.Index(resourceType, "_"); index > 0 {
		return resourceType[:index]
	}
	return resourceType
}
