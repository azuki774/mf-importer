package mfapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mf-importer/internal/model"
	"strconv"
	"strings"
	"time"
)

func snapshotID(source string, id int64) string { return fmt.Sprintf("v1:%s:%020d", source, id) }

func parseSnapshotID(value string) (string, int64, error) {
	p := strings.Split(value, ":")
	if len(p) != 3 || p[0] != "v1" || (p[1] != "sbi" && p[1] != "nrkn") {
		return "", 0, model.ErrInvalidFinancialAssetRequest
	}
	id, err := strconv.ParseInt(p[2], 10, 64)
	if err != nil || id <= 0 || snapshotID(p[1], id) != value {
		return "", 0, model.ErrInvalidFinancialAssetRequest
	}
	return p[1], id, nil
}

func positionID(parts []string) string {
	b, _ := json.Marshal(parts)
	return "v1:" + base64.RawURLEncoding.EncodeToString(b)
}

type financialCursor struct {
	Version   int       `json:"v"`
	Kind      string    `json:"kind"`
	Filter    string    `json:"filter"`
	ID        string    `json:"id,omitempty"`
	FetchedAt time.Time `json:"fetchedAt"`
	Period    string    `json:"period,omitempty"`
}

func financialFilter(sources []string, from, to, interval string, limit int) string {
	b, _ := json.Marshal([]any{sources, from, to, interval, limit})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func encodeFinancialCursor(c financialCursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeFinancialCursor(value, kind, filter string) (financialCursor, error) {
	var c financialCursor
	if len(value) == 0 || len(value) > 2048 {
		return c, model.ErrInvalidFinancialAssetRequest
	}
	b, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil {
		return c, model.ErrInvalidFinancialAssetRequest
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if d.Decode(&c) != nil || c.Version != 1 || c.Kind != kind || c.Filter != filter {
		return c, model.ErrInvalidFinancialAssetRequest
	}
	var trailing any
	if d.Decode(&trailing) != io.EOF {
		return c, model.ErrInvalidFinancialAssetRequest
	}
	return c, nil
}

func containsSource(sources []string, source string) bool {
	for _, s := range sources {
		if s == source {
			return true
		}
	}
	return false
}
func timeKey(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}
