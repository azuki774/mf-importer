# Docker イメージの選択的公開設計

## 目的

`master` への push で、変更の影響を受ける Docker イメージだけを GitHub Container Registry へ公開する。`master` と `v*` タグ push のいずれにも短縮コミット SHA タグを付け、`v*` タグ push では既存の semver / `latest` タグも維持する。

## 対象イメージ

- `ghcr.io/azuki774/mf-importer`
- `ghcr.io/azuki774/mf-importer-maw`
- `ghcr.io/azuki774/mf-importer-api`
- `ghcr.io/azuki774/mf-importer-metrics`
- `ghcr.io/azuki774/mf-importer-fe`

## イベント別の動作

### master push

workflow の先頭に変更検出ジョブを置き、各イメージ用の真偽値を job output として公開する。後続の5つの build-and-push job は対応する output が `true` の場合だけ実行する。

`docker/metadata-action` は `type=sha,format=short,prefix=` を master と `v*` タグ push の両方で有効にする。これにより、両イベントから公開されるタグは `sha-` prefix のない短縮 SHA になる。

### v* タグ push

変更検出結果にかかわらず5イメージをすべて公開する。短縮 SHA タグに加えて、既存の version、major.minor、major、`latest` タグを維持する。

## 変更検出

`dorny/paths-filter` v4.0.3 をコミット SHA `ceb8a2b8f2d89434be7ff52d3de7ec3738c5cc9d` に固定して利用する。master push の `before` と現在の SHA を比較し、次のルールで対象を判定する。

| 変更パス | 再ビルド対象 |
| --- | --- |
| `frontend/**`, `build/fe/Dockerfile` | frontend |
| `cmd/mf-importer/**`, `build/Dockerfile` | importer |
| `cmd/mf-importer-maw/**`, `build/maw/Dockerfile` | maw |
| `cmd/mf-importer-api/**`, `build/api/Dockerfile` | api |
| `cmd/mf-importer-metrics/**`, `build/metrics/Dockerfile` | metrics |
| `internal/**`, `go.mod`, `go.sum` | Go 4イメージ |
| `.dockerignore`, `.github/workflows/publish.yml` | 5イメージすべて |

`internal/**` は複数コマンドから利用される共有コードである。import graph を workflow 内で動的解析すると削除・移動・依存追加時の判定が複雑になるため、安全側に倒して Go 4イメージを再ビルドする。ドキュメント、テスト fixture、deployment 設定など、Docker build context の成果物に影響しない変更では build-and-push job を起動しない。

## 構成と権限

- 変更検出ジョブは checkout と path filter のみを実行する。
- 既存5ジョブの build context、Dockerfile、`linux/amd64`、GHCR login、`packages: write` は維持する。
- secrets は既存の `GITHUB_TOKEN` だけを利用する。
- build、login、push の失敗は従来どおり該当 job の失敗として扱う。

## 検証

- `actionlint` で workflow の YAML、式、job dependency を検証する。
- 差分を目視し、各フィルターと各 build job の対応、master と v* のタグ動作を確認する。
- `make test` を実行してリポジトリ全体の回帰がないことを確認する。

workflow YAML の構成変更であり、実行環境そのものをローカル単体テストで再現できないため、TDD の configuration-file 例外を適用する。代わりに `actionlint` と静的な対応関係の確認を完了条件とする。
