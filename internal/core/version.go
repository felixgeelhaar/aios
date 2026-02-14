package core

var (
	Version   = "0.1.0"
	Commit    = "dev"
	BuildDate = "unknown"
)

type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

func CurrentBuildInfo() BuildInfo {
	return BuildInfo{Version: Version, Commit: Commit, BuildDate: BuildDate}
}
