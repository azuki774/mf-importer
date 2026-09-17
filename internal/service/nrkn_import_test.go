package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mf-importer/internal/model"
	"mf-importer/internal/repository"
	"mf-importer/internal/service"
)

type nrknTestDB struct {
	calls int
	fail  bool
	seen  map[string]bool
}

func (d *nrknTestDB) ImportNrknSnapshot(ctx context.Context, s *model.NrknSnapshot, h []model.NrknHolding) (bool, error) {
	d.calls++
	if d.fail {
		return false, errors.New("synthetic DB failure")
	}
	key := s.FetchedAt.String()
	if d.seen[key] {
		return false, nil
	}
	d.seen[key] = true
	return true, nil
}

func TestNrknImporterFiles(t *testing.T) {
	raw, err := os.ReadFile("../../test/nrkn_example.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []string{"success", "invalid", "database", "empty", "cancelled"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			db := &nrknTestDB{seen: make(map[string]bool), fail: scenario == "database"}
			if scenario != "empty" {
				if err := os.Mkdir(filepath.Join(dir, "nested"), 0700); err != nil {
					t.Fatal(err)
				}
				files := map[string][]byte{
					"00-future.json":           []byte(`{"schema_version":"2099-01-01","fetched_at":false}`),
					"01-valid.JSON":            raw,
					"nested/02-duplicate.json": raw,
					"ignored.txt":              []byte("not JSON"),
				}
				if scenario == "invalid" {
					files["02-invalid.json"] = []byte(`{"schema_version":"2026-09-14"}`)
				}
				for path, data := range files {
					if err := os.WriteFile(filepath.Join(dir, path), data, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			if scenario == "cancelled" {
				cancel()
			}
			result, err := service.NewNrknImporter(nil, db, &repository.NrknJSONOperator{}).Start(ctx, dir)
			if scenario == "success" {
				if err != nil || result.Files != 3 || result.Inserted != 1 || result.Skipped != 1 || result.Unsupported != 1 {
					t.Fatalf("result=%+v error=%v", result, err)
				}
			} else if err == nil {
				t.Fatal("expected failure")
			}
			if scenario == "invalid" && (db.calls != 1 || result.Inserted != 1) {
				t.Fatal("did not stop after invalid file")
			}
			if scenario == "cancelled" && db.calls != 0 {
				t.Fatal("DB called after cancellation")
			}
			if err != nil && strings.Contains(err.Error(), "ダミー商品") {
				t.Fatal("holding data leaked")
			}
		})
	}
}
