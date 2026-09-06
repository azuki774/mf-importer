package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"mf-importer/internal/model"
)

type fakeSbiJSONOperator struct {
	files      []string
	snapshots  map[string]*model.SbiSnapshot
	holdings   map[string][]model.SbiHolding
	loadErrors map[string]error
	loadOrder  []string
}

func (f *fakeSbiJSONOperator) GetSbiTargetFiles(context.Context, string) ([]string, error) {
	return append([]string(nil), f.files...), nil
}

func (f *fakeSbiJSONOperator) LoadSbiJSON(_ context.Context, path string) (*model.SbiSnapshot, []model.SbiHolding, error) {
	f.loadOrder = append(f.loadOrder, path)
	if err := f.loadErrors[path]; err != nil {
		return nil, nil, err
	}
	snapshot := *f.snapshots[path]
	holdings := append([]model.SbiHolding(nil), f.holdings[path]...)
	return &snapshot, holdings, nil
}

type fakeSbiDBClient struct {
	insertedByFetchedAt map[time.Time]bool
	calls               []sbiImportCall
	err                 error
}

type sbiImportCall struct {
	snapshot model.SbiSnapshot
	holdings []model.SbiHolding
}

func (f *fakeSbiDBClient) ImportSbiSnapshot(_ context.Context, snapshot *model.SbiSnapshot, holdings []model.SbiHolding) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	f.calls = append(f.calls, sbiImportCall{
		snapshot: *snapshot,
		holdings: append([]model.SbiHolding(nil), holdings...),
	})
	return f.insertedByFetchedAt[snapshot.FetchedAt], nil
}

func newTestSbiSnapshot(fetchedAt time.Time, status model.SbiStatus) *model.SbiSnapshot {
	return &model.SbiSnapshot{FetchedAt: fetchedAt, Status: status}
}

func TestSbiImporter_StartProcessesSortedFilesAndCounts(t *testing.T) {
	timeA := time.Date(2026, 8, 1, 1, 2, 3, 4, time.UTC)
	timeB := time.Date(2026, 8, 2, 1, 2, 3, 4, time.UTC)
	timeC := time.Date(2026, 8, 3, 1, 2, 3, 4, time.UTC)
	op := &fakeSbiJSONOperator{
		files: []string{"z.json", "a.json", "m.json"},
		snapshots: map[string]*model.SbiSnapshot{
			"z.json": newTestSbiSnapshot(timeB, model.SbiStatusOK),
			"a.json": newTestSbiSnapshot(timeA, model.SbiStatusOK),
			"m.json": newTestSbiSnapshot(timeC, model.SbiStatusMaintenance),
		},
		holdings: map[string][]model.SbiHolding{
			"a.json": {{Name: "ダミー銘柄A"}},
			"m.json": {{Name: "ダミー銘柄M"}},
		},
	}
	db := &fakeSbiDBClient{insertedByFetchedAt: map[time.Time]bool{timeA: true, timeB: false, timeC: true}}
	i := &SbiImporter{DBClient: db, Operator: op}

	got, err := i.Start(context.Background(), "/data")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if want := (SbiImportResult{Files: 3, Inserted: 2, Skipped: 1}); !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
	if want := []string{"a.json", "m.json", "z.json"}; !reflect.DeepEqual(op.loadOrder, want) {
		t.Fatalf("load order = %#v, want %#v", op.loadOrder, want)
	}
	if len(db.calls) != 3 {
		t.Fatalf("DB calls = %d, want 3", len(db.calls))
	}
	if len(db.calls[0].holdings) != 1 || len(db.calls[1].holdings) != 0 || len(db.calls[2].holdings) != 0 {
		t.Fatalf("holding lengths = %d, %d, %d; want 1, 0, 0", len(db.calls[0].holdings), len(db.calls[1].holdings), len(db.calls[2].holdings))
	}
}

func TestSbiImporter_StartRejectsEmptyDirectory(t *testing.T) {
	i := &SbiImporter{Operator: &fakeSbiJSONOperator{}}
	if _, err := i.Start(context.Background(), "/data"); err == nil {
		t.Fatal("Start() error = nil, want an empty-target error")
	}
}

func TestSbiImporter_StartStopsOnLoadError(t *testing.T) {
	wantErr := errors.New("synthetic parse error")
	op := &fakeSbiJSONOperator{
		files:      []string{"a.json", "b.json"},
		snapshots:  map[string]*model.SbiSnapshot{"a.json": newTestSbiSnapshot(time.Unix(0, 0), model.SbiStatusOK)},
		loadErrors: map[string]error{"a.json": wantErr},
	}
	i := &SbiImporter{Operator: op, DBClient: &fakeSbiDBClient{}}
	if _, err := i.Start(context.Background(), "/data"); !errors.Is(err, wantErr) {
		t.Fatalf("Start() error = %v, want %v", err, wantErr)
	}
	if want := []string{"a.json"}; !reflect.DeepEqual(op.loadOrder, want) {
		t.Fatalf("load order = %#v, want %#v", op.loadOrder, want)
	}
}
