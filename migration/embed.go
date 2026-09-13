// Package migration は、DB マイグレーションに使う SQL ファイルを保持します。
package migration

import "embed"

// Files には、このファイルと同じディレクトリにある db/*.sql
// （リポジトリ内の migration/db/*.sql）が入ります。
// 下の go:embed は、Go のビルド時に SQL ファイルの内容をバイナリ内部へ取り込む指定です。
// 実行時はここから SQL を読み込めるため、バイナリの隣に SQL ファイルを置く必要はありません。
// Dockerfile はこのバイナリをイメージへコピーするので、SQL も一緒に含まれます。
// そのため、Dockerfile で SQL ファイルを個別にコピーする必要もありません。
// SQL ファイルを追加・変更した場合は、バイナリをビルドし直すと反映されます。
//
//go:embed db/*.sql
var Files embed.FS
