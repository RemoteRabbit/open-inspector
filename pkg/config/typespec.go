// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/zclconf/go-cty/cty"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// ctyTypeToSpec converts a parsed cty.Type (and the optional-attribute
// defaults that came with it) into a model.TypeSpec tree. defaults may be
// nil when the type has no optional attributes anywhere.
func ctyTypeToSpec(t cty.Type, defaults *typeexpr.Defaults) *model.TypeSpec {
	switch {
	case t == cty.String:
		return &model.TypeSpec{Kind: model.TypeString}
	case t == cty.Number:
		return &model.TypeSpec{Kind: model.TypeNumber}
	case t == cty.Bool:
		return &model.TypeSpec{Kind: model.TypeBool}
	case t == cty.DynamicPseudoType:
		return &model.TypeSpec{Kind: model.TypeDynamic}

	case t.IsListType():
		return &model.TypeSpec{Kind: model.TypeList, Element: ctyTypeToSpec(t.ElementType(), childDefaults(defaults, ""))}
	case t.IsSetType():
		return &model.TypeSpec{Kind: model.TypeSet, Element: ctyTypeToSpec(t.ElementType(), childDefaults(defaults, ""))}
	case t.IsMapType():
		return &model.TypeSpec{Kind: model.TypeMap, Element: ctyTypeToSpec(t.ElementType(), childDefaults(defaults, ""))}

	case t.IsTupleType():
		elementTypes := t.TupleElementTypes()
		spec := &model.TypeSpec{Kind: model.TypeTuple, Elements: make([]*model.TypeSpec, len(elementTypes))}
		for index, elementType := range elementTypes {
			spec.Elements[index] = ctyTypeToSpec(elementType, nil)
		}
		return spec

	case t.IsObjectType():
		spec := &model.TypeSpec{Kind: model.TypeObject, Attributes: map[string]*model.ObjectAttr{}}
		for name, attributeType := range t.AttributeTypes() {
			attribute := &model.ObjectAttr{
				Type:     ctyTypeToSpec(attributeType, childDefaults(defaults, name)),
				Optional: t.AttributeOptional(name),
			}
			if defaults != nil {
				if value, ok := defaults.DefaultValues[name]; ok {
					attribute.Default = ctyValueToValue(value) // from item 4
				}
			}
			spec.Attributes[name] = attribute
		}
		return spec

	default:
		// Unknown / unsupported construct: fall back to dynamic so callers
		// still get a node rather than nil.
		return &model.TypeSpec{Kind: model.TypeDynamic}
	}
}

// childDefaults returns the Defaults node for a child element/attribute, or
// nil when there is none. Collections store their single child under "".
func childDefaults(defaults *typeexpr.Defaults, key string) *typeexpr.Defaults {
	if defaults == nil {
		return nil
	}
	return defaults.Children[key]
}
