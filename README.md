# mf-importer

マネーフォワードME の家計簿データ（CSV）を取り込んで MariaDB に格納し、REST API と UI で閲覧・管理するためのマイクロサービス群です。

## サービス

| サービス | 役割 |
| --- | --- |
| mf-importer | CSV インポート（重複除外、月末表記変換、S3 ダウンロード対応）と SBI / NRKN JSON 取り込み |
| mf-importer-api | データ閲覧・管理用 REST API |
| mf-importer-maw | [mawinter-server](https://github.com/azuki774/mawinter-server) へのデータ連携 |
| mf-importer-fe | 管理用 Web UI（Nuxt 3） |

## 必要な環境

- Go 1.25 / Node.js / Docker & Docker Compose
- Nix を使う場合は `nix develop`（direnv なら初回のみ `direnv allow`）で必要なツールがそろいます。

## Quickstart

```bash
nix develop # または direnv allow
make build
make start # 全サービス起動（停止は make stop、ログ確認は make debug）
make migration # 内蔵 SQL で DB を更新（API などの起動前にも実行可能）
```

## DB マイグレーション

`mf-importer start` / `sbi-import` / `nrkn-import` は、DB 接続後、ダウンロード・取り込み前に未適用の SQL を自動適用します。`start --dry-run` は適用せず、既存スキーマを前提に動作します。API・maw は自動適用しません。

既存の `migration/db/*.sql` を `go:embed` でバイナリに埋め込み、`github.com/rubenv/sql-migrate` の Go API から実行します。外部マイグレーションバイナリ、実行時の SQL 配置、Git アクセスは不要です。適用対象はビルド時の SQL で、適用履歴は既存の `gorp_migrations` を引き継ぎます。

```bash
make migration # go run ./cmd/mf-importer migrate up
mf-importer migrate up
mf-importer migrate down # 直前の1件を戻す（手動操作のみ）
mf-importer migrate down --limit 0 # 全件を戻す。空の検証 DB 用
```

接続には `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASS` / `DB_NAME` を使い、旧小文字名も引き続き利用できます（大文字優先）。未指定時の値は従来のローカル開発用設定です。

同じ DB への適用は MariaDB の名前付きロックで直列化し、取得待ちは60秒です。失敗時は非ゼロ終了し、取り込みに進みません。

## 開発

```bash
make test # gofmt / vet / staticcheck / go test
```

フロントエンドの開発・動作確認（DB 不要のモック起動含む）は [docs/local-verify.md](docs/local-verify.md) を参照してください。

## Docs

- API ドキュメント: [docs/api.md](docs/api.md)（OpenAPI 定義: `internal/openapi/mfimporter-api.yaml`）
- DB スキーマの説明: [docs/schema.md](docs/schema.md)
- S3 取り込みの詳細: [docs/s3-download.md](docs/s3-download.md)
- SBI JSON取り込み: [docs/sbi-import.md](docs/sbi-import.md)
- NRKN JSON取り込み: [docs/nrkn-import.md](docs/nrkn-import.md)
- ローカル確認: [docs/local-verify.md](docs/local-verify.md)
- 各サービスのコマンド詳細は `--help` を参照してください。

## Renovate の更新方針

- 設定は [`renovate.jsonc`](renovate.jsonc)。通常の更新 PR は毎週金曜日の 20:00〜20:59（`Asia/Tokyo`）に作成します。20:00 ちょうどの実行保証ではなく、Renovate の実行タイミングに依存する時間枠です。
- 通常の更新 PR は同時に最大 7 件で、時間あたりの作成数制限はありません。
- 脆弱性アラートに基づくセキュリティ更新は、通常のスケジュールと同時 7 件の制限を迂回して即時 PR 作成の対象になります。GitHub の dependency graph と Dependabot alerts の有効化、および Renovate がアラートを読み取り PR を作成できる権限が必要です。
- GitHub Actions はコミット SHA（SHA-1）に固定し、Actions の更新だけを自動マージ対象にします。他の依存関係は自動マージしません。
