# Docker イメージの master マージ時公開設計

## 目的

`master` にマージされた変更のうち、`docs/` 配下だけの変更を除き、既存の 5 つの Docker イメージを GitHub Container Registry へ公開する。公開イメージはコミットを特定できる短縮コミットハッシュでタグ付けする。

## 対象範囲

- 対象ブランチは `master`。
- `docs/**` のみが変更された push は対象外とする。
- `docs/**` とコードなどが同時に変更された push は対象とする。
- 対象イメージは既存の次の 5 つとする。
  - `ghcr.io/azuki774/mf-importer`
  - `ghcr.io/azuki774/mf-importer-maw`
  - `ghcr.io/azuki774/mf-importer-api`
  - `ghcr.io/azuki774/mf-importer-metrics`
  - `ghcr.io/azuki774/mf-importer-fe`
- 既存の `v*` タグ push による semver / `latest` 公開は維持する。

## 設計

既存の `.github/workflows/publish.yml` を拡張する。`push` トリガーに `master` ブランチと `paths-ignore: ['docs/**']` を加え、現在の `v*` タグトリガーを残す。GitHub Actions のパスフィルターは変更された全パスが除外対象の場合だけ workflow をスキップするため、`docs/` と他のパスが混在する変更は実行される。

Docker metadata のタグ生成に短縮 SHA タグを追加する。master push では各イメージに短縮コミットハッシュを付与し、既存の v* タグ push では従来どおり semver と `latest` を生成する。タグ生成は既存の `docker/metadata-action` に集約し、5 ジョブで同じ方式を利用する。

各ジョブの build context、Dockerfile、`linux/amd64`、GHCR ログイン、`packages: write` 権限は変更しない。

## エラー処理と権限

- build、registry login、push のいずれかが失敗した場合は Actions job を失敗させる。
- GitHub token は既存どおり `secrets.GITHUB_TOKEN` を使用し、workflow に新しい秘密情報を追加しない。
- workflow の実行対象を `master` と v* に限定し、docs-only push では不要な publish を行わない。

## 確認方法

- YAML の構文と差分を確認する。
- 変更内容が master push、docs-only 除外、docs 混在、v* タグ維持、短縮 SHA タグの要件を満たすことを静的に確認する。
- 既存の `make test` を実行し、workflow 変更によるリポジトリ側の回帰がないことを確認する。
