#!/usr/bin/env bash
# 実マネーデータの混入を差分から簡易検査する（ヒューリスティック）。
# 使い方:
#   ./scripts/check-sensitive-data.sh --staged
#   ./scripts/check-sensitive-data.sh --all
#   ./scripts/check-sensitive-data.sh --range BASE...HEAD
# CI では警告も失敗にする。ローカルで明示する場合は SENSITIVE_DATA_STRICT=1。
# 詳細ルールは .agents/skills/no-real-money-data/SKILL.md を参照。
set -uo pipefail

MODE=staged
RANGE=
while (($# > 0)); do
  case "$1" in
    --staged)
      MODE=staged
      ;;
    --all)
      MODE=all
      ;;
    --range)
      if (($# < 2)); then
        echo "--range requires BASE...HEAD" >&2
        exit 2
      fi
      MODE=range
      RANGE=$2
      shift
      ;;
    -h|--help)
      echo "usage: $0 [--staged|--all|--range BASE...HEAD]"
      exit 0
      ;;
    *)
      echo "unknown argument: $1 (expected --staged|--all|--range BASE...HEAD)" >&2
      exit 2
      ;;
  esac
  shift
done

case "$MODE" in
  staged) DIFF_ARGS=(--cached) ;;
  all)    DIFF_ARGS=(HEAD) ;;
  range)  DIFF_ARGS=("$RANGE") ;;
esac

FAIL=0
WARN=0
STRICT=${SENSITIVE_DATA_STRICT:-false}
if [ "${CI:-false}" = true ]; then
  STRICT=true
fi
TMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/mf-sensitive.XXXXXX")
trap 'rm -rf "$TMP_DIR"' EXIT

fail() {
  printf '[FAIL] %s\n' "$1" >&2
  FAIL=1
}

warn() {
  printf '[WARN] %s\n' "$1"
  WARN=1
}

