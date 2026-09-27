# API ドキュメント

## API 仕様

API は `internal/openapi/mfimporter-api.yaml` に定義されています。

生成コードは `make generate` で `internal/openapi/*.gen.go` に生成されます。

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
| GET | `/financial-assets/snapshots` | 金融資産スナップショット一覧（summary のみ） | Query: 繰返し可 `source` (`sbi` / `nrkn`、省略時は両方)、`from` / `to` (RFC 3339、開始を含み終了を含まない)、`limit` (default: 100、最大500)、`cursor` | 設計済み・未実装 |
| GET | `/financial-assets/snapshots/{snapshotId}` | スナップショットと保有明細 | Path: `snapshotId` (opaque ID) | 設計済み・未実装 |
| GET | `/financial-assets/balances` | 現在・指定時点残高、または日次/月次推移 | Query: `source` 任意、単一時点は `at` (RFC 3339、inclusive、省略時は現在)。期間は `from` / `to` (Asia/Tokyo の日付、`[from,to)`、両方必須)、`interval` (`day` / `month`、default: `day`)、`limit`、`cursor` | 設計済み・未実装 |

金融資産 API の金額・数量は decimal string とし、未取得・不明値は `null` とします。スナップショットは不変で、一覧はメタデータと totals を返します。`fetchedAt` はデータ取得時刻、`importedAt` はシステムへの取り込み時刻です。スナップショット ID は安定した不透明 ID です。

残高応答は常に `items` 配列と nullable な `nextCursor` を持ち、現在/指定時点モードは1点、期間モードは日・月ごとに1点です。データがない期間も点を返し、合計を null として `missingSources` を設定します。各点は単一時点の `timestamp`、または期間の `periodStart`/`periodEnd` を使い、使わないフィールドは null です。期間は Asia/Tokyo の境界を使い、各点で期間終了時刻未満の最新スナップショットを採用して前方補完します。月途中の期間指定では境界を指定範囲に合わせます。`at` と期間指定は併用できず、`interval`/`limit`/`cursor` は期間指定時のみ有効です。

ソースまたは値が欠ける場合、合計は項目ごとに `null` となります。`missingSources` は該当するスナップショット自体がないソースだけを列挙します。カーソルはフィルターとページ条件に紐づき、継続ページ間でデータ集合は固定されません。現在のサーバーでは、これらのエンドポイントは `501` を返します。

`/details/{id}` の PATCH は、仕様上 required な `ope` に `reset` を指定します。未知の操作名は 400 です。

## スキーマ概要

型と必須項目は YAML の定義に従います。YAML に `nullable` の指定はありません。`required` に含まれないプロパティの未設定時の表現は YAML では定義されていません。

- `Detail`: `id` (integer)、`useDate` (string, date)、`name` (string)、`price` (integer)、`registDate` (string, date-time)、`importJudgeDate` (string, date-time)、`importDate` (string, date-time)。必須: `id`, `useDate`, `name`, `price`, `registDate`。`importJudgeDate` と `importDate` は必須指定なし。
- `DetailsCount`: `count` (integer)。必須: `count`。
- `Rule`: `id` (integer)、`fieldName` (string)、`value` (string)、`categoryId` (integer)、`exactMatch` (integer, 0-1)。必須: `id`, `fieldName`, `value`, `categoryId`, `exactMatch`。
- `RuleRequest`: `fieldName` (string)、`value` (string)、`categoryId` (integer)、`exactMatch` (integer, 0-1)。必須: `fieldName`, `value`, `categoryId`, `exactMatch`。
