package wire

import (
	"os"
	"strings"
	"testing"
)

func TestRepositoryAutomationContract(t *testing.T) {
	t.Parallel()

	required := map[string][]string{
		"CLAUDE.md": {"@AGENTS.md"},
		"Makefile": {
			"release-patch",
			"release-minor",
			"release-major",
			"scripts/release.sh",
		},
		"llms.txt": {
			"# go-wire",
			"llms-full.txt",
			"docs/QUICKSTART.md",
		},
		"llms-full.txt": {"# go-wire", "# Quickstart"},
		"README.md":     {"llms.txt", "llms-full.txt", "CHANGELOG.md"},
		".github/workflows/ci.yml": {
			"go test -race ./...",
			"scripts/check-coverage.sh",
			"staticcheck ./...",
			"go vet ./...",
			"gofmt",
			"scripts/check-docs.sh",
		},
		".github/workflows/fuzz.yml": {
			"FuzzDecode",
			"FuzzParse",
			"FuzzRoundTrip",
			"jsonwire",
			"xmlwire",
			"soap",
			"yamlwire",
			"tomlwire",
			"msgpackwire",
			"cborwire",
			"bsonwire",
		},
		".github/workflows/benchmark.yml": {
			"-bench", "upload-artifact", "yamlwire", "tomlwire",
			"msgpackwire", "cborwire", "bsonwire",
		},
		".github/workflows/security.yml": {"govulncheck", "dependency-review"},
		".github/workflows/release.yml": {
			"tags:",
			`"v*"`,
			"go run ./cmd/semvercheck",
			"merge-base --is-ancestor",
			"gh release create",
			"sha256sum",
		},
		".github/dependabot.yml": {"gomod", "github-actions"},
		"scripts/check-docs.sh": {
			"go test ./...",
			"Markdown links",
			"generate-llms.py --check",
		},
		"scripts/generate-llms.py": {"README.md", "--check"},
		"scripts/release.sh": {
			"git tag -a",
			"origin/main",
			"scripts/check-coverage.sh",
			"FuzzDecode",
			"FuzzParse",
			"FuzzRoundTrip",
			"yamlwire",
			"tomlwire",
			"msgpackwire",
			"cborwire",
			"bsonwire",
		},
	}

	for path, fragments := range required {
		path, fragments := path, fragments
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", path, err)
			}
			for _, fragment := range fragments {
				if !strings.Contains(string(contents), fragment) {
					t.Errorf("%s does not contain %q", path, fragment)
				}
			}
		})
	}
}

func TestRepositoryRequiresPatchedGo125OrNewer(t *testing.T) {
	t.Parallel()

	required := map[string]string{
		"go.mod":             "go 1.25.8",
		"README.md":          "Go 1.25.8 or newer",
		"CONTRIBUTING.md":    "Go 1.25.8 or newer",
		"docs/QUICKSTART.md": "Go 1.25.8 and newer",
	}
	for path, fragment := range required {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		if !strings.Contains(string(contents), fragment) {
			t.Errorf("%s does not contain %q", path, fragment)
		}
	}
}

func TestCodecDocumentationUsesRealAPIsAndSemantics(t *testing.T) {
	t.Parallel()

	prohibited := map[string][]string{
		"SECURITY.md": {
			"YAML duplicate keys, aliases, merge keys, excessive alias expansion",
		},
		"docs/ADOPTION.md": {
			"Keep YAML aliases, merge keys, duplicate keys, and multiple documents disabled",
		},
		"docs/TROUBLESHOOTING.md": {
			"AllowAliases",
			"DecodeAllDocuments",
			"AllowNonStringMapKeys",
			"must still be exact",
		},
		"docs/MIGRATION.md": {
			"duplicate keys, aliases, merge keys, and multiple documents are rejected by default",
			"exact double-to-integer conversion",
			"ProfileCanonical",
			"ProfileCoreDeterministic",
			"ProfileCTAP2",
		},
	}
	for path, fragments := range prohibited {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", path, err)
		}
		normalized := strings.Join(strings.Fields(string(contents)), " ")
		for _, fragment := range fragments {
			if strings.Contains(normalized, fragment) {
				t.Errorf("%s advertises invalid codec behavior %q", path, fragment)
			}
		}
	}
}
