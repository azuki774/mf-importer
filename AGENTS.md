# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` に各サービスのエントリポイント、`internal/` に共通ロジックと OpenAPI 生成物（`internal/openapi/*.gen.go` は手編集しない）を配置。
- `build/` は Dockerfile 群と生成バイナリ出力先 `build/bin/`、`deployment/compose.yml` で開発用スタックを起動。
- `migration/` は SQL マイグレーション、`vendor_ci/sql-migrate` 経由で適用。`docs/` は API ドキュメント、`frontend/` は Nuxt 3 + Bootstrap UI。`test/` は共通テストリソース。
- 生成物やログはリポジトリ外に置き、`git status` がクリーンになるよう `.gitignore` に従う。
- `flake.nix` は Nix 開発環境の定義。`.envrc` で direnv 連携。`.direnv/` / `result*` は自動生成なので `.gitignore` に登録済み。

## Build, Test, and Development Commands

### Nix 開発環境（推奨）
- direnv を使用する場合、ディレクトリに入ると自動的に Nix 環境が有効化される（`direnv allow` が初回のみ必要）。
- 手動で Nix 環境を有効化する場合は `nix develop` を実行（Go 1.25、Node 22、gh、docker、golangci-lint 等が利用可能になる）。
- `flake.nix` を編集した場合は `nix flake update` で依存関係を更新し、`flake.lock` を更新してコミット。
- DB（MariaDB）は Nix 外で Docker コンテナとして実行。`.devcontainer/docker-compose.yml` で定義され、`docker-compose -f .devcontainer/docker-compose.yml up -d` で起動。

### ビルド・テスト・開発コマンド
- `make bin` 静的リンクの Go バイナリを `build/bin/` に生成（タグ・リビジョンを埋め込み）。
- `make build` 各サービスの Docker イメージを構築。CI と同じ環境で確認したいときに使用。
- `make start` / `make debug` / `make stop` で docker compose による全サービス起動・前景実行・停止。
- `make migration` で MariaDB へローカルマイグレーション、`make generate` で OpenAPI コード再生成、`make doc` で HTML ドキュメント更新。
- フロントエンドは `cd frontend && npm install` 後、`npm run dev` で開発、`npm run build` / `npm run generate` / `npm run preview` でビルド・静的生成・プレビュー。

## Coding Style & Naming Conventions
- Go は gofmt / go vet / staticcheck 準拠。小文字パッケージ名、受け渡しが必要な型のみエクスポート。
- テストデータや環境変数名は snake_case、構造体フィールドは Go の慣例に合わせる。
- フロントエンドは Composition API 推奨、型は TypeScript を想定。スタイルは Bootstrap 5 のユーティリティを優先し、独自 CSS はスコープを限定。

## Testing Guidelines
- Go の最低限は `go test ./...`。事前チェックとして `make test` を推奨（gofmt / vet / staticcheck / go test をまとめて実行）。
- マイグレーションを伴う変更は `make migration` で適用テストし、ロールバック手順を PR に記載。
- フロントエンドは自動テスト未整備のため、主要ページを `npm run preview` で手動確認し、重要ロジックには E2E 追加を検討。

## Commit & Pull Request Guidelines
- コミットメッセージは英語・命令形で 50–72 文字目安（例: `Add maw export filter`）。複数サービスに渡る場合はプレフィックスで明示（例: `api: ...`, `fe: ...`）。
- PR は日本語で概要・動機・影響範囲を記載し、関連 Issue、スクリーンショット（UI 変更時）、`make test` / 必要な `npm run build` の結果を添付。
- 破壊的変更やスキーマ変更はマイグレーション手順とロールバック方法を明記。生成コード更新時は元 YAML と実行コマンドを併記。
- GitHub へのプッシュや PR 作成は `gh` コマンド（例: `gh pr create --fill`）を利用し、テンプレートに沿ってメタデータを整える。

## Security & Configuration Tips
- DB 接続は `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASS` / `DB_NAME` を利用。`--with-download` 使用時は AWS 資格情報を環境変数かプロファイルで注入。
- 秘密情報や本番データのコミット禁止。ローカル設定は `.env.local` など git 管理外に置く。
- S3 取り込みや外部連携（mawinter-server）はネットワーク前提のため、スタブが必要な場合は別環境でテストする。
