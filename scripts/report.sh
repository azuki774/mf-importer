#!/bin/sh
# mf-importer の取り込み結果を 1 つの JSON サマリとして出力する
# 使い方:
#   URL=http://127.0.0.1:8080 LATEST=5 ./scripts/report.sh
#   ./scripts/report.sh URL=http://127.0.0.1:8080/api LATEST=10
# 必要コマンド: curl, jq
set -eu

command -v curl >/dev/null 2>&1 || { echo "curl is required" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "jq is required" >&2; exit 1; }

API_URL="${URL:-http://127.0.0.1:8080}"
LATEST="${LATEST:-5}"
PAGE_SIZE="${PAGE_SIZE:-100}"

# 引数の key=value も環境変数と同じ扱いにする
for arg in "$@"; do
    case "$arg" in
        URL=*) API_URL="${arg#URL=}" ;;
        LATEST=*) LATEST="${arg#LATEST=}" ;;
        PAGE_SIZE=*) PAGE_SIZE="${arg#PAGE_SIZE=}" ;;
        *) echo "unknown argument: $arg (expected KEY=VALUE)" >&2; exit 1 ;;
    esac
done

API_URL="${API_URL%/}"

get() {
    curl -fsS "${API_URL}$1"
}

acc_file="$(mktemp)"
page_file="$(mktemp)"
trap 'rm -f "$acc_file" "$page_file" "$acc_file.tmp"' EXIT
echo "[]" >"$acc_file"

offset=0
while :; do
    get "/details?limit=${PAGE_SIZE}&offset=${offset}" >"$page_file"
    page_len="$(jq 'length' "$page_file")"
    jq -s '.[0] + .[1]' "$acc_file" "$page_file" >"$acc_file.tmp"
    mv "$acc_file.tmp" "$acc_file"
    if [ "$page_len" -lt "$PAGE_SIZE" ]; then
        break
    fi
    offset=$((offset + PAGE_SIZE))
done

total_count="$(get '/details/count' | jq '.count')"
rules_count="$(get '/rules' | jq 'length')"
monthly_summary="$(jq 'group_by(.useDate[:7]) | map({yyyymm: .[0].useDate[:7], count: length, sum_price: (map(.price) | add // 0)})' "$acc_file")"
latest_details="$(jq --argjson n "$LATEST" '.[:$n]' "$acc_file")"
fetched_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

jq -n \
    --arg api_url "$API_URL" \
    --arg fetched_at "$fetched_at" \
    --argjson total_count "$total_count" \
    --argjson monthly_summary "$monthly_summary" \
    --argjson latest_details "$latest_details" \
    --argjson rules_count "$rules_count" \
    '{api_url: $api_url, fetched_at: $fetched_at, total_count: $total_count, monthly_summary: $monthly_summary, latest_details: $latest_details, rules_count: $rules_count}'
