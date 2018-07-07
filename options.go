package noodles

import "time"

type Options struct {
	RedisAddr        string        // redis地址,例如:127.0.0.1:6379
	RedisPass        string        // redis认证密码
	RedisDB          string        // redis数据库
	RedisMaxIdle     int           // 最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
	RedisMaxActive   int           // 最大的激活连接数，表示同时最多有N个连接
	RedisIdleTimeout time.Duration // 最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭

	NameSpace   string // 命名空间
	QueueName   string // 队列名,个多队列逗号隔开
	Concurrency int    // 并发执行数量
}
