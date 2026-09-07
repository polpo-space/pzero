{{ if has "job" .Features }}package job

import (
	runtimejob "github.com/polpo-space/pzero/runtime/job"

	"{{ .Module }}/internal/svc"
)

// Registry 返回所有 job handler。名字必须与 etc.yaml 里 job.jobs 的 key
// 完全一致：配置里有的这里必须有，这里有的配置里也必须有，否则启动失败。
func Registry(svcCtx *svc.ServiceContext) []runtimejob.NamedHandler {
	example := NewExampleJob(svcCtx)

	return []runtimejob.NamedHandler{
		runtimejob.Named("exampleInterval", example.ExampleInterval),
		runtimejob.Named("exampleMinute", example.ExampleMinute),
	}
}
{{ end -}}
