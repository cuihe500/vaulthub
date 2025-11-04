package version

import (
	"fmt"
	"runtime"
)

var (
	// 以下变量在构建时通过 ldflags 注入
	Version   = "dev"             // 版本号
	GitCommit = "unknown"         // Git commit hash
	BuildTime = "unknown"         // 构建时间
	GoVersion = runtime.Version() // Go 版本
)

// Info 返回版本信息结构
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get 获取版本信息
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String 返回格式化的版本字符串
func (i Info) String() string {
	return fmt.Sprintf(
		"Version:    %s\nGit Commit: %s\nBuild Time: %s\nGo Version: %s\nPlatform:   %s",
		i.Version,
		i.GitCommit,
		i.BuildTime,
		i.GoVersion,
		i.Platform,
	)
}
