// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package schema

import (
	"encoding/json"
	"fmt"
	"io"

	tfjson "github.com/hashicorp/terraform-json"
)

// Load reads a `tofu/terraform providers schema -json` document from r and
// returns a Schema. It returns an error if the document cannot be decoded
// or is missing its format_version marker.
func Load(r io.Reader) (*Schema, error) {
	var raw tfjson.ProviderSchemas
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode providers schema: %w", err)
	}
	if raw.FormatVersion == "" {
		return nil, fmt.Errorf("schema document missing format_version")
	}
	return &Schema{underlying: &raw}, nil
}
