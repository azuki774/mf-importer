# API ドキュメント

## 正本と確認方法

API の正本は `internal/openapi/mfimporter-api.yaml` です。

生成コードは `make generate` で `internal/openapi/*.gen.go` に生成されます。生成コードは手編集しません。

モック API の確認方法は [`docs/local-verify.md`](local-verify.md) を参照してください。`make mock-api` でモック API を起動し、curl で各エンドポイントを確認できます。`make report` では一覧・件数・ルールを集約したサマリを確認できます。

## エンドポイント一覧

| Method | Path | 概要 | Query/Body | 状態 |
|---|---|---|---|---|
| GET | `/details` | 明細一覧取得 | Query: `limit` (integer)、`offset` (integer、default: `0`)、`sort` (string、default: `useDate`、値: `useDate`, `name`, `price`, `registDate`, `importJudgeDate`, `importDate`)、`order` (string、default: `desc`、値: `asc` / `desc`) | 仕様記載あり（200: `Detail` 配列） |
| GET | `/details/count` | 明細総件数取得 | なし | 仕様記載あり（200: `DetailsCount`） |
| GET | `/health` | ヘルスチェック | Body: `text/plain` の string（default: `OK`） | 仕様要確認（YAML の `responses` が空） |
| GET | `/details/{id}` | 明細取得 | Path: `id` (integer、必須) | WIP（YAML summary。200: `Detail`） |
| PATCH | `/details/{id}?ope=reset` | 明細の状態変更 | Path: `id` (integer、必須)。Query: `ope` (string、必須、`reset`)。Body: YAML 上は `content: {}` | 仕様記載あり（200: OK、400: 未知の操作名） |
| DELETE | `/details/{id}` | 明細削除 | Path: `id` (integer、必須) | WIP（YAML summary。204） |
| GET | `/histories` | 取り込み履歴取得 | なし | WIP（YAML summary）。現状のモックは空レスポンス |
| GET | `/rules` | 抽出ルール一覧取得 | Query: `sort` (string、default: `id`、値: `id`, `fieldName`, `value`, `exactMatch`, `categoryId`)、`order` (string、default: `asc`、値: `asc` / `desc`) | 仕様記載あり（200: `Rule` 配列） |
| POST | `/rules` | 抽出ルール追加 | Body: `application/json` の `RuleRequest` | 仕様記載あり（201: `Rule`） |
| GET | `/rules/{id}` | 抽出ルール取得 | Path: `id` (integer、必須) | 仕様記載あり（200: `Rule`） |
| DELETE | `/rules/{id}` | 抽出ルール削除 | Path: `id` (integer、必須)。Body: YAML 上は `content: {}` | 仕様記載あり（204） |

`/details/{id}` の PATCH は、仕様上 required な `ope` に `reset` を指定します。未知の操作名は 400 です。

## スキーマ概要

型と必須項目は YAML の定義に従います。YAML に `nullable` の指定はありません。`required` に含まれないプロパティの未設定時の表現は YAML では定義されていません。

- `Detail`: `id` (integer)、`useDate` (string, date)、`name` (string)、`price` (integer)、`registDate` (string, date-time)、`importJudgeDate` (string, date-time)、`importDate` (string, date-time)。必須: `id`, `useDate`, `name`, `price`, `registDate`。`importJudgeDate` と `importDate` は必須指定なし。
- `DetailsCount`: `count` (integer)。必須: `count`。
- `Rule`: `id` (integer)、`fieldName` (string)、`value` (string)、`categoryId` (integer)、`exactMatch` (integer, 0-1)。必須: `id`, `fieldName`, `value`, `categoryId`, `exactMatch`。
- `RuleRequest`: `fieldName` (string)、`value` (string)、`categoryId` (integer)、`exactMatch` (integer, 0-1)。必須: `fieldName`, `value`, `categoryId`, `exactMatch`。

## 更新手順

1. `internal/openapi/mfimporter-api.yaml` を編集する。
2. `make generate` を実行して `internal/openapi/*.gen.go` を生成する。生成コードは手編集しない。
3. この案内表とスキーマ概要を YAML に合わせて更新する。
