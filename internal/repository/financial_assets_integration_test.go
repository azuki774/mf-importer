package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	driver "github.com/go-sql-driver/mysql"
	migrate "github.com/rubenv/sql-migrate"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"mf-importer/internal/migration"
	"mf-importer/internal/model"
	"mf-importer/internal/repository"
)

func financialTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("MF_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set MF_MIGRATION_TEST_DSN to disposable MariaDB")
	}
	cfg, err := driver.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid test DSN")
	}
	cfg.DBName = ""
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal("open disposable MariaDB")
	}
	t.Cleanup(func() { admin.Close() })
	name := fmt.Sprintf("mf_financial_read_test_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal("create disposable database")
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP DATABASE " + name); err != nil {
			t.Error("drop test database")
		}
	})
	cfg.DBName, cfg.ParseTime, cfg.Loc = name, true, time.FixedZone("Asia/Tokyo", 9*60*60)
	sqlDB, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatal("open test database")
	}
	t.Cleanup(func() { sqlDB.Close() })
	if _, err := migration.Run(context.Background(), sqlDB, migrate.Up, 0); err != nil {
		t.Fatal("apply migrations")
	}
	gdb, err := gorm.Open(gormmysql.New(gormmysql.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("open GORM test connection")
	}
	sqlDB.SetMaxOpenConns(1)
	if err := gdb.Exec("SET time_zone = '+00:00'").Error; err != nil {
		t.Fatal("set UTC SQL session")
	}
	return gdb
}

