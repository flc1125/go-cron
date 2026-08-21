package cron

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedVersion = "go 1.25.0"

func expectedVersionFor(file string) string {
	if filepath.Clean(file) == filepath.Join("internal", "tools", "go.mod") {
		return "go 1.26.0"
	}
	return expectedVersion
}

func TestAllGoModVersions(t *testing.T) {
	var modFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if !info.IsDir() && filepath.Base(path) == "go.mod" {
			modFiles = append(modFiles, path)
		}
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, modFiles)

	for _, file := range modFiles {
		t.Run(file, func(t *testing.T) {
			bytes, err := os.ReadFile(file)
			require.NoError(t, err)

			content := string(bytes)
			assert.NotContains(t, content, "toolchain")

			contents := strings.Split(content, "\n")
			goVersionFound := false

			for _, line := range contents {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "go ") {
					goVersionFound = true
					assert.Equal(t, expectedVersionFor(file), line)
					break
				}
			}

			assert.True(t, goVersionFound)
		})
	}
}
