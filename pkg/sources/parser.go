// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sources

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// labelPattern matches a single registry path label (namespace, name, or provider): letters, digits, hyphens,
// or underscores.
var labelPattern = regexp.MustCompile(`^[0-9A-Za-z_-]+$`)

// SourceKind classifies a module source string by the transport used to fetch it.
type SourceKind int

// Source kinds recognized by ParseSource.
const (
	SourceLocal SourceKind = iota
	SourceRegistry
	SourceGit
	SourceHTTP
	SourceUnknown
)

// ParsedSource is the classified, decomposed form of a module source string.
type ParsedSource struct {
	Kind      SourceKind
	Raw       string // original
	Host      string // for registry: "registry.terraform.io" if implicit
	Namespace string // registry only
	Name      string // registry only
	Provider  string // registry only
	URL       string // git, http
	Ref       string // git only
	Subdir    string // shared across kinds
}

// ParseSource classifies a Terraform/OpenTofu module source string.
func ParseSource(raw string) (ParsedSource, error) {
	parsed := ParsedSource{Raw: raw}
	source := strings.TrimSpace(raw)
	if source == "" {
		return parsed, fmt.Errorf("empty module source")
	}

	switch {
	case isLocal(source):
		parsed.Kind = SourceLocal
		return parsed, nil
	case strings.HasPrefix(source, "git::"),
		strings.HasPrefix(source, "github.com/"),
		strings.HasPrefix(source, "bitbucket.org/"),
		strings.HasPrefix(source, "gitlab.com/"):
		return parseGit(source, parsed)
	case strings.HasPrefix(source, "https://"), strings.HasPrefix(source, "http://"):
		return parseHTTP(source, parsed)
	}

	address, subdir := splitSubdir(source)
	if parts := strings.Split(address, "/"); looksLikeRegistry(parts) {
		return parseRegistry(parts, subdir, parsed)
	}

	parsed.Kind = SourceUnknown
	return parsed, fmt.Errorf("unrecognized module source %q", raw)
}

func isLocal(source string) bool {
	switch {
	case source == ".", source == "..":
		return true
	case strings.HasPrefix(source, "./"), strings.HasPrefix(source, "../"):
		return true
	case strings.HasPrefix(source, "/"):
		return true
	}
	return false
}

func parseGit(source string, parsed ParsedSource) (ParsedSource, error) {
	parsed.Kind = SourceGit
	rest := strings.TrimPrefix(source, "git::")

	if index := strings.Index(rest, "?"); index >= 0 {
		if values, err := url.ParseQuery(rest[index+1:]); err == nil {
			parsed.Ref = values.Get("ref")
		}
		rest = rest[:index]
	}
	rest, parsed.Subdir = splitSubdir(rest)

	switch {
	case strings.HasPrefix(rest, "github.com/"), strings.HasPrefix(rest, "bitbucket.org/"), strings.HasPrefix(rest, "gitlab.com/"):
		parsed.URL = "https://" + strings.TrimSuffix(rest, ".git") + ".git"
	default:
		parsed.URL = rest
	}
	return parsed, nil
}

func parseHTTP(source string, parsed ParsedSource) (ParsedSource, error) {
	parsed.Kind = SourceHTTP
	parsed.URL, parsed.Subdir = splitSubdir(source)
	return parsed, nil
}

func looksLikeRegistry(parts []string) bool {
	switch len(parts) {
	case 3:
		return labelPattern.MatchString(parts[0]) &&
			labelPattern.MatchString(parts[1]) &&
			labelPattern.MatchString(parts[2])
	case 4: // explicit host (must contain a dot) + namespace/name/provider
		return strings.Contains(parts[0], ".") &&
			labelPattern.MatchString(parts[1]) &&
			labelPattern.MatchString(parts[2]) &&
			labelPattern.MatchString(parts[3])
	}
	return false
}

func parseRegistry(parts []string, subdir string, parsed ParsedSource) (ParsedSource, error) {
	parsed.Kind = SourceRegistry
	parsed.Subdir = subdir
	if len(parts) == 4 {
		parsed.Host = strings.ToLower(parts[0])
		parts = parts[1:]
	} else {
		parsed.Host = "registry.terraform.io"
	}
	parsed.Namespace, parsed.Name, parsed.Provider = parts[0], parts[1], parts[2]
	return parsed, nil
}

// splitSubdir separates a "//subdir" suffix from the address, ignoring the "//" inside
// a "scheme://" prefix. The query string (if any) is stripped from the returned address;
// callers that need `?ref=` parse it before calling.
func splitSubdir(source string) (address, subdir string) {
	stripped := source
	if index := strings.Index(stripped, "?"); index >= 0 {
		stripped = stripped[:index]
	}
	searchFrom := 0
	if index := strings.Index(stripped, "://"); index >= 0 {
		searchFrom = index + 3
	}
	idx := strings.Index(stripped[searchFrom:], "//")
	if idx < 0 {
		return stripped, ""
	}
	idx += searchFrom
	return stripped[:idx], strings.TrimPrefix(stripped[idx:], "//")
}
