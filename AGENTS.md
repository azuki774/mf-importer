# Repository Guidelines

このファイルは、`mf-importer` の開発・レビュー・コミット時に守るルールです。
リポジトリ固有のルールを優先して確認し、作業開始前に現在の差分とブランチを確認してください。

## 作業場所とブランチ

- Git の編集は worktree で行う。標準の配置先は `~/src/worktrees/mf-importer/<worktree-name>`。
- `master` / `main` へ直接コミット・マージしない。目的ごとにブランチを作り、PR で統合する。
- 他の作業者や worktree の変更を巻き戻さない。自分の変更対象でない差分はそのまま残す。
- 生成物、ログ、実データはリポジトリ外または `.gitignore` 対象の場所に置く。

## 構成

- `cmd/mf-importer/`: MoneyForward ME の CSV インポーターと S3 ダウンロード。
- `cmd/mf-importer-api/`: DB API と DB 不要のモック API。
- `cmd/mf-importer-maw/`: `mawinter-server` 連携。
- `cmd/mf-importer-metrics/`: Prometheus メトリクス。
- `internal/`: サービス共通のモデル、リポジトリ、サービス、HTTP サーバー、SBI 資産処理など。
- `internal/openapi/mfimporter-api.yaml`: API の入力仕様。`*.gen.go` は `make generate` の生成物で手編集しない。
- `migration/db/`: sql-migrate の適用ファイル。現在の資産系スキーマを含む。`docs/schema.md` はこのディレクトリを正本とする手管理文書。
- `deployment/`: Docker Compose と抽出ルール。`build/`: Dockerfile とバイナリ出力先。
- `frontend/`: Nuxt 3 + Tailwind CSS の管理 UI。Bootstrap 前提で変更しない。
- `test/`: 明示的に作成したダミー CSV / JSON / SQL とモックサーバー。新しい fixture は実データでないことを目視確認する。
- `frontend/e2e/`: Playwright の画面テスト。成果物は `frontend/e2e/artifacts/`（git 管理外）。
- `scripts/`: ローカル確認用スクリプト。
- `.agents/skills/`: エージェントがデータを扱うときに読むリポジトリ固有スキル。

## 開発環境

- 推奨は `nix develop`。direnv を使う場合は初回だけ `direnv allow` を実行する。
- `flake.nix` は Go、Node.js、Docker、MariaDB クライアント、`gh`、静的解析ツールを提供する。
- 開発用 MariaDB は `.devcontainer/docker-compose.yml` または `deployment/compose.yml` で起動する。既存のローカル開発用デフォルト値を実運用 credential として再利用せず、実運用の認証情報は環境変数で渡す。
- `flake.nix` を変更した場合は、依存関係の更新要否を確認し、必要なら `flake.lock` も更新する。
- Go の `version ... does not match go tool version ...` が出た場合は、Nix の Go と mise 等の `GOROOT` が混在している。`env -u GOROOT -u GOTOOLDIR make test` で再実行する。

## ビルド・テスト・確認

- `make bin`: 全 Go サービスの静的バイナリを `build/bin/` に生成する。
- `make api-bin`: モック API / ローカル UI 用の API バイナリを生成する。
- `make build`: 各サービスの Docker イメージをビルドする。
- `make test`: `gofmt -l`、`go vet`、`staticcheck`、`go test -v ./...` を実行する。`gofmt -l` の出力は空にする。
- `go test ./...`: DB を使わない Go テストを直接実行する。
- `make generate`: `internal/openapi/mfimporter-api.yaml` から OpenAPI コードを再生成する。`docs/api.md` はこの YAML を正本とする手管理文書。
- `make migration`: ローカル MariaDB にマイグレーションを適用する。スキーマ変更時は up / down の両方を確認する。
- `make start` / `make debug` / `make stop`: Docker Compose の全サービスを起動、前景起動、停止する。
- `make mock-api`: `test/` のダミー CSV を使う DB 不要の API を起動する。
- `make local-ui`: フロントエンドを静的生成し、モック API と同一プロセスで配信する。
- `make report`: モック API の API 結果を一時的な確認用 JSON に集約する。出力をコミット・PR・Issue に貼らない。
- `make e2e`: Playwright の画面テストを実行する。初回は `cd frontend && npm install` が必要。
- `cd frontend && npm ci`: `package-lock.json` に従って依存関係を再現する。
- `cd frontend && npm run dev`: 開発サーバーを起動する。`npm run build`、`npm run generate`、`npm run preview` も利用する。

