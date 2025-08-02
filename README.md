# mf-importer

マネーフォワードME の家計簿データをデータベースに取り込み、管理・分析するためのマイクロサービス群です。

## 概要

mf-importerは、マネーフォワードMEから出力されたCSVファイルを解析してMariaDBに格納し、REST APIやフロントエンドUIを通じてデータの閲覧・管理を行うことができます。また、外部サービス（mawinter-server）との連携機能も提供しています。

## アーキテクチャ

本プロジェクトは以下の5つのサービスから構成されています：

### 🔄 mf-importer
メインのCSVインポートサービス
- マネーフォワードMEの家計簿データ（CSV）をDBに取り込み
- S3からのCSVファイル自動ダウンロード対応
- 重複データの検出・除外
- 月末表記（例：2025-05月末）の自動日付変換
- 資産履歴データの管理

### 🌐 mf-importer-api
REST APIサービス
- データ閲覧・管理用のREST API提供
- OpenAPI仕様に基づいた自動生成コード
- 抽出ルールの管理機能
- インポート履歴の管理
- APIドキュメント: https://azuki774.github.io/mf-importer/api.html

### 🔗 mf-importer-maw
外部サービス連携
- DBから条件に合うデータを抽出
- [mawinter-server](https://github.com/azuki774/mawinter-server) との連携
- 抽出ルールに基づいた自動データ転送

### 📊 mf-importer-metrics
メトリクス監視サービス
- Prometheusメトリクスの提供
- システム監視・運用支援

### 🖥️ mf-importer-fe
フロントエンドUI（Nuxt.js）
- インポート状況の可視化
- 抽出ルールの管理画面
- Bootstrap使用のレスポンシブデザイン

## 必要な環境

- **Go**: 1.23.0以上
- **Node.js**: フロントエンド開発用
- **Docker & Docker Compose**: コンテナ実行環境
- **MariaDB**: 10.x（Docker Composeで自動起動）

## セットアップ

### 1. リポジトリのクローン
```bash
git clone <repository-url>
cd mf-importer
```

### 2. 依存関係のインストール
```bash
# Go modules
go mod download

# フロントエンド依存関係
cd frontend
npm install
cd ..
```

### 3. データベースの準備
```bash
# DBマイグレーション実行
make migration
```

### 4. ビルド
```bash
# バイナリビルド
make bin

# Dockerイメージビルド
make build
```

## 使用方法

### Docker Composeを使用した起動

#### 全サービス起動（推奨）
```bash
make start
```

#### デバッグモード（ログ表示）
```bash
make debug
```

#### 停止
```bash
make stop
```

### 個別サービスの起動

#### mf-importer（CSVインポート）
```bash
# 基本的なインポート
./build/bin/mf-importer start -d /path/to/csv/

# S3からダウンロード後インポート
./build/bin/mf-importer start --with-download -d /data/

# ドライランモード（実際の更新なし）
./build/bin/mf-importer start --dry-run -d /path/to/csv/
```

#### mf-importer-api（REST APIサーバー）
```bash
./build/bin/mf-importer-api start
# ポート8080でAPIサーバーが起動
```

#### mf-importer-maw（外部連携）
```bash
# 通常実行
./build/bin/mf-importer-maw regist

# ドライランモード
./build/bin/mf-importer-maw regist --dry-run
```

### 環境変数

各サービスで使用可能な環境変数：

#### データベース接続
- `DB_HOST`: データベースホスト（デフォルト: 127.0.0.1）
- `DB_PORT`: データベースポート（デフォルト: 3306）
- `DB_USER`: データベースユーザー（デフォルト: root）
- `DB_PASS`: データベースパスワード（デフォルト: password）
- `DB_NAME`: データベース名（デフォルト: mfimporter）

#### mf-importer-maw固有
- `api_uri`: mawinter-serverのAPIエンドポイント
- `jobname`: ジョブ名（インポート履歴用）

#### AWS S3（--with-downloadオプション使用時）
- AWS認証情報（AWS CLI設定またはIAMロール）

## 対応データ形式

### 家計簿データ（cf.csv, cf_lastmonth.csv）
```csv
計算対象,日付,内容,金額（円）,保有金融機関,大項目,中項目,メモ,振替,ID
,07/16(火),ローソン,-291,三井住友カード,食費,食料品,,,
```

### 資産履歴データ（asset_history.csv）
```csv
日付,合計,預金・現金・暗号資産,株式(現物),投資信託,ポイント,詳細
2025-08-02,"5,000,000","3,500,000",0,"1,400,000","100,000",テスト詳細
2025-05月末,"4,800,000","3,200,000","500,000","1,000,000","100,000",5月末テスト詳細
```

**特徴:**
- 文字エンコーディング: Shift_JIS
- 月末表記（YYYY-MM月末）の自動変換対応
- 重複データの自動検出・除外

## 開発コマンド

### テスト実行
```bash
make test
```

### API仕様書生成
```bash
make generate  # OpenAPIコード生成
make doc       # HTML形式のAPIドキュメント生成
```

### フロントエンド開発
```bash
cd frontend
npm run dev        # 開発サーバー起動
npm run build      # 本番ビルド
npm run generate   # 静的生成
npm run preview    # 本番プレビュー
```

## データベーススキーマ

### detail（家計簿データ）
- 日付、内容、金額、カテゴリ等の家計簿情報
- 重複検出用のユニークキー

### asset_history（資産履歴）
- 日付別の資産内訳
- 預金・現金・暗号資産、株式、投資信託、ポイント等

### extract_rule（抽出ルール）
- mawinter-server連携用のデータ抽出条件
- フィールド名、値、完全一致/部分一致の設定

### import_history（インポート履歴）
- ファイル別のインポート実行履歴
- 処理済みエントリ数、新規追加数の記録

## API仕様

REST APIの詳細仕様は以下で確認できます：
- **APIドキュメント**: https://azuki774.github.io/mf-importer/api.html
- **OpenAPI仕様書**: `internal/openapi/mfimporter-api.yaml`

主要エンドポイント：
- `GET /details`: 家計簿データ一覧
- `GET /rules`: 抽出ルール一覧
- `POST /rules`: 抽出ルール追加
- `DELETE /rules/{id}`: 抽出ルール削除

## 外部連携

### mawinter-server連携
- extract_ruleテーブルの条件に基づいてdetailテーブルからデータを抽出
- RESTful APIでmawinter-serverにデータを送信
- ドライランモードでの事前確認が可能

## 技術スタック

### バックエンド
- **言語**: Go 1.23.0
- **フレームワーク**: Chi (HTTP router)
- **CLI**: Cobra
- **ORM**: GORM
- **ログ**: Zap
- **API生成**: oapi-codegen
- **テスト**: 標準testing + httpmock

### フロントエンド
- **フレームワーク**: Nuxt.js 3
- **UI**: Bootstrap 5
- **言語**: TypeScript

### インフラ
- **データベース**: MariaDB 10
- **コンテナ**: Docker, Docker Compose
- **監視**: Prometheus metrics
- **ストレージ**: AWS S3（オプション）

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**
   - MariaDBコンテナが起動しているか確認
   - 環境変数の設定を確認

2. **CSVファイルが認識されない**
   - ファイルエンコーディングがShift_JISであることを確認
   - ファイル名が規定の形式（cf.csv, asset_history.csv等）であることを確認

3. **月末日付の変換エラー**
   - 「YYYY-MM月末」の形式であることを確認
   - 有効な年月であることを確認

### ログの確認
```bash
# デバッグモードでログ表示
make debug

# 特定サービスのログ確認
docker logs mf-importer-api
```
