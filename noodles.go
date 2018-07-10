package noodles

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/apsdehal/go-logger"
	"github.com/garyburd/redigo/redis"
)

var (
	RedisClient *redis.Pool
	Log         *logger.Logger
	WorkFunc    = make(map[string]func([]interface{}) error)
)

func init() {
	var err error
	Log, err = logger.New("Noodles", 1, os.Stdout)
	if err != nil {
		panic(err)
	}
}

type Noodles struct {
	NameSpace   string
	QueueName   string
	Concurrency int
}

func (n *Noodles) Init(options Options) *Noodles {
	RedisClient = NewRedisPool(options)
	return n
}

// 读取任务并调用注册函数
func (n *Noodles) Worker() error {

	var (
		wg    sync.WaitGroup
		tasks = make(chan Task)
	)

	// 读取任务写入chan
	go func() {
		queues := strings.Split(n.QueueName, ",")
		for {
			for _, queue := range queues {
				conn := RedisClient.Get()
				if conn.Err() != nil {

					// redis可能存在随时中断问题,这里延迟3秒重连,防止这种情况发生
					Log.Critical(conn.Err().Error())
					Log.Info("Retry after 3 seconds")
					conn.Close()
					time.Sleep(time.Second * 3)
					continue
				}

				reply, err := conn.Do("LPOP", fmt.Sprintf("%s:queue:%s", n.NameSpace, queue))
				if err != nil {
					Log.Critical(err.Error())
					continue
				}
				if reply != nil {
					task := Task{}
					decoder := json.NewDecoder(bytes.NewReader(reply.([]byte)))
					if err := decoder.Decode(&task); err != nil {
						Log.Critical(err.Error())
					}
					tasks <- task
				}

				//else{} // 可用于实现读取任务为空终止,这里就没有意义要实现了,因为.....目前想不到有这种需求

				// 并非close,而是放回pool
				conn.Close()
			}
		}
	}()

	// 运行任务并控制并发数
	for i := 0; i < n.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 消费任务
			for task := range tasks {

				if workFunc, ok := WorkFunc[task.FuncName]; ok {

					// call function.
					(func(workFunc func([]interface{}) error, task Task) {

						var err error
						defer func() {
							if err := recover(); err != nil {
								n.LogError(err.(string))
							}
						}()

						err = workFunc(task.Args)
						if err != nil {
							n.LogError(err.Error())
						}

					})(workFunc, task)

				} else {
					err := "No registered " + task.FuncName + " function"
					Log.Critical(err)
					n.LogError(err)
				}
			}
		}()
	}

	wg.Wait()

	return nil
}

// 添加任务到队列
func (n *Noodles) AddTask(task Task) error {
	conn := RedisClient.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	taskJson, err := json.Marshal(task)
	if err != nil {
		Log.Critical(err.Error())
		return err
	}
	_, err = conn.Do("RPUSH", fmt.Sprintf("%s:queue:%s", n.NameSpace, task.QueueName), taskJson)
	if err != nil {
		Log.Critical(err.Error())
		return err
	}
	_, err = conn.Do("SADD", fmt.Sprintf("%s:queues", n.NameSpace), task.QueueName)
	if err != nil {
		Log.Critical(err.Error())
		return err
	}
	return nil
}

// 注册函数
func (n *Noodles) Register(funcName string, function func([]interface{}) error) {
	WorkFunc[funcName] = function
}

func (n *Noodles) LogError(error string) {
	for {
		conn := RedisClient.Get()
		if conn.Err() != nil {
			Log.Critical(conn.Err().Error())
			Log.Info("Retry after 3 seconds")
			conn.Close()
			time.Sleep(time.Second * 3)
			continue
		}

		nowTime := time.Now().UTC().Format(time.UnixDate)
		_, err := conn.Do("RPUSH", fmt.Sprintf("%s:errors", n.NameSpace), fmt.Sprintf("[%s] %s", nowTime, error))
		if err != nil {
			Log.Critical(err.Error())
			continue
		}
		conn.Close()
		break
	}
}

func New(options Options) *Noodles {
	n := &Noodles{
		NameSpace:   options.NameSpace,
		QueueName:   options.QueueName,
		Concurrency: options.Concurrency,
	}
	return n.Init(options)
}
