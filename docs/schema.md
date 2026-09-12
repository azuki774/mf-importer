# DB スキーマ説明書

## 正本と適用

DB スキーマの正本は `migration/db/*.sql` です。マイグレーションは importer 起動時、または `make migration` / `mf-importer migrate up` で、埋め込み SQL を sql-migrate の Go API から適用します。この文書は、AI または人間がマイグレーションと同じ変更で手更新する手管理の説明書です。

## テーブル一覧

### Application tables

| Table | 由来 migration | 用途 |
| --- | --- | --- |
| [`detail`](#detail) | `001_init.sql` | 取り込んだ明細の日時、名称、金額、分類、および mawinter の確認日時を保持する。 |
| [`extract_rule`](#extract_rule) | `001_init.sql` | 明細から抽出するフィールドと値を mawinter のカテゴリに対応付けるルールを保持する。 |
| [`import_history`](#import_history) | `002_import_history.sql`、`003_add_imp_his_filename.sql` | インポートジョブごとの処理件数と入力元ファイルを保持する。 |
| [`asset_history`](#asset_history) | `004_asset_history.sql` | 日付ごとの資産合計と内訳を保持する。 |
| [`sbi_snapshot`](#sbi_snapshot) | `005_sbi_snapshot.sql` | SBI の資産情報を取得した時点のサマリーを保持する。 |
| [`sbi_holding`](#sbi_holding) | `006_sbi_holding.sql` | SBI のスナップショットに含まれる銘柄・商品ごとの保有情報を保持する。 |

## マイグレーション一覧

| Migration | Up の対象 | Down の対象 |
| --- | --- | --- |
| `001_init.sql` | [`detail`](#detail)、[`extract_rule`](#extract_rule) を作成 | [`detail`](#detail)、[`extract_rule`](#extract_rule) を削除 |
| `002_import_history.sql` | [`import_history`](#import_history) を作成 | [`import_history`](#import_history) を削除 |
| `003_add_imp_his_filename.sql` | [`import_history`](#import_history).`src_file` と `idx2` を追加 | [`import_history`](#import_history).`src_file` 列を削除（SQL本文に `idx2` の個別削除はない） |
| `004_asset_history.sql` | [`asset_history`](#asset_history) を作成 | [`asset_history`](#asset_history) を削除 |
| `005_sbi_snapshot.sql` | [`sbi_snapshot`](#sbi_snapshot) を作成 | [`sbi_snapshot`](#sbi_snapshot) を削除 |
| `006_sbi_holding.sql` | [`sbi_holding`](#sbi_holding) を作成 | [`sbi_holding`](#sbi_holding) を削除 |

## 共通事項

- `Null` が `No` の列は SQL に `NOT NULL` が指定されています。`Yes` の列は SQL に NULL 可否の指定がありません。
- `Default` が `—` の列は SQL に `DEFAULT` の指定がありません。
- `created_at` と `updated_at` は各 SQL の `DEFAULT` および `ON UPDATE` をそのまま記載しています。NULL 可否の指定がない場合は `Yes` としています。
- sql-migrate が管理する `gorp_migrations` はツール管理メタデータです。`migration/db/*.sql` の `CREATE TABLE` 対象ではなく、application table には含めません。

## detail

由来 migration: `001_init.sql`

用途: 取り込んだ明細と、明細の分類・mawinter 確認状態を保持する。

SQL 上のテーブル属性: `CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`

### Columns

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `id` | `INT AUTO_INCREMENT` | No | `—` | primary id |
| `yyyymm_id` | `INT` | No | `—` | id for each yyyymm |
| `date` | `DATE` | No | `—` | record date yyyymm |
| `name` | `TEXT` | Yes | `—` | detail name |
| `price` | `INT` | Yes | `—` | `—` |
| `fin_ins` | `TEXT` | Yes | `—` | finance instrcument name |
| `l_category` | `TEXT` | Yes | `—` | large category name |
| `m_category` | `TEXT` | Yes | `—` | medium category name |
| `regist_date` | `DATE` | No | `—` | date running importer |
| `maw_check_date` | `DATE` | Yes | `—` | mawinter check date |
| `maw_regist_date` | `DATE` | Yes | `—` | mawinter regist check date |
| `raw_date` | `TEXT` | Yes | `—` | `—` |
| `raw_price` | `TEXT` | Yes | `—` | `—` |
| `created_at` | `datetime` | Yes | `current_timestamp` | `—` |
| `updated_at` | `timestamp` | Yes | `current_timestamp on update current_timestamp` | `—` |

### Keys and indexes

| Kind | Name | Columns |
| --- | --- | --- |
| PRIMARY KEY | — | `id` |
| UNIQUE KEY | — | なし |
| INDEX | `idx1` | `maw_check_date` |
| INDEX | `idx2` | `name` |
| INDEX | `idx3` | `raw_price` |

## extract_rule

由来 migration: `001_init.sql`

用途: 抽出対象のフィールド・値と mawinter カテゴリの対応ルールを保持する。

SQL 上のテーブル属性: `CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`

### Columns

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `id` | `INT AUTO_INCREMENT` | No | `—` | primary id |
| `field_name` | `TEXT` | No | `—` | extract field name (m_category or name) |
| `value` | `TEXT` | No | `—` | `—` |
| `exact_match` | `INT` | Yes | `—` | exact match = 1 or not 0 |
| `category_id` | `INT` | No | `—` | mawinter category id |
| `created_at` | `datetime` | Yes | `current_timestamp` | `—` |
| `updated_at` | `timestamp` | Yes | `current_timestamp on update current_timestamp` | `—` |

### Keys and indexes

| Kind | Name | Columns |
| --- | --- | --- |
| PRIMARY KEY | — | `id` |
| UNIQUE KEY | — | なし |
| INDEX | — | なし |

## import_history

由来 migration: `002_import_history.sql`（本体）、`003_add_imp_his_filename.sql`（`src_file` と `idx2` を追加）

用途: インポートジョブのラベル、処理件数、入力元ファイルを保持する。

SQL 上のテーブル属性: `CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`

### Columns

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `id` | `INT AUTO_INCREMENT` | No | `—` | primary id |
| `job_label` | `TEXT` | Yes | `—` | importer sets joblabel |
| `parsed_entry_num` | `INT` | No | `—` | `—` |
| `new_entry_num` | `INT` | No | `—` | `—` |
| `created_at` | `datetime` | Yes | `current_timestamp` | `—` |
| `updated_at` | `timestamp` | Yes | `current_timestamp on update current_timestamp` | `—` |
| `src_file` | `TEXT` | Yes | `—` | `—` |

`src_file` は `003_add_imp_his_filename.sql` の `ALTER TABLE` で追加されます。

### Keys and indexes

| Kind | Name | Columns |
| --- | --- | --- |
| PRIMARY KEY | — | `id` |
| UNIQUE KEY | — | なし |
| INDEX | `idx1` | `job_label` |
| INDEX | `idx2` | `src_file` |

## asset_history

由来 migration: `004_asset_history.sql`

用途: 日付単位の資産合計と、預金・株式・投資信託・ポイントなどの内訳を保持する。

SQL 上のテーブル属性: `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`（照合順序は SQL で未指定）

### Columns

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `id` | `INT AUTO_INCREMENT` | No | `—` | `—` |
| `date` | `DATE` | No | `—` | `—` |
| `total_amount` | `INT` | No | `—` | 合計 |
| `cash_deposit_crypto` | `INT` | No | `—` | 預金・現金・暗号資産 |
| `stocks` | `INT` | No | `—` | 株式(現物) |
| `investment_trusts` | `INT` | No | `—` | 投資信託 |
| `points` | `INT` | No | `—` | ポイント |
| `details` | `TEXT` | Yes | `—` | 詳細 |
| `created_at` | `TIMESTAMP` | Yes | `CURRENT_TIMESTAMP` | `—` |
| `updated_at` | `TIMESTAMP` | Yes | `CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | `—` |

### Keys and indexes

| Kind | Name | Columns |
| --- | --- | --- |
| PRIMARY KEY | — | `id` |
| UNIQUE KEY | `unique_date` | `date` |
| INDEX | — | なし（上記の UNIQUE KEY を除く） |

## sbi_snapshot

由来 migration: `005_sbi_snapshot.sql`、nullable化と時刻COMMENT修正: `007_nullable_sbi_values.sql`

用途: SBI から取得した資産サマリーを、取得時刻ごとのスナップショットとして保持する。

SQL 上のテーブル属性: `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

`fetched_at` は `Asia/Tokyo` の壁時計時刻へ変換してから `DATETIME(6)` に保存するため、DB上は microsecond（マイクロ秒）精度です。

### 資産値と `NULL` の扱い

#### `status = OK`

- スナップショットと保有明細の必須資産値を検証します。
- 欠損があれば取り込みエラーとし、スナップショットを保存しません。
- 明示的なゼロはゼロとして保存し、`NULL` のまま保存しません。

#### `status = MAINTENANCE / ERROR`

- 取得試行の記録として `sbi_snapshot` を保存します。
- すべての資産値を `NULL` とします。
- `sbi_holding` は保存しません。

#### 保有米国株の前日比

- `OK` でも生成元の仕様上取得できないため、`prev_day_jpy` と `prev_day_pct` だけを `NULL` とします。

`status` の `scraper emits ...` や `schema_version` の `CurrentSchemaVersion` などは migration SQL の COMMENT を転記したものです。このリポジトリ内では、これらの外部契約を定義・説明しておらず、外部契約として断定しません。

### Columns

#### Core / total

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `id` | `INT AUTO_INCREMENT` | No | `—` | `—` |
| `fetched_at` | `DATETIME(6)` | No | `—` | Assets.fetched_at (Asia/Tokyo wall-clock, truncated to microseconds) |
| `status` | `VARCHAR(16)` | No | `—` | OK\|MAINTENANCE\|ERROR (normalized to uppercase on ingest; scraper emits ok/maintenance) |
| `schema_version` | `INT` | No | `—` | Assets.schema_version (CurrentSchemaVersion) |
| `grand_total_jpy` | `DECIMAL(14,2)` | Yes | `—` | grand_total_jpy = nisa+old_nisa+cash+others; NULL when incomplete |

#### NISA summary

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `nisa_total_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.total_jpy |
| `nisa_prev_day_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.prev_day_jpy |
| `nisa_prev_day_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.prev_day_pct |
| `nisa_prev_month_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.prev_month_jpy |
| `nisa_prev_month_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.prev_month_pct |
| `nisa_pnl_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.pnl_jpy (評価損益) |
| `nisa_pnl_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.pnl_pct |

#### NISA domestic stocks

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `nisa_domestic_value_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.domestic_stocks.value_jpy |
| `nisa_domestic_pnl_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.domestic_stocks.pnl_jpy |
| `nisa_domestic_pnl_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.domestic_stocks.pnl_pct |
| `nisa_domestic_prev_day_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.domestic_stocks.prev_day_jpy |
| `nisa_domestic_prev_day_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.domestic_stocks.prev_day_pct |
| `nisa_domestic_prev_month_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.domestic_stocks.prev_month_jpy |
| `nisa_domestic_prev_month_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.domestic_stocks.prev_month_pct |

#### NISA US stocks

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `nisa_us_value_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.us_stocks.value_jpy |
| `nisa_us_pnl_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.us_stocks.pnl_jpy |
| `nisa_us_pnl_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.us_stocks.pnl_pct |
| `nisa_us_prev_day_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.us_stocks.prev_day_jpy |
| `nisa_us_prev_day_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.us_stocks.prev_day_pct |
| `nisa_us_prev_month_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.us_stocks.prev_month_jpy |
| `nisa_us_prev_month_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.us_stocks.prev_month_pct |

#### NISA funds

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `nisa_funds_value_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.funds.value_jpy |
| `nisa_funds_pnl_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.funds.pnl_jpy |
| `nisa_funds_pnl_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.funds.pnl_pct |
| `nisa_funds_prev_day_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.funds.prev_day_jpy |
| `nisa_funds_prev_day_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.funds.prev_day_pct |
| `nisa_funds_prev_month_jpy` | `DECIMAL(14,2)` | Yes | `—` | nisa.funds.prev_month_jpy |
| `nisa_funds_prev_month_pct` | `DECIMAL(10,4)` | Yes | `—` | nisa.funds.prev_month_pct |

#### Old NISA

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `old_nisa_total_jpy` | `DECIMAL(14,2)` | Yes | `—` | old_nisa.total_jpy |
| `old_nisa_prev_day_jpy` | `DECIMAL(14,2)` | Yes | `—` | old_nisa.prev_day_jpy |
| `old_nisa_prev_day_pct` | `DECIMAL(10,4)` | Yes | `—` | old_nisa.prev_day_pct |
| `old_nisa_pnl_jpy` | `DECIMAL(14,2)` | Yes | `—` | old_nisa.pnl_jpy |
| `old_nisa_pnl_pct` | `DECIMAL(10,4)` | Yes | `—` | old_nisa.pnl_pct |

#### Cash / others

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `cash_jpy_amount` | `DECIMAL(14,2)` | Yes | `—` | cash.jpy.amount (== value_jpy for JPY) |
| `cash_jpy_value_jpy` | `DECIMAL(14,2)` | Yes | `—` | cash.jpy.value_jpy |
| `cash_usd_amount` | `DECIMAL(14,4)` | Yes | `—` | cash.usd.amount (USD) |
| `cash_usd_value_jpy` | `DECIMAL(14,2)` | Yes | `—` | cash.usd.value_jpy (JPY converted) |
| `other_funds_amount` | `DECIMAL(14,2)` | Yes | `—` | others.funds.amount |
| `other_funds_value_jpy` | `DECIMAL(14,2)` | Yes | `—` | others.funds.value_jpy |

#### Timestamps

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `created_at` | `TIMESTAMP` | Yes | `CURRENT_TIMESTAMP` | `—` |
| `updated_at` | `TIMESTAMP` | Yes | `CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | `—` |

### Keys and indexes

| Kind | Name | Columns |
| --- | --- | --- |
| PRIMARY KEY | — | `id` |
| UNIQUE KEY | `uq_fetched_at` | `fetched_at` |
| INDEX | — | なし（上記の UNIQUE KEY を除く） |

## sbi_holding

由来 migration: `006_sbi_holding.sql`、前日比nullable化: `007_nullable_sbi_values.sql`

用途: SBI の各スナップショットに含まれる保有銘柄・商品の数量、単価、評価額および損益を保持する。

SQL 上のテーブル属性: `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`

`snapshot_id` は `sbi_snapshot.id` を指す、DB 外部キー制約のない論理参照です。列の `FK -> sbi_snapshot.id (no DB FK, enforced at app layer)` は内部モデル/SQL COMMENT 上の想定の転記であり、アプリケーション層での強制をこの文書から断定しません。この文書は DB の参照整合性を保証しません。

### Columns

| Column | Type | Null | Default | Description (COMMENT) |
| --- | --- | --- | --- | --- |
| `id` | `INT AUTO_INCREMENT` | No | `—` | `—` |
| `snapshot_id` | `INT` | No | `—` | FK -> sbi_snapshot.id (no DB FK, enforced at app layer) |
| `section` | `VARCHAR(32)` | No | `—` | nisa_domestic\|nisa_us\|nisa_funds\|old_nisa_funds |
| `name` | `TEXT` | No | `—` | Holding.name (銘柄名) |
| `quantity` | `DECIMAL(18,6)` | No | `—` | Holding.quantity (口数/株数) |
| `unit_cost` | `DECIMAL(18,6)` | No | `—` | Holding.unit_cost (取得単価, USD for US stocks, JPY for others) |
| `unit_price` | `DECIMAL(18,6)` | No | `—` | Holding.unit_price (現在値) |
| `prev_day_jpy` | `DECIMAL(14,2)` | Yes | `—` | Holding.prev_day_jpy (per-unit JPY change; per 10k units for funds; NULL when unavailable) |
| `prev_day_pct` | `DECIMAL(10,4)` | Yes | `—` | Holding.prev_day_pct (NULL when unavailable) |
| `pnl_jpy` | `DECIMAL(14,2)` | No | `—` | Holding.pnl_jpy (評価損益 円) |
| `pnl_pct` | `DECIMAL(10,4)` | No | `—` | Holding.pnl_pct (評価損益%) |
| `value_jpy` | `DECIMAL(14,2)` | No | `—` | Holding.value_jpy (評価額 円) |
| `created_at` | `TIMESTAMP` | Yes | `CURRENT_TIMESTAMP` | `—` |
| `updated_at` | `TIMESTAMP` | Yes | `CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | `—` |

### Keys and indexes

| Kind | Name | Columns |
| --- | --- | --- |
| PRIMARY KEY | — | `id` |
| UNIQUE KEY | — | なし |
| INDEX | `idx_snapshot_id` | `snapshot_id` |
| INDEX | `idx_snapshot_section` | `snapshot_id`, `section` |

## 更新手順

1. スキーマ変更では新しい migration を `migration/db/*.sql` に追加し、適用済み migration は変更しない。
2. 未適用の作業中変更であれば、SQL とこの文書を同期する。
3. 文書更新後、`make migration` などでマイグレーションの適用を確認する。


## マイグレーションの運用

- `start` と `sbi-import` だけが処理前に未適用の Up を適用する。`start --dry-run`、API・metrics・maw は適用しない。API 等を先に更新する場合は、対応する新しい importer の `migrate up` を先に実行する。
- `migration/db/*.sql` のファイル名は履歴 ID である。既存 SQL と `gorp_migrations` は変更・初期化しない。SQL の追加後はバイナリを再ビルドする。バイナリにない履歴を持つ DB への適用はエラーとし、自動ダウングレードしない。
- DB 名に対応する名前付きロックを、履歴確認・SQL 適用・履歴更新の全体で保持する。取得待ちは60秒。接続が失われた場合は、ロックを失ったまま再接続して処理を続けない。
- MariaDB の DDL は暗黙コミットされるため、複数ステートメントの途中失敗ではスキーマだけ一部変更され、失敗した migration の履歴が未登録になり得る。自動 Down やアプリ内での適用再試行は行わない。再起動時は未適用分を再確認するため、障害時は Job 等の再試行を止めてから復旧する。
- 失敗時は安全なログの migration ID・MySQL エラー番号・失敗段階を確認し、管理者がスキーマと履歴を照合する。必要に応じてバックアップから戻すか部分適用を修復し、整合が取れてから `migrate up` を再実行する。履歴だけを自動で適用済みにしない。
- 手動ロールバックは `mf-importer migrate down`（既定1件）、件数指定は `--limit N`、全件は `--limit 0`。Down にはテーブル削除や nullable から NOT NULL への変更があるため、実行前にバックアップ・データ条件・アプリ互換性を確認する。アプリの旧版への切替だけで自動 Down はしない。
- 既存の外部マイグレーション Job はこのロックに参加しない。ashley-infra の切替は別作業とし、旧 Job と importer が同時にマイグレーションしないよう調整する。

### 統合テスト

`MF_MIGRATION_TEST_DSN` に明示的に指定した使い捨て MariaDB サーバーでのみ実行する。テストは専用 DB を新規作成・削除するので、テスト用ユーザーにはその権限が必要。実 DB の接続設定は指定しない。

```bash
MF_MIGRATION_TEST_DSN='root:password@tcp(127.0.0.1:3306)/?parseTime=true' \
  go test -v ./internal/migration ./cmd/mf-importer
```

CI は専用 MariaDB コンテナーを使用し、既存履歴からの更新、再実行、Up / Down、排他制御、途中失敗、importer の起動順序を確認する。
