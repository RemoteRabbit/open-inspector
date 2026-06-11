// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/remoterabbit/open-inspector/pkg/model"
)

// mergeOverrides applies each override module to base in order. For each top-level construct, identity is matched by:
//   - Variable / Output / Local: Name
//   - Resource / DataResource / EphemeralResource: Type + Name
//   - ModuleCall: Name
//   - ProviderConfig: Name + Alias
//
// Argument replacement: override wins for any field the override set. Validation blocks REPLACE the entire validation
// set (per TF docs). depends_on REPLACES rather than appending.
//
// TODO: an override block with no matching base construct is silently ignored (best-effort merge); elevate this to
// a diagnostic so typo's override targets are surfaced rather than dropped.
func mergeOverrides(base *model.Module, overrides []*model.Module) {
	for _, override := range overrides {
		mergeVariables(base, override.Variables)
		mergeOutputs(base, override.Outputs)
		mergeLocals(base, override.Locals)
		mergeResources(&base.ManagedResources, override.ManagedResources)
		mergeResources(&base.DataResources, override.DataResources)
		mergeEphemeralResources(&base.EphemeralResources, override.EphemeralResources)
		mergeModuleCalls(base, override.ModuleCalls)
		mergeProviderConfig(base, override.Providers)
		// RequiredProviders and RequiredCore: append; overrides can refine but Terraform doesn't define a strict
		// merge rule here, so we leave both untouched on override.
	}
}

// mergeVariables applies override variables onto base, matched by name.
// Each field the override set wins; validation blocks replace the whole set.
func mergeVariables(base *model.Module, overrides []model.Variable) {
	for _, override := range overrides {
		for index := range base.Variables {
			if base.Variables[index].Name != override.Name {
				continue
			}
			if override.Type != "" {
				base.Variables[index].Type = override.Type
				base.Variables[index].TypeSpec = override.TypeSpec // may be nil if override type was unparsable
			}
			if override.Default != nil {
				base.Variables[index].Default = override.Default
				base.Variables[index].DefaultValue = override.DefaultValue
			}
			if override.Description != "" {
				base.Variables[index].Description = override.Description
			}
			if override.Sensitive {
				base.Variables[index].Sensitive = override.Sensitive
			}
			if override.Nullable != nil {
				base.Variables[index].Nullable = override.Nullable
			}
			if override.Ephemeral {
				base.Variables[index].Ephemeral = override.Ephemeral
			}
			if override.Comment != "" {
				base.Variables[index].Comment = override.Comment
			}
			if len(override.Validations) > 0 {
				base.Variables[index].Validations = override.Validations
			}
			goto next
		}
	next:
	}
}

// mergeOutputs applies override outputs onto base, matched by name. Each
// field the override set wins; depends_on replaces rather than appends.
func mergeOutputs(base *model.Module, overrides []model.Output) {
	for _, override := range overrides {
		for index := range base.Outputs {
			if base.Outputs[index].Name != override.Name {
				continue
			}
			if override.Value.Source != "" {
				base.Outputs[index].Value = override.Value
			}
			if override.Description != "" {
				base.Outputs[index].Description = override.Description
			}
			if override.Sensitive {
				base.Outputs[index].Sensitive = override.Sensitive
			}
			if override.Ephemeral {
				base.Outputs[index].Ephemeral = override.Ephemeral
			}
			if len(override.DependsOn) > 0 {
				base.Outputs[index].DependsOn = override.DependsOn
			}
			goto next
		}
	next:
	}
}

// mergeLocals applies override locals onto base, matched by name,
// replacing the value of any local that the override redefines.
func mergeLocals(base *model.Module, overrides []model.Local) {
	for _, override := range overrides {
		for index := range base.Locals {
			if base.Locals[index].Name != override.Name {
				continue
			}
			if override.Value.Source != "" {
				base.Locals[index].Value = override.Value
			}
			goto next
		}
	next:
	}
}

// mergeResources applies override resources onto base, matched by type and
// name. Each meta-argument the override set wins; depends_on replaces.
func mergeResources(base *[]model.Resource, overrides []model.Resource) {
	for _, override := range overrides {
		for index := range *base {
			if (*base)[index].Type != override.Type || (*base)[index].Name != override.Name {
				continue
			}
			if override.Provider != "" {
				(*base)[index].Provider = override.Provider
			}
			if override.Count != nil {
				(*base)[index].Count = override.Count
			}
			if override.ForEach != nil {
				(*base)[index].ForEach = override.ForEach
			}
			if len(override.DependsOn) > 0 {
				(*base)[index].DependsOn = override.DependsOn
			}
			if override.Lifecycle != nil {
				(*base)[index].Lifecycle = override.Lifecycle
			}
			goto next
		}
	next:
	}
}

// mergeEphemeralResources applies override ephemeral resources onto base,
// matched by type and name, with the same field-replacement rules as
// mergeResources.
func mergeEphemeralResources(base *[]model.EphemeralResource, overrides []model.EphemeralResource) {
	for _, override := range overrides {
		for index := range *base {
			if (*base)[index].Type != override.Type || (*base)[index].Name != override.Name {
				continue
			}
			if override.Provider != "" {
				(*base)[index].Provider = override.Provider
			}
			if override.Count != nil {
				(*base)[index].Count = override.Count
			}
			if override.ForEach != nil {
				(*base)[index].ForEach = override.ForEach
			}
			if len(override.DependsOn) > 0 {
				(*base)[index].DependsOn = override.DependsOn
			}
			if override.Lifecycle != nil {
				(*base)[index].Lifecycle = override.Lifecycle
			}
			goto next
		}
	next:
	}
}

// mergeModuleCalls applies override module calls onto base, matched by
// name. Each field the override set wins; depends_on replaces.
func mergeModuleCalls(base *model.Module, overrides []model.ModuleCall) {
	for _, override := range overrides {
		for index := range base.ModuleCalls {
			if base.ModuleCalls[index].Name != override.Name {
				continue
			}
			if override.Source != "" {
				base.ModuleCalls[index].Source = override.Source
			}
			if override.Version != "" {
				base.ModuleCalls[index].Version = override.Version
			}
			if override.Count != nil {
				base.ModuleCalls[index].Count = override.Count
			}
			if override.ForEach != nil {
				base.ModuleCalls[index].ForEach = override.ForEach
			}
			if len(override.DependsOn) > 0 {
				base.ModuleCalls[index].DependsOn = override.DependsOn
			}
			if len(override.Providers) > 0 {
				base.ModuleCalls[index].Providers = override.Providers
			}
			goto next
		}
	next:
	}
}

// mergeProviderConfig applies override provider blocks onto base, matched
// by name and alias, replacing fields the override set.
func mergeProviderConfig(base *model.Module, overrides []model.ProviderConfig) {
	for _, override := range overrides {
		for index := range base.Providers {
			if base.Providers[index].Name != override.Name || base.Providers[index].Alias != override.Alias {
				continue
			}
			if override.ForEach != nil {
				base.Providers[index].ForEach = override.ForEach
			}
			goto next
		}
	next:
	}
}
