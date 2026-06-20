package repository

import (
	"context"
	"mf-importer/internal/logger"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/japanese"
)

const cfFixtureUTF8 = `,2024/07/19,はま寿司,-1705,三井住友カード,食費,外食,,,
,2024/07/16,ローソン,-291,三井住友カード,食費,食料品,,,
`

const bsFixtureUTF8 = `日付,合計,預金・現金・暗号資産,株式(現物),投資信託,ポイント,詳細
2025-08-02,"5,000,000","3,500,000",0,"1,400,000","100,000",テスト詳細
2025-08-01,"4,800,000","3,200,000",0,"1,500,000","100,000",テスト詳細
`

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
}

func encodeSJIS(t *testing.T, s string) []byte {
	t.Helper()
	enc := japanese.ShiftJIS.NewEncoder()
	out, err := enc.Bytes([]byte(s))
	if err != nil {
		t.Fatalf("failed to encode SJIS: %v", err)
	}
	return out
}

func newOp(encoding string) *DetailCSVOperator {
	return &DetailCSVOperator{Logger: logger.NewLogger(), Encoding: encoding}
}

func TestLoadCfCSV_UTF8_DefaultEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cf.csv")
	writeFile(t, path, []byte(cfFixtureUTF8))

	op := newOp("")
	got, err := op.LoadCfCSV(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadCfCSV() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(details) = %d, want 2", len(got))
	}
	if got[0].Name != "はま寿司" {
		t.Errorf("Name[0] = %q, want はま寿司", got[0].Name)
	}
	if got[1].Name != "ローソン" {
		t.Errorf("Name[1] = %q, want ローソン", got[1].Name)
	}
	if got[0].FinIns != "三井住友カード" {
		t.Errorf("FinIns[0] = %q, want 三井住友カード", got[0].FinIns)
	}
}

func TestLoadCfCSV_UTF8_Explicit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cf.csv")
	writeFile(t, path, []byte(cfFixtureUTF8))

	op := newOp("utf8")
	got, err := op.LoadCfCSV(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadCfCSV() error = %v", err)
	}
	if len(got) != 2 || got[0].Name != "はま寿司" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestLoadCfCSV_SJIS(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cf.csv")
	writeFile(t, path, encodeSJIS(t, cfFixtureUTF8))

	op := newOp("sjis")
	got, err := op.LoadCfCSV(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadCfCSV() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(details) = %d, want 2", len(got))
	}
	if got[0].Name != "はま寿司" {
		t.Errorf("Name[0] = %q, want はま寿司 (SJIS decode failed?)", got[0].Name)
	}
	if got[1].Name != "ローソン" {
		t.Errorf("Name[1] = %q, want ローソン (SJIS decode failed?)", got[1].Name)
	}
}

func TestLoadCfCSV_SJIS_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cf.csv")
	writeFile(t, path, encodeSJIS(t, cfFixtureUTF8))

	op := newOp("SJIS")
	got, err := op.LoadCfCSV(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadCfCSV() error = %v", err)
	}
	if len(got) != 2 || got[0].Name != "はま寿司" {
		t.Errorf("unexpected result with SJIS uppercase: %+v", got)
	}
}

func TestLoadCfCSV_DefaultFailsOnSJISBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cf.csv")
	writeFile(t, path, encodeSJIS(t, cfFixtureUTF8))

	op := newOp("utf8")
	got, err := op.LoadCfCSV(context.Background(), path)
	if err != nil {
		// Errors are acceptable but not required.
		return
	}
	for _, d := range got {
		if d.Name == "はま寿司" || d.Name == "ローソン" {
			t.Errorf("expected garbled text for SJIS bytes read as UTF-8, but got correct value: %q", d.Name)
		}
	}
}

func TestLoadBsHistoryCSV_UTF8(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "asset_history.csv")
	writeFile(t, path, []byte(bsFixtureUTF8))

	op := newOp("")
	got, err := op.LoadBsHistoryCSV(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadBsHistoryCSV() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(histories) = %d, want 2", len(got))
	}
	if got[0].Details != "テスト詳細" {
		t.Errorf("Details[0] = %q, want テスト詳細", got[0].Details)
	}
	if got[0].TotalAmount != 5000000 {
		t.Errorf("TotalAmount[0] = %d, want 5000000", got[0].TotalAmount)
	}
}

func TestLoadBsHistoryCSV_SJIS(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "asset_history.csv")
	writeFile(t, path, encodeSJIS(t, bsFixtureUTF8))

	op := newOp("sjis")
	got, err := op.LoadBsHistoryCSV(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadBsHistoryCSV() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(histories) = %d, want 2", len(got))
	}
	if got[0].Details != "テスト詳細" {
		t.Errorf("Details[0] = %q, want テスト詳細 (SJIS decode failed?)", got[0].Details)
	}
}
