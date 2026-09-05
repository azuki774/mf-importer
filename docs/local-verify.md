# ローカルでのインポート結果確認ガイド

DB 不要のモック API で mf-importer の取り込み結果を確認する手順。

## 最短の流れ

ターミナルを 2 つ開いて、3 ステップです。

### 1. モック API を立てる（ターミナル1）

```bash
make mock-api
```

`http://127.0.0.1:8080` で REST API が起動する（`test/` 配下の CSV を fixture 化）。

### 2. フロントエンドを立てる（ターミナル2）

```bash
cd frontend && npm install && npm run dev   # npm install は初回のみ
```

`http://localhost:3000` で UI が起動する（API は Nitro が :8080 に中継するので設定不要）。

別マシン（コンテナ / WSL など）からアクセスする場合は以下:

```bash
cd frontend && npm run dev:host   # 0.0.0.0:3000 で待ち受け (http://<ホストIP>:3000)
```

API 側は `make mock-api` が既定で全インタフェース待ち受け（`:8080`）なのでそのままでよい。

### 3. 開く

- **人間**: ブラウザで `http://localhost:3000` を開く。取り込み履歴・ルール設定の閲覧と操作（再判定 / ルール追加・削除）ができる
- **AI / CLI**: 同じ API に対して curl で JSON を取得する

```bash
curl -s http://127.0.0.1:8080/details/count
curl -s "http://127.0.0.1:8080/details?limit=5&offset=0"
curl -s http://127.0.0.1:8080/rules
make report   # 上記を集約したサマリ JSON を 1 つにまとめて出力
```

### 止めるとき

両ターミナルで `Ctrl-C`。

---

## それ以外のやり方

| コマンド | 内容 |
|---|---|
| `make local-ui` | 上記 1・2 を 1 コマンドにしたもの（静的ビルド UI を `:8080` で配信、ブラウザで `http://127.0.0.1:8080` を開く。API は `/api/*` 配下） |
| `make e2e` | Playwright が画面操作を自動テストし、`frontend/e2e/artifacts/*.png` にスクリーンショットを保存する（AI は画像を読める。初回のみ `cd frontend && npm install`） |
| `make report` | `GET /details`, `/details/count`, `/rules` を集約した JSON サマリを出力（`URL=` `LATEST=` で対象変更） |

ポートが塞がっている場合の例:

```bash
./build/bin/mf-importer-api mock --addr :8000          # モック API を :8000 へ
cd frontend && NUXT_PUBLIC_API_BASE_ENDPOINT=http://127.0.0.1:8000 npm run dev
```

### make e2e の検証内容

モック API + 静的ビルド UI に対して Playwright がブラウザ操作を行い、以下を検証する:

- 取り込み履歴テーブルへの fixture 表示（件数・並び・金額フォーマット・import 判定日時の有無）
- ページネーション（10 ページ構成での前後移動、端ページでのボタン無効化、表示件数変更による再取得）
- 「再判定」操作（確認ダイアログ → PATCH → 一覧反映 → トースト表示）
- ルール一覧表示・追加・削除

E2E でも `test/cf.csv`（200 件 = 10 ページ構成）をそのまま使い、ページネーションが発生する状態で検証する。テスト本体は `frontend/e2e/*.spec.ts`、設定（ポートなど）は `frontend/playwright.config.ts`。


## モックデータの仕様

- `--input-dir`（既定 `./test/`）配下の `*.csv` を読む。パースに失敗したファイルは warn して skip（例: `test/cf_lastmonth.csv` は日付形式が異なるため対象外）
- detail は日付昇順に `id = 1..` を採番（新規取り込み直後の DB と同じ並び）。金額は importer 本体と同じく符号を除いた値
- 3 件に 1 件は mawinter 連携済みの状態（`importJudgeDate` / `importDate` 設定済み）
- rules は固定 2 件（`deployment/extract_rule.csv` と同じ内容）。書き込み操作もメモリ上で動作し、プロセス再起動で初期化される
- `GET /histories` は未実装（常時空レスポンス）

## 付録: 実 DB (MariaDB) での確認

```bash
docker compose -f deployment/compose.yml up -d mf-importer-db
make migration

mkdir -p /tmp/mf-csv && cp test/cf.csv /tmp/mf-csv/
./build/bin/mf-importer start -d /tmp/mf-csv/     # make bin でバイナリ生成が必要
./build/bin/mf-importer-api start                  # :8080
```

接続先は `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASS` / `DB_NAME`（旧小文字名 `db_host` 等も利用可）。API 仕様は [docs/api.md](api.md)（仕様の正本: `internal/openapi/mfimporter-api.yaml`）を参照。
