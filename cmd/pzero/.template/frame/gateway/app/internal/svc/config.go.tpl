package svc

import "{{ .Module }}/internal/config"

func (svcCtx *ServiceContext) GetConfig() config.Config {
	return svcCtx.Config
}

func (svcCtx *ServiceContext) MustGetConfig() config.Config {
	return svcCtx.Config
}
