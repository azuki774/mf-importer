# S3 Downloader

`mf-importer start --with-download` で、S3 上の CSV を `-- -d` で指定したディレクトリへダウンロードする。

## インプット

| 環境変数 | 必須 | 用途 |
|---|---|---|
| `AWS_ACCESS_KEY_ID` | ○ | 静的クレデンシャル |
| `AWS_SECRET_ACCESS_KEY` | ○ | 静的クレデンシャル |
| `AWS_REGION` | ○ | バケットのリージョン |
| `BUCKET_NAME` | ○ | 取得元バケット名 |
| `BUCKET_DIR` | ○ | バケット内のプレフィックス (例: `cfo/`) |
| `BUCKET_URL` | × | MinIO など S3 互換エンドポイントの URL |

## アウトプット

`<SaveDir>/` 直下に、S3 キーからディレクトリ部分を除いた **ファイル名のみ** を保存する。

```
BUCKET_NAME = my-bucket
BUCKET_DIR  = cfo/
SaveDir     = /data/

s3://my-bucket/cfo/cf.csv             -> /data/cf.csv
s3://my-bucket/cfo/2025/01/old.csv    -> /data/old.csv
s3://my-bucket/cfo/sub/a.csv          -> /data/a.csv
s3://my-bucket/cfo/sub/b.csv          -> /data/b.csv
```

`SaveDir` が存在しなければ `0755` で作成する。

## 制約

- **1 リクエストで最大 1000 件**。1000 件超のバケットはページネーションせず 1001 件目以降を破棄し、`l.Warn` 1 行だけ出す
- **キー末尾が `/` のオブジェクト** はスキップ
- **キーが `BUCKET_DIR` と完全一致するオブジェクト** はスキップ
- **逐次ダウンロード** (並列度 1)。ファイル 1 件ずつ `GetObject` → ファイルへ書き出し
- `BUCKET_URL` 設定時は path-style アドレスが有効になる。未設定なら virtual-hosted-style (AWS 標準)