is_allowlisted() {
  case "$1" in
    .agents/skills/no-real-money-data/*|scripts/check-sensitive-data.sh|deployment/extract_rule.csv|test/cf.csv|test/cf_lastmonth.csv|test/sbi_example_new.json|test/testjson.json|test/seed_100_details.sql|test/mock_server.py)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

is_untracked() {
  local candidate
  for candidate in "${UNTRACKED_FILES[@]}"; do
    if [ "$candidate" = "$1" ]; then
      return 0
    fi
  done
  return 1
}

git_diff_names() {
  local output="$1"
  if ! git diff "${DIFF_ARGS[@]}" --name-only -z --diff-filter=ACMRT -- >"$output"; then
    echo "check-sensitive-data: git diff failed ($MODE); refusing to treat it as clean" >&2
    exit 2
  fi
}

git_diff_body() {
  local output="$1"
  shift
  if (($# == 0)); then
    : >"$output"
    return
  fi
  if ! git --literal-pathspecs diff --no-ext-diff --no-color "${DIFF_ARGS[@]}" -U0 -- "$@" >"$output"; then
    echo "check-sensitive-data: unable to read git diff body ($MODE)" >&2
    exit 2
  fi
}

append_untracked_diff() {
  local output="$1"
  shift
  local file rc
  for file in "$@"; do
    git --literal-pathspecs diff --no-index --no-ext-diff --no-color -U0 -- /dev/null "$file" >>"$output"
    rc=$?
    # git diff --no-index returns 1 when files differ, which is expected here.
    if ((rc > 1)); then
      echo "check-sensitive-data: unable to read untracked file: $file" >&2
      exit 2
    fi
  done
}

added_lines() {
  # Keep lines beginning with '+' but drop only the standard diff file header.
  grep -E '^\+' "$1" | grep -Ev '^\+\+\+ (b/|/dev/null)' || true
}

check_pattern() {
  local body="$1"
  local label="$2"
  local pattern="$3"
  local level="$4"
  local matches

  # Never print matching content: CI logs must not become a second data leak.
  matches=$(grep -Ec -- "$pattern" <<<"$body" || true)
  if [ "$matches" -eq 0 ]; then
    return
  fi
  if [ "$level" = fail ]; then
    fail "$label"
    printf '[INFO] matching added lines: %s\n' "$matches" >&2
  else
    warn "$label"
    printf '[INFO] matching added lines: %s\n' "$matches"
  fi
}

NAMES_FILE="$TMP_DIR/names"
UNTRACKED_FILE="$TMP_DIR/untracked"
NEW_NAMES_FILE="$TMP_DIR/new-names"
ALL_DIFF_FILE="$TMP_DIR/all-diff"
TARGET_DIFF_FILE="$TMP_DIR/target-diff"
git_diff_names "$NAMES_FILE"
: >"$UNTRACKED_FILE"
: >"$NEW_NAMES_FILE"

if [ "$MODE" = all ]; then
  if ! git ls-files --others --exclude-standard -z >"$UNTRACKED_FILE"; then
    echo "check-sensitive-data: unable to list untracked files" >&2
    exit 2
  fi
fi

mapfile -d '' -t TRACKED_FILES <"$NAMES_FILE"
mapfile -d '' -t UNTRACKED_FILES <"$UNTRACKED_FILE"
FILES=("${TRACKED_FILES[@]}")
if [ "$MODE" = all ]; then
  FILES+=("${UNTRACKED_FILES[@]}")
fi

if ((${#FILES[@]} == 0)); then
  printf 'check-sensitive-data: no changed files (%s). OK\n' "$MODE"
  exit 0
fi

# リポジトリ外に置くべき一時ファイル・秘密鍵の混入は常に失敗させる。
for file in "${FILES[@]}"; do
  case "$file" in
    .env.example)
      ;;
    memo|memo/*|tmp/*|.env|.env.*|*.pem|*.key|*_rsa|*_ed25519)
      fail "forbidden path changed: $file（リポジトリ外に置き、除外すること）"
      ;;
  esac
done

TARGETS=()
TARGET_TRACKED=()
TARGET_UNTRACKED=()
for file in "${FILES[@]}"; do
  if ! is_allowlisted "$file"; then
    TARGETS+=("$file")
    if [ "$MODE" = all ] && is_untracked "$file"; then
      TARGET_UNTRACKED+=("$file")
    else
      TARGET_TRACKED+=("$file")
    fi
  fi
done

# 追加行だけを検査する。allow-list はダミー fixture の内容検査を除外するが、
# high-confidence な秘密情報検査は全変更ファイルに対して行う。
git_diff_body "$ALL_DIFF_FILE" "${TRACKED_FILES[@]}"
if [ "$MODE" = all ]; then
  append_untracked_diff "$ALL_DIFF_FILE" "${UNTRACKED_FILES[@]}"
fi
if ((${#TARGET_TRACKED[@]} > 0)); then
  git_diff_body "$TARGET_DIFF_FILE" "${TARGET_TRACKED[@]}"
else
  : >"$TARGET_DIFF_FILE"
fi
if [ "$MODE" = all ]; then
  append_untracked_diff "$TARGET_DIFF_FILE" "${TARGET_UNTRACKED[@]}"
fi

ALL_DIFF_BODY=$(added_lines "$ALL_DIFF_FILE")
TARGET_DIFF_BODY=$(added_lines "$TARGET_DIFF_FILE")

# 秘密鍵・AWS credential・口座番号の値は fixture でも許可しない。
CREDENTIAL_PATTERN='AKIA[0-9A-Z]{16}|ASIA[0-9A-Z]{16}|BEGIN[[:space:]].*PRIVATE KEY|([Aa][Ww][Ss]_?[Ss][Ee][Cc][Rr][Ee][Tt]_?[Aa][Cc][Cc][Ee][Ss][Ss]_?[Kk][Ee][Yy])[[:space:]]*[\"]?[[:space:]]*[:=][[:space:]]*[\"]?[A-Za-z0-9/+=]{16,}|(password|passwd|token|secret|api[_-]?key|client[_-]?secret)[[:space:]]*[\"]?[[:space:]]*[:=][[:space:]]*[\"]?[A-Za-z0-9._~+/=-]{12,}|口座番号[：:][[:space:]]*[\"]?[0-9]{6,}|account[ _-]?(number|no)[\"]?[[:space:]]*[:=][[:space:]]*[\"]?[0-9]{6,}'
check_pattern "$ALL_DIFF_BODY" \
  "credential detected（秘密情報を削除し、必要ならローテーションする）" \
  "$CREDENTIAL_PATTERN" \
  fail

if [ "$MODE" = range ]; then
  HISTORY_RANGE="$RANGE"
  if [[ "$RANGE" == *...* ]]; then
    HISTORY_RANGE="${RANGE%%...*}..${RANGE##*...}"
  fi
  HISTORY_FILE="$TMP_DIR/history"
  if ! git log --full-history --diff-merges=separate --no-ext-diff --no-color --format= -p "$HISTORY_RANGE" -- >"$HISTORY_FILE"; then
    echo "check-sensitive-data: git history scan failed; refusing to treat it as clean" >&2
    exit 2
  fi
  HISTORY_DIFF_BODY=$(added_lines "$HISTORY_FILE")
  check_pattern "$HISTORY_DIFF_BODY" \
    "credential detected in commit history（履歴から削除し、必要ならローテーションする）" \
    "$CREDENTIAL_PATTERN" \
    fail
fi

# 既存 fixture の変更は候補を警告し、目視確認を要求する。
check_pattern "$ALL_DIFF_BODY" \
  "suspicious transaction row（具体的な日付×金額の実データをダミー化する）" \
  '"202[0-9]/[0-9]{1,2}/[0-9]{1,2}"[^\n]*"-?[0-9]{3,}"|202[0-9]/[0-9]{1,2}/[0-9]{1,2}.*,.*[0-9]{4,}|202[0-9]-[0-9]{2}-[0-9]{2}[^\n]*[0-9]{4,}' \
  warn
check_pattern "$ALL_DIFF_BODY" \
  "suspicious security value（銘柄名・証券コード・数量・単価を実値で貼らない）" \
  '証券コード[：:][[:space:]]*[0-9]{4,5}|銘柄名[：:]|"quantity"[[:space:]]*:[[:space:]]*[0-9]+|"unit_price"[[:space:]]*:[[:space:]]*[0-9]+' \
  warn

# fixture 以外で高確度の転載形式を検出した場合は失敗させる。
check_pattern "$TARGET_DIFF_BODY" \
  "raw transaction export outside fixture（実 CSV / JSON の転載ではなく手書き fixture を使う）" \
  '計算対象.*金額（円）|"holdings"[[:space:]]*:[[:space:]]*\[' \
  fail
check_pattern "$TARGET_DIFF_BODY" \
  "security identifier outside fixture（実在の銘柄名・証券コードを含めない）" \
  '証券コード[：:][[:space:]]*[0-9]{4,5}|銘柄名[：:][[:space:]]*[^<[:space:]]' \
  fail

if ! git diff "${DIFF_ARGS[@]}" --name-only -z --diff-filter=A -- >"$NEW_NAMES_FILE"; then
  echo "check-sensitive-data: unable to list new files ($MODE)" >&2
  exit 2
fi
if [ "$MODE" = all ]; then
  cat "$UNTRACKED_FILE" >>"$NEW_NAMES_FILE"
fi
mapfile -d '' -t NEW_FILES <"$NEW_NAMES_FILE"

# 許可した fixture 以外に新しい表形式データや画像を増やさない。
for file in "${NEW_FILES[@]}"; do
  case "$file" in
    test/*|deployment/extract_rule.csv)
      warn "new fixture requires manual review: $file（実データでないことを確認する）"
      ;;
    *.csv|*.json|*.tsv|*.png|*.jpg|*.jpeg|*.webp)
      fail "new data or image file outside fixture allow-list: $file"
      ;;
  esac
done

if [ "$FAIL" -eq 1 ]; then
  echo "check-sensitive-data: FAILED（.agents/skills/no-real-money-data/SKILL.md に従い差分を作り直す）" >&2
  exit 1
fi
if [ "$WARN" -eq 1 ]; then
  if [ "$STRICT" = 1 ] || [ "$STRICT" = true ]; then
    echo "check-sensitive-data: FAILED because warnings are strict in CI" >&2
    exit 1
  fi
  echo "check-sensitive-data: warnings（目視で実データ混入がないことを確認してから進める）"
  exit 0
fi
printf 'check-sensitive-data: OK (%s)\n' "$MODE"
