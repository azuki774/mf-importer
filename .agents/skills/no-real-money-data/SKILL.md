---
name: no-real-money-data
description: Use when handling MoneyForward ME or SBI data, fixtures, API responses, logs, screenshots, commits, or PRs where real amounts, securities, account information, or credentials could be exposed.
---

# no-real-money-data — 実マネーデータをコミットしない

このリポジトリは MoneyForward ME / SBI 証券の実データを扱う。
具体的すぎる数字・銘柄情報が git 履歴に入ると削除が困難になるため、
コミット・PR・レビュー・スクショ貼付の前に本スキルを適用する。

## 1. コミット禁止データ（実データ）

以下に該当するものは、コード・ドキュメント・テスト・Issue・PR・スクショの
いずれにも含めない:

- 具体的な金額・残高・損益・前日比・数量・単価（円・株・口座残高など）
- 銘柄名・証券コード・ファンド名（NISA / 特定口座 / 現物 / 投資信託を含む）
- 金融機関名・支店名・口座番号・口座種別と金額の組み合わせ
- 取引日・内容・メモの組み合わせで本人特定や資産推定が可能な明細行
- MF / SBI からエクスポート・ダウンロードした CSV / JSON の実データ全文または一部転載
- 実データが写った管理 UI・API レスポンス・ログ・スクリーンショット
- AWS 認証情報・実運用 DB パスワード等のシークレット（`.env.local` 等の管理外ファイルに置く）。Compose / Nix にある既存の開発用デフォルト値は実運用値として扱わない。

疑わしい場合は「実データ」として扱い、含めない判断を優先する。

## 2. コミット許可データ（ダミーのみ）

- 既存の合成 fixture（`test/cf.csv`、`test/cf_lastmonth.csv`、`test/sbi_example_new.json` 等）。新しいデータファイルは自動許可されない。
- `deployment/extract_rule.csv` のルール定義
- 明らかに仮名と分かる例（`テスト明細 001`、`ダミー銘柄A`、`テストカード` 等）
- 金額を丸めた・マスクした・桁数を変えた説明用の例（例: `約 10 万円`、`**** 円`）

新しい fixture を追加する場合は、既存の命名（`テスト〜` / `ダミー〜`）に寄せ、
実在の銘柄・金額に見えない値を使う。

## 3. 執筆ルール

- Issue / PR / docs の例示は、マスク・丸め・ダミー置換のいずれかを行う。
- API レスポンス (`/details`、`/rules` 等) の実データをそのまま貼らない。
  必要な形だけ `make report` のモック結果か、手書きのダミー JSON で示す。
- スクリーンショットは `make mock-api` / `make local-ui` / `make e2e` の
  モック表示のみ使う。実 DB 接続中の画面は撮影しない。
- ローカル検証用の実 CSV は `/tmp/mf-csv/` 等のリポジトリ外に置き、
  `cp` で使う。リポジトリ内にコピーしない。
- `memo/`、`tmp/`、`.env.local`、鍵ファイルはコミットしない
  （`.gitignore` 済み。`git status` で混入を確認する）。

## 4. コミット前のチェック（必須）

コミットする前に以下をすべて行う:

```bash
git status --short
git diff --cached --stat
git diff --cached  # 実金額・銘柄・口座情報が含まれていないか目視する
./scripts/check-sensitive-data.sh --staged
```

`check-sensitive-data.sh` が警告・失敗した場合は、そのファイルを含めずに
差分を作り直す。スクリプトは補助であり、通過＝安全ではない。
最終判断は必ず人間の目視で行う。

## 5. 違反を見つけたとき

- コミット前: 該当ファイルを `git restore --staged` で除外し、作業ツリー側も削除・ダミー化する。
- プッシュ前: 該当コミットを履歴から除去する手順を実行する前に、共有ブランチや他の worktree への影響を確認する。
- プッシュ後: 無理に force-push せず、秘匿化の手順（履歴書換え・ローテーション）を
  別途相談する。S3・DB の認証情報は直ちにローテーションする。

## 6. 自動ガード

- `make test` は `check-sensitive-data.sh --all` を先に実行する。
- PR / master への push では GitHub Actions が base 側の検査スクリプトで差分を検査する。
- CI では候補の警告も失敗になる。fixture を変更した場合も、合成データであることを確認して検査対象から除外するのではなく、差分を見直して必要最小限にする。
- guard の workflow / script 自体は同じ PR で変更しない。更新が必要な場合は別 PR として、検査を無効化しないことを目視確認する。
- 検査はヒューリスティックであり、fixture や新しいデータ形式を完全には判定できない。
  警告・成功にかかわらず、staged 差分の目視確認を省略しない。
