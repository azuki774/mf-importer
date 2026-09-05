# mf-importer

マネーフォワードME の家計簿データ（CSV）を取り込んで MariaDB に格納し、REST API と UI で閲覧・管理するためのマイクロサービス群です。

## サービス

| サービス | 役割 |
| --- | --- |
| mf-importer | CSV インポート（重複除外、月末表記変換、S3 ダウンロード対応） |
| mf-importer-api | データ閲覧・管理用 REST API |
| mf-importer-maw | [mawinter-server](https://github.com/azuki774/mawinter-server) へのデータ連携 |
| mf-importer-metrics | Prometheus メトリクス提供 |
| mf-importer-fe | 管理用 Web UI（Nuxt 3） |

## 必要な環境

- Go 1.25 / Node.js / Docker & Docker Compose
- Nix を使う場合は `nix develop`（direnv なら初回のみ `direnv allow`）で必要なツールがそろいます。

## Quickstart

```bash
nix develop # または direnv allow
make build
make start # 全サービス起動（停止は make stop、ログ確認は make debug）
make migration # DB マイグレーション（初回のみ）
```

## 開発

```bash
make test # gofmt / vet / staticcheck / go test
```

フロントエンドの開発・動作確認（DB 不要のモック起動含む）は [docs/local-verify.md](docs/local-verify.md) を参照してください。

## Docs

- API ドキュメント: [docs/api.md](docs/api.md)（仕様の正本: `internal/openapi/mfimporter-api.yaml`）
- DB スキーマの説明: [docs/schema.md](docs/schema.md)
- S3 取り込みの詳細: [docs/s3-download.md](docs/s3-download.md)
- ローカル確認: [docs/local-verify.md](docs/local-verify.md)
- 各サービスのコマンド詳細は `--help` を参照してください。
