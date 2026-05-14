package scanner

import "testing"

func TestFilterExcludesDoubleStarDirectory(t *testing.T) {
	filter := NewFilter([]string{"**/.codex/**"})

	paths := []string{
		"/Users/example/.codex",
		"/Users/example/.codex/skills/.system/openai-docs/assets/openai.png",
	}

	for _, path := range paths {
		if !filter.IsExcluded(path) {
			t.Fatalf("expected %q to be excluded", path)
		}
	}
}

func TestFilterExcludesOtherDoubleStarDirectory(t *testing.T) {
	filter := NewFilter([]string{"**/node_modules/**"})

	if !filter.IsExcluded("/Users/example/project/node_modules/pkg/index.js") {
		t.Fatal("expected node_modules file to be excluded")
	}
}
