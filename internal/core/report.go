package core

import (
	"fmt"
	"os"
)

func ExportStatusReport(path string, build BuildInfo, doctor DoctorReport, health map[string]any) error {
	body := "# AIOS Status Report\n\n"
	body += fmt.Sprintf("- Version: %s\n- Commit: %s\n- Build Date: %s\n", build.Version, build.Commit, build.BuildDate)
	body += "\n## Doctor\n"
	state := "PASS"
	if !doctor.Overall {
		state = "FAIL"
	}
	body += fmt.Sprintf("Overall: %s\n", state)
	for _, c := range doctor.Checks {
		mark := "PASS"
		if !c.OK {
			mark = "FAIL"
		}
		body += fmt.Sprintf("- %s %s (%s)\n", mark, c.Name, c.Detail)
	}
	body += "\n## Health\n"
	for k, v := range health {
		body += fmt.Sprintf("- %s: %v\n", k, v)
	}
	return os.WriteFile(path, []byte(body), 0o600)
}
