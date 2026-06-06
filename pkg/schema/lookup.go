// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schema

import (
	"sort"

	tfjson "github.com/hashicorp/terraform-json"
)

// LookupResource returns the schema for resourceType and the provider
// source address it was found under. It first tries an exact match on
// providerSource, then falls back to any provider that defines
// resourceType (deterministically, by sorted source address). The
// fallback also matches across registries (for example a config that
// declares registry.terraform.io while the schema was emitted by OpenTofu
// under registry.opentofu.org). It returns (nil, "") when no provider
// defines resourceType.
func (s *Schema) LookupResource(providerSource, resourceType string) (*tfjson.Schema, string) {
	return s.lookup(providerSource, resourceType, func(p *tfjson.ProviderSchema) map[string]*tfjson.Schema {
		return p.ResourceSchemas
	})
}

// LookupDataSource is LookupResource for data sources.
func (s *Schema) LookupDataSource(providerSource, dataSourceType string) (*tfjson.Schema, string) {
	return s.lookup(providerSource, dataSourceType, func(p *tfjson.ProviderSchema) map[string]*tfjson.Schema {
		return p.DataSourceSchemas
	})
}

// LookupEphemeralResource is LookupResource for ephemeral resources.
func (s *Schema) LookupEphemeralResource(providerSource, resourceType string) (*tfjson.Schema, string) {
	return s.lookup(providerSource, resourceType, func(p *tfjson.ProviderSchema) map[string]*tfjson.Schema {
		return p.EphemeralResourceSchemas
	})
}

// lookup implements the shared exact-then-fallback resolution used by the
// resource, data source, and ephemeral resource lookups. selector picks
// the relevant schema map off a provider schema.
func (s *Schema) lookup(
	providerSource, blockType string,
	selector func(*tfjson.ProviderSchema) map[string]*tfjson.Schema,
) (*tfjson.Schema, string) {
	if s == nil || s.underlying == nil {
		return nil, ""
	}

	if providerSource != "" {
		if provider, ok := s.underlying.Schemas[providerSource]; ok {
			if block, ok := selector(provider)[blockType]; ok {
				return block, providerSource
			}
		}
	}

	sources := make([]string, 0, len(s.underlying.Schemas))
	for source := range s.underlying.Schemas {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	for _, source := range sources {
		if block, ok := selector(s.underlying.Schemas[source])[blockType]; ok {
			return block, source
		}
	}
	return nil, ""
}
