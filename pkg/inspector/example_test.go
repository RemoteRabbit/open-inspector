// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package inspector_test

import (
	"fmt"
	"os"

	"github.com/remoterabbit/open-inspector/pkg/inspector"
)

func Example() {
	module, err := inspector.Inspect("../../testdata/fixtures/simple")
	if err != nil {
		fmt.Println("inspect failed:", err)
		return
	}
	fmt.Println("variables:", len(module.Variables))
	fmt.Println("resources:", len(module.ManagedResources))
	fmt.Println("outputs:", len(module.Outputs))
	// Output:
	// variables: 1
	// resources: 1
	// outputs: 1
}

func Example_withModuleGraph() {
	module, err := inspector.Inspect("../../testdata/fixtures/multi-module",
		inspector.WithModuleGraph())
	if err != nil {
		fmt.Println("inspect failed:", err)
		return
	}
	fmt.Println("children:", len(module.Children))
	// Output:
	// children: 2
}

func Example_withSchema() {
	schemaFile, err := os.Open("../config/testdata/schemas/null.json")
	if err != nil {
		fmt.Println("open schema:", err)
		return
	}
	defer func() { _ = schemaFile.Close() }()

	module, err := inspector.Inspect("../../testdata/fixtures/simple",
		inspector.WithSchema(schemaFile))
	if err != nil {
		fmt.Println("inspect failed:", err)
		return
	}
	resource := module.ManagedResources[0]
	if resource.SchemaFindings == nil {
		fmt.Println("no findings")
	} else {
		fmt.Println("unknown:", len(resource.SchemaFindings.UnknownAttrs))
	}
	// Output:
	// no findings
}
