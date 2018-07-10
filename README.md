# Noodles


Noodles(面条)是一款超轻量级分布式任务调度类库(太轻量级了,谈不上框架),类似于python的celery,大量参考benmanns的goworker.

Noodles实现代码不到200行,能够实现goworker的核心功能,半小时读懂Noodles代码,自行定制功能。

## 安装


```sh
go get github.com/dean2021/noodles
```

引入noodles包

```go
import "github.com/dean2021/noodles"
```

## 快速入门

添加任务

```go

package main

import (
	"github.com/dean2021/noodles"
	"time"
	"fmt"
)

func main() {

	noodle := noodles.New(noodles.Options{
		RedisAddr:        "127.0.0.1:6379",  //redis链接地址
		RedisPass:        "",                //redis认证密码
		RedisDB:          "0",               //redis数据库
		RedisMaxActive:   10,                // 最大的激活连接数，表示同时最多有N个连接
		RedisMaxIdle:     10,                //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		RedisIdleTimeout: 180 * time.Second, // 最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭

		NameSpace: "noodles", // 命名空间
	})

	i := 0
	for i < 100 {
		noodle.AddTask(noodles.Task{
			QueueName: "myQueue",
			FuncName:  "myFunc",
			Args: []interface{}{
				"parameter 1",
				"parameter 2",
				"parameter 3",
				"parameter 4",
				"parameter 5",
				"parameter 6",
			}},
		)
		i++
	}
	fmt.Println("完成100个任务添加.")
}

```

等待任务,执行任务

```go

package main

import (
	"github.com/dean2021/noodles"
	"fmt"
	"time"
)

func main() {

	noodle := noodles.New(noodles.Options{
		RedisAddr:        "127.0.0.1:6379",  //redis链接地址
		RedisPass:        "",                //redis认证密码
		RedisDB:          "0",               //redis数据库
		RedisMaxActive:   10,                // 最大的激活连接数，表示同时最多有N个连接
		RedisMaxIdle:     10,                //最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态
		RedisIdleTimeout: 180 * time.Second, // 最大的空闲连接等待时间，超过此时间后，空闲连接将被关闭

		NameSpace:   "noodles", // 命名空间
		QueueName:   "myQueue", // 队列名,个多队列逗号隔开
		Concurrency: 2,         // 并发执行数量
	})

	// 注册函数
	noodle.Register("myFunc", myFunc)

	// 监听新任务并调度执行
    err := noodle.Worker()
    if err != nil {
        fmt.Println("Error:", err)
    }
}

func myFunc(args []interface{}) error {
	fmt.Printf("call func : %v\n", args)
	time.Sleep(time.Second)
	return nil
}

```
### noodles.Options 介绍
```go

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

```

