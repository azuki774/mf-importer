package migration

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	migrate "github.com/rubenv/sql-migrate"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"mf-importer/internal/model"
	"mf-importer/internal/repository"
)

func TestIntegrationNrknRoundTripAndRollback(t *testing.T) {
	db := testDatabase(t)
	assertRun(t, db, migrate.Up, 8, 8)
	assertRun(t, db, migrate.Up, 0, 1)
	raw, err := os.ReadFile("../../test/nrkn_example.json")
	if err != nil {
		t.Fatal(err)
	}
	gdb, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	client := &repository.DBClient{Conn: gdb}
	s, h, err := model.ParseNrknJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	inserted, err := client.ImportNrknSnapshot(context.Background(), s, h)
	if err != nil || !inserted {
		t.Fatalf("inserted=%v error=%v", inserted, err)
	}
	var got model.NrknSnapshot
	if err := gdb.Table("nrkn_snapshot").First(&got, s.ID).Error; err != nil {
		t.Fatal(err)
	}
	// The driver converts DATETIME using the DSN's location on both write and read.
	if !got.FetchedAt.Equal(s.FetchedAt) {
		t.Fatal("timestamp did not round trip")
	}
	got.FetchedAt, got.CreatedAt, got.UpdatedAt = s.FetchedAt, s.CreatedAt, s.UpdatedAt
	if !reflect.DeepEqual(got, *s) {
		t.Fatal("snapshot did not round trip")
	}
	var gotHolding model.NrknHolding
	if err := gdb.Table("nrkn_holding").First(&gotHolding).Error; err != nil {
		t.Fatal(err)
	}
	// MySQL's DATE is decoded as RFC3339 when scanned into a string.
	if len(gotHolding.ReferenceDate) < 10 || gotHolding.ReferenceDate[:10] != h[0].ReferenceDate {
		t.Fatal("reference date did not round trip")
	}
	gotHolding.ReferenceDate = h[0].ReferenceDate
	gotHolding.CreatedAt, gotHolding.UpdatedAt = h[0].CreatedAt, h[0].UpdatedAt
	if !reflect.DeepEqual(gotHolding, h[0]) {
		t.Fatal("holding did not round trip")
	}
	s2, h2, _ := model.ParseNrknJSON(raw)
	inserted, err = client.ImportNrknSnapshot(context.Background(), s2, h2)
	if err != nil || inserted {
		t.Fatal("duplicate not skipped")
	}
	s3, h3, _ := model.ParseNrknJSON(raw)
	s3.FetchedAt = s3.FetchedAt.Add(time.Hour)
	h3 = append(h3, h3[0])
	inserted, err = client.ImportNrknSnapshot(context.Background(), s3, h3)
	if err == nil || inserted {
		t.Fatal("duplicate product accepted")
	}
	var count int64
	if err := gdb.Table("nrkn_snapshot").Count(&count).Error; err != nil || count != 1 {
		t.Fatal("failed import did not roll back")
	}
	// Separate products may share a FIGI.
	s4, h4, _ := model.ParseNrknJSON(raw)
	s4.FetchedAt = s4.FetchedAt.Add(2 * time.Hour)
	h4 = append(h4, h4[0])
	h4[1].ProductCode = "000002"
	inserted, err = client.ImportNrknSnapshot(context.Background(), s4, h4)
	if err != nil || !inserted {
		t.Fatal("shared FIGI rejected")
	}
	assertRun(t, db, migrate.Down, 1, 1)
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name IN ('nrkn_snapshot', 'nrkn_holding')").Scan(&count); err != nil || count != 0 {
		t.Fatal("NRKN tables not removed")
	}
	assertRun(t, db, migrate.Up, 0, 1)
}
