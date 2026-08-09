package version

import (
	"context"
	"net/http"
	"runtime"

	"github.com/zeromicro/go-zero/core/logx"

	"{{.Module}}/internal/buildinfo"
	"{{.Module}}/internal/svc"
	types "{{.Module}}/internal/types/version"
)

type Version struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewVersion(ctx context.Context, svcCtx *svc.ServiceContext, r *http.Request) *Version {
	return &Version{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *Version) Version(req *types.VersionRequest) (resp *types.VersionResponse, err error) {
	return &types.VersionResponse{
		Version:   buildinfo.Version,
		GoVersion: runtime.Version(),
		Commit:    buildinfo.Commit,
		Date:      buildinfo.Date,
	}, nil
}
