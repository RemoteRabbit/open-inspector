// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/zclconf/go-cty/cty"

	"github.com/remoterabbit/open-inspector/pkg/model"
)

// ctyValueToValue converts a known, constant cty.Value into a model.Value.
// It returns nil for unknown values (which cannot be serialized faithfully).
// Null values become a typed model.Value with Kind == ValueNull.
func ctyValueToValue(value cty.Value) *model.Value {
	if !value.IsKnown() {
		return nil
	}
	if value.IsNull() {
		return &model.Value{Kind: model.ValueNull}
	}

	t := value.Type()
	switch {
	case t == cty.String:
		return &model.Value{Kind: model.ValueString, String: value.AsString()}
	case t == cty.Number:
		return &model.Value{Kind: model.ValueNumber, Number: value.AsBigFloat().Text('f', -1)}
	case t == cty.Bool:
		return &model.Value{Kind: model.ValueBool, Bool: value.True()}

	case t.IsListType(), t.IsSetType():
		return &model.Value{Kind: model.ValueList, List: elementsOf(value)}
	case t.IsTupleType():
		return &model.Value{Kind: model.ValueTuple, Tuple: elementsOf(value)}

	case t.IsMapType():
		return &model.Value{Kind: model.ValueMap, Map: entriesOf(value)}
	case t.IsObjectType():
		return &model.Value{Kind: model.ValueObject, Object: entriesOf(value)}

	default:
		return nil
	}
}

// elementsOf converts the elements of a list/set/tuple value in iteration
// order. For sets, cty iterates in a deterministic sorted order.
func elementsOf(value cty.Value) []*model.Value {
	var out []*model.Value
	for iterator := value.ElementIterator(); iterator.Next(); {
		_, element := iterator.Element()
		if converted := ctyValueToValue(element); converted != nil {
			out = append(out, converted)
		}
	}
	return out
}

// entriesOf converts the key/value pairs of a map/object value. Keys are
// always strings for both map and object types.
func entriesOf(value cty.Value) map[string]*model.Value {
	out := map[string]*model.Value{}
	for iterator := value.ElementIterator(); iterator.Next(); {
		key, element := iterator.Element()
		if converted := ctyValueToValue(element); converted != nil {
			out[key.AsString()] = converted
		}
	}
	return out
}
