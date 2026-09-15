# NRKN JSON 取り込み

`mf-importer nrkn-import` は、[myscrapers PR #60](https://github.com/azuki774/myscrapers/pull/60) の NRKN 日次 JSON を `nrkn_snapshot` と `nrkn_holding` へ保存します。
入力仕様は [NRKN の実装](https://github.com/azuki774/myscrapers/blob/4c9390ff5cfe03355656df7bfea2ab6e23302d9f/myscraper/internal/nrkn/fetch.go) を参照してください。

## 実行方法

```bash
# 東京の現在月を S3 から取り込む
mf-importer nrkn-import
# 指定月を取り込む
mf-importer nrkn-import --month 200001
# ローカルの JSON を再帰的に取り込む
mf-importer nrkn-import --input-dir /tmp/nrkn-json
```

`--input-dir` と `--month` は併用できません。ローカルモードは S3 に接続しません。
S3 の取得範囲は `<NRKN_BUCKET_DIR>/<YYYY>/<MM>/` です。JSON のみ取得し、ページ送りにも対応します。
一時ディレクトリは `0700`、ファイルは `0600` で作成し、終了時に一時ディレクトリを削除します。

| 環境変数 | 必須 | 用途 |
| --- | --- | --- |
| `NRKN_BUCKET_NAME` | 必須 | NRKN のバケット |
| `NRKN_BUCKET_DIR` | 必須 | NRKN JSON のルートプレフィックス |
| `AWS_REGION` | 必須 | リージョン |
| `NRKN_BUCKET_URL` | 任意 | S3 互換エンドポイント。省略時は `BUCKET_URL`、両方なければ AWS 標準 |
| `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | 条件付き | 静的認証情報を使う場合。両方が揃わなければ SDK の既定の認証方法を使用 |

バケット名・プレフィックスは NRKN 専用設定が必要です。`BUCKET_NAME` / `BUCKET_DIR` へのフォールバックはありません。
DB 接続は既存の `DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASS`、`DB_NAME` を使います。

## 保存とエラー処理

- 対応する `schema_version` は文字列 `2026-09-14`。未知の文字列版は警告し、そのファイルをスキップします。
- バージョンの欠落・`null`・型不一致、壊れた JSON、対応版の必須項目不備はエラーです。
- `status` は `ok`（大小文字を区別しない）のみ保存します。エラーやメンテナンスをゼロ残高として保存しません。
- `fetched_at` は東京時刻・マイクロ秒精度に揃え、同じ日時の取り込みをスキップします。
- 合計値、商品ごとの数量・評価額・取得価額累計・損益・解約価額・解約時評価額・構成比を保存します。合計や評価額を再計算しません。
- 商品コードの先頭ゼロ、基準日、価格の元表記（`*` を含む）を保持します。基準日と取得日時は別々に保存します。
- `composite_figi` は12文字の ASCII 英大文字・数字を必須にします。同じ取得日時内の商品コードの重複はエラーです。異なる商品コードが同じ FIGI を持つ場合は保存できます。
- 必須数値の欠落・`null` を拒否し、明示されたゼロを保存します。空の商品一覧はエラーです。
- 1ファイルの合計と明細は同じトランザクションで保存します。明細保存の失敗時は合計もロールバックします。
- ファイルはパスの辞書順に処理します。取得・解析・DB エラーで停止し、それ以前に保存したファイルは保持します。対象 JSON がなければエラーです。
- 完了ログはファイル数・新規登録数・重複数・未対応版数のみです。入力 JSON や資産明細をログに出しません。

## マイグレーション

起動時に未適用の SQL が自動適用されます。手動では `mf-importer migrate up` を実行します。
`009_nrkn_snapshot.sql` は NRKN 用の2テーブルと商品コードの一意制約、FIGI の検索インデックスを追加します。

直前が `009_nrkn_snapshot.sql` の場合、`mf-importer migrate down --limit 1` で戻せます。
Down は NRKN の明細・合計テーブルを削除するため、保存データが必要なら先に退避してください。SBI のテーブルには影響しません。

## 合成データでの確認

`test/nrkn_example.json` は手書きの合成データです。ローカル確認では、このファイルだけを専用ディレクトリに配置してください。
`test/` 全体には他の形式の JSON もあるため、NRKN の入力ディレクトリとして指定しないでください。

```bash
make test
# 使い捨て MariaDB を指定した場合のみ DB 統合テストも実行
MF_MIGRATION_TEST_DSN='<TEST_DSN>' go test ./internal/migration -run Integration
```
