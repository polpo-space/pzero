package config

import (
    "github.com/zeromicro/go-zero/core/logx"
    "github.com/zeromicro/go-zero/zrpc"
    {{ if has "model" .Features }}"github.com/zeromicro/go-zero/core/stores/sqlx"{{ end }}
    {{ if has "redis" .Features }}"github.com/zeromicro/go-zero/core/stores/redis"{{ end }}
    {{ if has "cache" .Features }}"github.com/zeromicro/go-zero/core/stores/cache"
    "github.com/zeromicro/go-zero/core/stores/redis"{{ end }}
)

type Config struct {
	Zrpc    ZrpcConf
	Log     LogConf
	{{ if has "model" .Features }}Sqlx SqlxConf{{ end }}
    {{ if has "redis" .Features }}Redis RedisConf{{ end }}
    {{ if has "cache" .Features }}Cache CacheConf{{ end }}
	{{ if has "job" .Features }}Job JobConf{{ end }}
}

type ZrpcConf struct {
	zrpc.RpcServerConf
}

type LogConf struct {
	logx.LogConf
}

{{ if has "model" .Features }}type SqlxConf struct {
	sqlx.SqlConf
}{{ end }}
{{ if has "redis" .Features }}type RedisConf struct {
    redis.RedisConf
}{{ end }}
{{ if has "cache" .Features }}type CacheConf struct {
	Expiry        int64 `json:",default=300000"`  // 默认 300s
	NotFoundExpiry int64 `json:",default=60000"` // 默认 60s
	Redis         cache.CacheConf
}{{ end }}
{{ if has "job" .Features }}type JobConf struct {
	Enable bool `json:",default=false"`
}{{ end }}
