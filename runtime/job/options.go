package job

import "github.com/eddieowens/opts"

type Options struct {
	// Namespace 会作为 job 名的前缀，通常传服务名。
	// 分布式锁以 job 名作为 key，没有前缀时不同服务里的同名 job 会互相抢锁。
	Namespace string
}

// WithNamespace 设置 job 名前缀，最终 job 名为 "<namespace>.<name>"。
func WithNamespace(namespace string) opts.Opt[Options] {
	return func(o *Options) {
		o.Namespace = namespace
	}
}

func (o Options) DefaultOptions() Options {
	return Options{}
}
