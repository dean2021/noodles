package noodles

import (
	"fmt"
	"strings"
	"bytes"
	"sync"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"os"
	"github.com/apsdehal/go-logger"
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
	conn := RedisClient.Get()
	defer conn.Close()
	s := sync.WaitGroup{}
	queues := strings.Split(n.QueueName, ",")
	for _, queue := range queues {
		i := 0
		for i < n.Concurrency {
			reply, err := conn.Do("LPOP", fmt.Sprintf("%s:queue:%s", n.NameSpace, queue))
			if err != nil {
				return err
			}
			if reply != nil {
				task := Task{}
				decoder := json.NewDecoder(bytes.NewReader(reply.([]byte)))
				if err := decoder.Decode(&task); err != nil {
					return err
				}
				if workFunc, ok := WorkFunc[task.FuncName]; ok {
					n.Run(workFunc, task, &s)
				} else {
					Log.Critical("No registered " + task.FuncName + " function")
				}
			}
			i++
		}
	}
	s.Wait()
	return nil
}

func (n *Noodles) Run(workFunc func([]interface{}, ) error, task Task, s *sync.WaitGroup) {
	s.Add(1)
	go func() {
		var err error
		defer func() {
			if err := recover(); err != nil {
				n.LogError(err.(string))
			}
		}()
		err = workFunc(task.Args)
		s.Done()
		if err != nil {
			n.LogError(err.Error())
		}
	}()
}

// 添加任务到队列
func (n *Noodles) AddTask(task Task) error {
	conn := RedisClient.Get()
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

func (n *Noodles) LogError(error string) error {
	conn := RedisClient.Get()
	defer conn.Close()
	_, err := conn.Do("RPUSH", fmt.Sprintf("%s:errors", n.NameSpace), error)
	if err != nil {
		Log.Critical(err.Error())
		return err
	}
	return nil
}

func New(options Options) *Noodles {
	n := &Noodles{
		NameSpace:   options.NameSpace,
		QueueName:   options.QueueName,
		Concurrency: options.Concurrency,
	}
	return n.Init(options)
}
