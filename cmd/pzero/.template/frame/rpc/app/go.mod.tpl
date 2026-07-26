module {{ .Module }}

go {{ .GoVersion }}

{{if (VersionCompare .GoVersion ">=" "1.24")}}
tool (
	github.com/polpo-space/pzero/cmd/pzero
)
{{end}}