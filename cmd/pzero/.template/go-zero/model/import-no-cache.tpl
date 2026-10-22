import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"slices"
	{{if .time}}"time"{{end}}

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/huandu/go-sqlbuilder"
	"github.com/polpo-space/pzero/core/stores/condition"
	"github.com/polpo-space/pzero/core/stores/modelx"
	"github.com/spf13/cast"
    "github.com/eddieowens/opts"
	{{$pqInThird := false}}{{range .data.Fields}}{{if eq .ThirdPkg "github.com/lib/pq"}}{{$pqInThird = true}}{{end}}{{end}}
	{{if and .containsPQ (not $pqInThird)}}"github.com/lib/pq"{{end}}

	{{.third}}
)
