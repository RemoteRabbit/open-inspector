// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package inspector

import (
	"path/filepath"

	"github.com/remoterabbit/open-inspector/pkg/graph"
	"github.com/remoterabbit/open-inspector/pkg/sources"
)

// Option configures an Inspect call.
type Option func(*options)

type options struct {
	moduleGraph bool
	maxDepth    int
	cacheDir    string
}

// WithModuleGraph enables recursive resolution of module calls into Module.Children.
func WithModuleGraph() Option {
	return func(opts *options) {
		opts.moduleGraph = true
	}
}

// WithMaxDepth limits how deep the module graph is resolved.
func WithMaxDepth(depth int) Option {
	return func(opts *options) {
		opts.maxDepth = depth
	}
}

// WithCache sets the directory used to cache fetched (registry/git/http) modules.
func WithCache(dir string) Option {
	return func(opts *options) {
		opts.cacheDir = filepath.Clean(dir)
	}
}

func defaultOptions() options {
	return options{
		moduleGraph: false,
		maxDepth:    16,
		cacheDir:    sources.DefaultCacheDir(),
	}
}

// toGraphOptions projects the inspector-level options onto the graph-owned Options struct.
func (opts options) toGraphOptions() graph.Options {
	return graph.Options{
		MaxDepth: opts.maxDepth,
		CacheDir: opts.cacheDir,
	}
}
