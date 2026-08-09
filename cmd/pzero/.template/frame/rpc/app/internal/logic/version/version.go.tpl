package versionlogic

import (
	"context"
	"runtime"

	"github.com/zeromicro/go-zero/core/logx"

	"{{.Module}}/internal/buildinfo"
	"{{.Module}}/internal/svc"
	"{{.Module}}/internal/types/version"
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
		Version:   buildinfo.Version,
		GoVersion: runtime.Version(),
		Commit:    buildinfo.Commit,
		Date:      buildinfo.Date,
	}, nil
}
