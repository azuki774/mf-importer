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

- JSON内の `fetched_at` を取り込みIDとして使用する。`Asia/Tokyo` の時刻へ変換し、DBのマイクロ秒精度へ正規化する。
- 同じ `fetched_at` が登録済みなら正常にスキップする。ファイル名やパスの変更は重複判定に影響しない。
- `OK` のJSONはスナップショットと保有情報を保存する。
- `MAINTENANCE` / `ERROR` のJSONはスナップショットだけ保存する。
- 1ファイルのスナップショットと保有情報は同一トランザクションで保存する。
- 複数ファイルは辞書順に処理し、解析・取得・DBエラーが起きた時点で停止する。
- 先に正常登録したファイルは保持され、失敗したファイルの途中状態はロールバックされる。

終了時には対象ファイル数、新規登録数、重複スキップ数をログへ出力する。資産金額、数量、銘柄名などの明細はログへ出力しない。

DB接続には既存コマンドと同じ `DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASS`、`DB_NAME` 環境変数を使う。
