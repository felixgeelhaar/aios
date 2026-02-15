package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/felixgeelhaar/aios/internal/runtime"
)

// fileExecutionReportStore implements runtime.ExecutionReportStore using the
// local filesystem for writing structured execution reports.
type fileExecutionReportStore struct{}

var _ runtime.ExecutionReportStore = fileExecutionReportStore{}

func (fileExecutionReportStore) WriteReport(path string, report runtime.ExecutionReport) error {
	body, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}
