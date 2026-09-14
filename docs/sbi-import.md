# SBI JSON取り込み

`mf-importer sbi-import` は、SBIの資産スナップショットJSONを `sbi_snapshot` と `sbi_holding` へ取り込む。

SBI JSONの仕様・生成処理は [`azuki774/myscrapers` のSBI実装](https://github.com/azuki774/myscrapers/blob/master/myscraper/internal/sbi/fetch.go) を参照する。サンプルJSONは [`example-assets.json`](https://github.com/azuki774/myscrapers/blob/master/myscraper/internal/sbi/testdata/example-assets.json) にある。

## S3から取り込む

通常はS3モードを使う。

```bash
mf-importer sbi-import
```

引数を省略すると、`Asia/Tokyo` の現在月を対象にして、次のプレフィックスだけを取得する。

```text
<SBI_BUCKET_DIR>/<YYYY>/<MM>/
```

過去月を再処理する場合は `YYYYMM` 形式で指定する。

```bash
mf-importer sbi-import --month 202608
```

取得したJSONは一時ディレクトリへ保存し、DB取り込み後に削除する。対象月にJSONがない場合はエラー終了する。

### S3環境変数

必須欄の凡例: `○` は必須、`△` は条件付き必須、`×` は任意です。

| 環境変数 | 必須 | 用途 |
| --- | --- | --- |
| `SBI_BUCKET_NAME` | ○ | SBI JSONを保存するバケット名 |
| `SBI_BUCKET_DIR` | ○ | SBI JSONのルートプレフィックス |
| `AWS_REGION` | ○ | バケットのリージョン |
| `AWS_ACCESS_KEY_ID` | △ | 静的認証情報を使う場合のアクセスキー |
| `AWS_SECRET_ACCESS_KEY` | △ | 静的認証情報を使う場合のシークレットキー |
| `SBI_BUCKET_URL` | × | S3互換ストレージのエンドポイント |

`SBI_*` のバケット設定が未指定の場合は、既存の汎用 `AWS_*` / `BUCKET_*` 設定も後方互換として参照する。

## ローカルJSONを取り込む

ローカル確認や復旧では `--input-dir` を指定する。指定ディレクトリ配下のJSONを再帰的に処理し、月計算やS3アクセスは行わない。

```bash
mf-importer sbi-import --input-dir /path/to/2026/08
```

`--input-dir` と `--month` は併用できない。JSONが1件もない場合はエラー終了する。

## 取り込みの仕様

- `schema_version` は文字列の `2026-09-12` のみ受け付けます。欠落、数値、旧形式、未対応の文字列は、statusにかかわらず取り込みません。
- JSON内の `fetched_at` を取り込みIDとして使用する。`Asia/Tokyo` の時刻へ変換し、DBのマイクロ秒精度へ正規化する。
- 同じ `fetched_at` が登録済みなら正常にスキップする。ファイル名やパスの変更は重複判定に影響しない。
- 1ファイルのスナップショットと保有情報は同一トランザクションで保存する。
- 複数ファイルは辞書順に処理し、解析・取得・DBエラーが起きた時点で停止する。
- 旧形式のJSONが1件でも混在すると、既に同じ取得時点がDBに登録済みでも、重複判定より前の解析で停止します。
- 先に正常登録したファイルは保持され、失敗したファイルの途中状態はロールバックされる。

### `status = OK`

- スナップショットと保有明細の必須資産値を検証します。
- 各保有明細の `composite_figi` は必須で、12文字のASCII英大文字・数字だけを受け付けます。
- 同じ区分内の `composite_figi` 重複は取り込みエラーとします。区分が異なる場合や取得時点が異なる場合は保存できます。
- 欠損があれば取り込みエラーとし、スナップショットを保存しません。
- 明示的なゼロはゼロとして保存し、`NULL` のまま保存しません。

### `status = MAINTENANCE / ERROR`

- 取得試行の記録として `sbi_snapshot` を保存します。
- すべての資産値を `NULL` とします。
- `sbi_holding` は保存しません。

### 保有米国株の前日比

- `OK` でも生成元の仕様上取得できないため、`prev_day_jpy` と `prev_day_pct` だけを `NULL` とします。

終了時には対象ファイル数、新規登録数、重複スキップ数をログへ出力する。資産金額、数量、銘柄名などの明細はログへ出力しない。

DB接続には既存コマンドと同じ `DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASS`、`DB_NAME` 環境変数を使う。

`008_sbi_schema_version_figi.sql` のDownは、`schema_version` が32bit整数へ変換できない行、またはFIGIが設定された保有明細がある場合に停止します。先に該当データを確認し、必要なら別途退避してから実行してください。

起動時には未適用マイグレーションが自動適用されます。手動で確認する場合は `mf-importer migrate up --limit 1` を、直前のマイグレーションを戻す場合は `mf-importer migrate down --limit 1` を使います。バックアップを作成してもDownの実行条件は変わりません。
