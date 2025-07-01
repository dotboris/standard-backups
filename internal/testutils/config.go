package testutils

import (
	"fmt"
	"os"
	"path"
	"testing"
)

type TestConfig struct {
	Dir         string
	BackendsDir string
	RecipesDir  string
	t           *testing.T
}

func NewTestConfig(t *testing.T) *TestConfig {
	t.Helper()
	configDir := t.TempDir()

	backendsDir := path.Join(configDir, "backends.d")
	err := os.Mkdir(backendsDir, 0o755)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}

	recipesDir := path.Join(configDir, "recipes.d")
	err = os.Mkdir(recipesDir, 0o755)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}

	return &TestConfig{
		Dir:         configDir,
		BackendsDir: backendsDir,
		RecipesDir:  recipesDir,
		t:           t,
	}
}

func (tc *TestConfig) AddBackend(name string, bin string) {
	err := os.WriteFile(
		path.Join(tc.BackendsDir, fmt.Sprintf("%s.yaml", name)),
		[]byte(DedentYaml(fmt.Sprintf(`
			version: 1
			name: %s
			protocol-version: 1
			bin: %s
		`, name, bin))),
		0o644,
	)
	if err != nil {
		tc.t.Error(err)
		tc.t.FailNow()
		return
	}
}

func (tc *TestConfig) AddRecipe(name string, content string) {
	err := os.WriteFile(
		path.Join(tc.RecipesDir, fmt.Sprintf("%s.yaml", name)),
		[]byte(content),
		0o644,
	)
	if err != nil {
		tc.t.Error(err)
		tc.t.FailNow()
		return
	}
}

func (tc *TestConfig) AddBogusRecipe(t *testing.T, name string) {
	d := t.TempDir()
	err := os.WriteFile(
		path.Join(d, "back-me-up.txt"),
		[]byte("back me up"),
		0o644,
	)
	if err != nil {
		tc.t.Error(err)
		tc.t.FailNow()
		return
	}
	tc.AddRecipe(name, DedentYaml(fmt.Sprintf(`
		version: 1
		name: %s
		paths:
			- %s
	`, name, d)))
}

func (tc *TestConfig) WriteConfig(content string) {
	err := os.WriteFile(
		path.Join(tc.Dir, "config.yaml"),
		[]byte(content),
		0o644,
	)
	if err != nil {
		tc.t.Error(err)
		tc.t.FailNow()
		return
	}
}