## 実データ・秘匿情報の取り扱い

詳細な判断基準は `.agents/skills/no-real-money-data/SKILL.md` に集約しています。家計・資産・SBI のデータに触れる作業では、編集前に必ずそのスキルを読みます。

- 本番・個人の MoneyForward / SBI データを、コード、テスト、ドキュメント、ログ、Issue、PR、スクリーンショットへ追加しない。
- 具体的な残高、損益、取引金額、数量、単価、日付と明細内容の組み合わせを記録しない。説明では `<AMOUNT>`、`<DATE>`、`<TICKER>` などのプレースホルダーを使う。
- 実在する銘柄名、証券コード、ファンド名、口座番号、支店名、金融機関名と資産情報の組み合わせを記録しない。
- MF / SBI から取得した CSV・JSON・API レスポンスを全文・一部とも転載しない。S3 ダウンロード先や実 CSV は `/tmp/` などリポジトリ外に置く。
- `make report`、`curl`、DB クエリの出力をファイル保存・貼付しない。必要な例は手書きのダミー JSON に置き換える。
- 画面画像は `make mock-api`、`make local-ui`、`make e2e` のダミー表示だけを使う。実 DB 接続中の画面を撮影しない。
- 新しい fixture や unit test 内の入力例は仮名・合成値で作り、実在の銘柄・口座・明細を元にしない。既存の `test/` fixture や `*_test.go` の入力例に実データを追記しない。
- AWS キー、実運用 DB パスワード、`.env`、秘密鍵はコミットしない。ローカル設定は `.env.local` など git 管理外に置く。Compose / Nix にある既存の開発用デフォルト値はローカル用途に限定し、実値へ置き換えない。
- スキルを読んでも安全が保証されるわけではない。`git diff --cached` を必ず目視し、疑わしい差分は除外してから進める。

## コーディング規約

- Go は gofmt、go vet、staticcheck に合わせる。パッケージ名は小文字、公開 API は必要な型・関数に限定する。
- Go のエラーは文脈を保ち、外部 I/O と DB の境界では適切にラップする。ログに入力データや credential を出さない。
- API 仕様を変更するときは YAML を先に更新し、`make generate` で生成コードを更新する。生成ファイルを直接編集しない。
- SQL は既存の sql-migrate 形式に従い、Up / Down を同じ変更単位で用意する。
- TypeScript / Vue は Composition API と既存の composable / component 構成に合わせる。スタイルは Tailwind の既存ユーティリティを優先する。
- フロントエンドの依存関係を変更したら `package-lock.json` も同時に更新し、`npm ci` と build / generate を確認する。

## コミットと PR

- コミットメッセージは Conventional Commits の英語・命令形を使う（例: `docs: refresh repository guidelines`、`api: add mock endpoint`）。
- 1 コミットの目的を小さく保ち、生成コードを変更した場合は元仕様と生成コマンドを明記する。
- PR は日本語で目的、変更内容、影響範囲、確認コマンドを記載する。UI 変更はモック画面のスクリーンショットを添付する。
- スキーマ変更は適用手順とロールバック手順を記載する。S3 / 外部 API 変更は必要な環境変数と失敗時の挙動を記載する。
- PR 作成・確認には `gh` を使う。CI の Go と Frontend が通ることを確認してからレビューを依頼する。

## コミット前チェックリスト

```bash
git status --short
git diff --cached --stat
git diff --cached
make test
```

データを含む差分では、上記に加えて、実データ・具体的な金額・銘柄情報・秘密情報がないことを目視で確認します。
