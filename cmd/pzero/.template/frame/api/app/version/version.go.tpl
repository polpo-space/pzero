package version

// 构建元数据，由 -ldflags -X 注入。本地 go run 时使用下方默认值。
// 示例：
//
//	go build -ldflags "-X '{{.Module}}/version.Version=v0.1.0' \
//	  -X '{{.Module}}/version.Commit=$(git rev-parse --short HEAD)' \
//	  -X '{{.Module}}/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" .
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
