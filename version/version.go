package version

import (
	_ "embed"
	"strings"
)

//go:embed version
var Version string

// GitCommit、BuildStamp 由 Makefile「ios」「android」目标的 gomobile -ldflags -X 注入。
// 若日志里仍是旧提交或 notset，说明 App 未装入本次编出的 xcframework/AAR。
var GitCommit = "notset"
var BuildStamp = "notset"

func init() {
	Version = strings.Trim(Version, "\n")
	Version = strings.TrimSpace(Version)
}
