package versionlogic

import (
	"context"
	"runtime"

	"github.com/zeromicro/go-zero/core/logx"

	"{{.Module}}/internal/svc"
	"{{.Module}}/internal/types/version"
	ver "{{.Module}}/version"
)

type Version struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVersion(ctx context.Context, svcCtx *svc.ServiceContext) *Version {
	return &Version{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *Version) Version(in *version.VersionRequest) (*version.VersionResponse, error) {
	return &version.VersionResponse{
		Version:   ver.Version,
		GoVersion: runtime.Version(),
		Commit:    ver.Commit,
		Date:      ver.Date,
	}, nil
}
