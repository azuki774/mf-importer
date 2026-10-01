package mfapi

import (
	"fmt"
	"mf-importer/internal/model"
	"strconv"
	"strings"
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