func TestFinancialAssetReadsMariaDB(t *testing.T) {
	gdb := financialTestDB(t)
	db := &repository.DBClient{Conn: gdb}
	ctx := context.Background()
	// All rows below are hand-written synthetic data, including the >2^53 integer
	// chosen to detect accidental float64 conversion and tiny DOUBLE quantity.
	exec := func(query string, args ...any) {
		t.Helper()
		if gdb.Exec(query, args...).Error != nil {
			t.Fatal("seed synthetic financial data")
		}
	}
	base := "2000-01-01 12:00:00.000000"
	exec("INSERT INTO nrkn_snapshot (id,fetched_at,status,schema_version,grand_total_jpy,total_cost_jpy,pnl_jpy,created_at) VALUES (1,?,'OK','test',9007199254740993,9007199254740992,-3,'2000-01-01 12:00:00')", base)
	exec("INSERT INTO nrkn_snapshot (id,fetched_at,status,schema_version,grand_total_jpy,total_cost_jpy,pnl_jpy) VALUES (2,?,'OK','test',10,7,3)", "2000-01-01 12:00:00.000002")
	exec("INSERT INTO nrkn_snapshot (id,fetched_at,status,schema_version,grand_total_jpy,total_cost_jpy,pnl_jpy) VALUES (3,?,'ERROR','test',12,9,3)", "2000-01-01 12:00:00.000003")
	exec("INSERT INTO nrkn_holding (snapshot_id,product_code,composite_figi,name,category,quantity,unit_price,value_jpy,cost_jpy,redemption_unit_price,redemption_value_jpy,pnl_jpy,reference_date,allocation_pct,unit_price_raw,redemption_unit_price_raw) VALUES (1,'TEST001','TESTFIGI0001','ダミー商品A','ダミー',1e-35,1,9007199254740993,9007199254740992,1,1,1,'2000-01-01',1,'1','1')")
	got, err := db.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Sources: []string{"nrkn"}})
	if err != nil || len(got) != 2 {
		t.Fatal("list NRKN/status filter")
	}
	if got[0].ID != 2 || *got[1].ValuationJpy != "9007199254740993" || *got[1].CostJpy != "9007199254740992" || *got[1].UnrealizedPnlJpy != "-3" {
		t.Fatal("NRKN integer precision")
	}
	page, err := db.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Sources: []string{"nrkn"}, Offset: 1, Limit: 1})
	if err != nil || len(page) != 1 || page[0].ID != 1 {
		t.Fatal("offset boundary")
	}
	holdings, err := db.GetFinancialHoldings(ctx, "nrkn", []int64{1, 2})
	if err != nil || len(holdings) != 1 {
		t.Fatal("bulk NRKN holdings")
	}
	if holdings[0].Quantity == nil || *holdings[0].Quantity != "0.00000000000000000000000000000000001" || *holdings[0].ValuationJpy != "9007199254740993" || *holdings[0].ReferenceDate != "2000-01-01" {
		t.Fatal("NRKN holding precision/date")
	}
	exec("INSERT INTO sbi_snapshot (id,fetched_at,status,schema_version,grand_total_jpy,created_at) VALUES (1,?,'OK','test',123456789012.34,'2000-01-01 12:00:00')", base)
	exec("INSERT INTO sbi_snapshot (id,fetched_at,status,schema_version,grand_total_jpy) VALUES (2,?,'OK','test',NULL)", "2000-01-02 12:00:00")
	exec("INSERT INTO sbi_holding (snapshot_id,section,name,quantity,unit_cost,unit_price,pnl_jpy,pnl_pct,value_jpy) VALUES (1,'nisa_funds','ダミー商品B',1.123456,2.123456,3.123456,4.56,0.1256,123456789012.34)")
	// Equal timestamps in different sources must share a deterministic global
	// order, with LIMIT/OFFSET applied after merging and filtering both sources.
	for offset, want := range []struct {
		source string
		id     int64
	}{{"sbi", 2}, {"nrkn", 2}, {"sbi", 1}, {"nrkn", 1}} {
		page, err := db.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Sources: []string{"nrkn", "sbi"}, Limit: 1, Offset: offset})
		if err != nil || len(page) != 1 || page[0].Source != want.source || page[0].ID != want.id {
			t.Fatalf("incorrect combined offset page %d", offset)
		}
		if want.source == "nrkn" && want.id == 1 && (page[0].ValuationJpy == nil || *page[0].ValuationJpy != "9007199254740993") {
			t.Fatal("UNION must preserve integer precision")
		}
		if want.source == "sbi" && want.id == 1 && (page[0].ValuationJpy == nil || *page[0].ValuationJpy != "123456789012.34") {
			t.Fatal("UNION must preserve decimal precision")
		}
		if want.source == "sbi" && (page[0].CostJpy != nil || page[0].UnrealizedPnlJpy != nil) {
			t.Fatal("UNION must preserve null metrics")
		}
	}
	for _, offset := range []int{4, 100, int(^uint(0) >> 1)} {
		page, err := db.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Sources: []string{"sbi", "nrkn"}, Limit: 2, Offset: offset})
		if err != nil || page == nil || len(page) != 0 {
			t.Fatal("out-of-range combined offset must return empty rows")
		}
	}
	snapshot, err := db.GetFinancialSnapshot(ctx, "sbi", 1)
	if err != nil || snapshot.ValuationJpy == nil || *snapshot.ValuationJpy != "123456789012.34" || snapshot.CostJpy != nil || snapshot.UnrealizedPnlJpy != nil {
		t.Fatal("SBI decimal/null mapping")
	}
	holdings, err = db.GetFinancialHoldings(ctx, "sbi", []int64{1})
	if err != nil || len(holdings) != 1 || *holdings[0].Quantity != "1.123456" || *holdings[0].ValuationJpy != "123456789012.34" || holdings[0].CostJpy != nil {
		t.Fatal("SBI holding precision")
	}
	snapshot, err = db.GetFinancialSnapshot(ctx, "sbi", 2)
	if err != nil || snapshot.ValuationJpy != nil {
		t.Fatal("SBI NULL total")
	}
	wantFetched := time.Date(2000, 1, 1, 12, 0, 0, 0, time.FixedZone("JST", 9*3600))
	wantImported := time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)
	if !got[1].FetchedAt.Equal(wantFetched) || !got[1].ImportedAt.Equal(wantImported) {
		t.Fatal("DATETIME JST or TIMESTAMP UTC session decoding")
	}
	// SQL TIMESTAMP session rendering must not shift the reported imported instant.
	exec("SET time_zone = '+09:00'")
	shifted, err := db.GetFinancialSnapshot(ctx, "nrkn", 1)
	if err != nil || !shifted.ImportedAt.Equal(wantImported) || !shifted.FetchedAt.Equal(wantFetched) {
		t.Fatal("session-independent importedAt")
	}
	for _, tc := range []struct {
		delta     time.Duration
		inclusive bool
		want      int64
	}{{0, true, 1}, {0, false, 0}, {500 * time.Nanosecond, true, 1}, {500 * time.Nanosecond, false, 1}, {-500 * time.Nanosecond, true, 0}, {-500 * time.Nanosecond, false, 0}} {
		row, err := db.LatestFinancialSnapshot(ctx, "nrkn", wantFetched.Add(tc.delta), tc.inclusive)
		if err != nil {
			t.Fatal(err)
		}
		if tc.want == 0 && row != nil || tc.want != 0 && (row == nil || row.ID != tc.want) {
			t.Fatal("latest microsecond boundary")
		}
	}
	from := wantFetched.Add(500 * time.Nanosecond)
	rows, err := db.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Sources: []string{"nrkn"}, From: &from})
	if err != nil || len(rows) != 1 || rows[0].ID != 2 {
		t.Fatal("inclusive lower submicro boundary")
	}
	to := wantFetched.Add(500 * time.Nanosecond)
	rows, err = db.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Sources: []string{"nrkn"}, To: &to})
	if err != nil || len(rows) != 1 || rows[0].ID != 1 {
		t.Fatal("exclusive upper submicro boundary")
	}
	to = wantFetched.Add(-500 * time.Nanosecond)
	rows, err = db.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Sources: []string{"nrkn"}, To: &to})
	if err != nil || len(rows) != 0 {
		t.Fatal("exclusive upper before snapshot")
	}
	if _, err := db.GetFinancialSnapshot(ctx, "nrkn", 3); !errors.Is(err, model.ErrRecordNotFound) {
		t.Fatal("non-OK detail exposed")
	}
	if _, err := db.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Sources: []string{"invalid-source"}}); err == nil {
		t.Fatal("invalid source accepted")
	}
	filteredTo := wantFetched.Add(time.Microsecond)
	page, err = db.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Sources: []string{"sbi", "nrkn"}, From: &wantFetched, To: &filteredTo, Limit: 1, Offset: 1})
	if err != nil || len(page) != 1 || page[0].Source != "nrkn" || page[0].ID != 1 {
		t.Fatal("date filters must apply before global offset")
	}
}
