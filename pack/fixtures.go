package pack

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"

	"github.com/furyheimdall/auth-rate-audit-pack/drift"
)

//go:embed testdata/*.json
var testdataFS embed.FS

// FixtureFS is the embedded fixture directory (no network, no cwd).
func FixtureFS() fs.FS {
	sub, err := fs.Sub(testdataFS, "testdata")
	if err != nil {
		panic(err)
	}
	return sub
}

// Fixture is a file- or memory-backed implementation of all four seat
// sources. Other seats can stay stubbed; tests and `arap pack` use these.
type Fixture struct {
	Name      string             `json:"name"`
	Baseline  BaselineSnapshot   `json:"baseline"`
	Drift     DriftFlag          `json:"drift"`
	Inventory InventoryChecklist `json:"inventory"`
	Staging   StagingMatrix      `json:"staging"`
}

// BaselineSnapshot implements BaselineSource.
func (f Fixture) BaselineSnapshot() (BaselineSnapshot, error) { return f.Baseline, nil }

// DriftFlag implements DriftSource.
func (f Fixture) DriftFlag() (DriftFlag, error) { return f.Drift, nil }

// InventoryChecklist implements InventorySource.
func (f Fixture) InventoryChecklist() (InventoryChecklist, error) { return f.Inventory, nil }

// StagingMatrix implements StagingSource.
func (f Fixture) StagingMatrix() (StagingMatrix, error) { return f.Staging, nil }

// Sources returns the fixture as the four pack sources.
func (f Fixture) Sources() Sources {
	return Sources{Baseline: f, Drift: f, Inventory: f, Staging: f}
}

// LoadFixture reads a named JSON fixture from dir (typically testdata).
// No network. Unknown names return an error listing available files.
func LoadFixture(dir fs.FS, name string) (Fixture, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Fixture{}, fmt.Errorf("pack: fixture name is required")
	}
	name = strings.TrimSuffix(name, ".json")
	raw, err := fs.ReadFile(dir, name+".json")
	if err != nil {
		avail, _ := FixtureNames(dir)
		return Fixture{}, fmt.Errorf("pack: fixture %q: %w (available: %s)", name, err, strings.Join(avail, ", "))
	}
	var f Fixture
	if err := json.Unmarshal(raw, &f); err != nil {
		return Fixture{}, fmt.Errorf("pack: fixture %q: %w", name, err)
	}
	if f.Name == "" {
		f.Name = name
	}
	if f.Drift.ThresholdPP == 0 {
		f.Drift.ThresholdPP = drift.ThresholdPP
	}
	return f, nil
}

// FixtureNames lists JSON fixture stems in dir.
func FixtureNames(dir fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".json"))
	}
	return names, nil
}
