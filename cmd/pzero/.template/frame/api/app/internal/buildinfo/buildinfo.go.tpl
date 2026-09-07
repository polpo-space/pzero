package buildinfo

// 构建元数据，由 -ldflags -X 注入。本地 go run 时使用下方默认值。
// 示例：
//
//	go build -ldflags "-X '{{.Module}}/internal/buildinfo.Version=v0.1.0' \
//	  -X '{{.Module}}/internal/buildinfo.Commit=$(git rev-parse --short HEAD)' \
//	  -X '{{.Module}}/internal/buildinfo.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" .
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
