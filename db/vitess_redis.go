package db

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"fmt"
	"sync"
	"config"
	"golang.org/x/net/context"
)

type ResourceConn struct {
	redis.Conn
}

type ResourcePool struct {
	*pools.ResourcePool
	host string
	db   int
}

var dbRedis2 *ResourcePool

func (r ResourceConn) Close() {
	r.Conn.Close()
}

func GetRedis2() *ResourcePool {
	if dbRedis2 == nil {
		return InitRedis2()
	}
	return dbRedis2
}

func InitRedis2() *ResourcePool {
	conf := config.GetConfig()
	dbNum := 0
	p := pools.NewResourcePool(func() (pools.Resource, error) {
		optTimeout := redis.DialConnectTimeout(time.Duration(conf.Redis.ConnTimeout) * time.Second)
		optSelectDB := redis.DialDatabase(dbNum)
		c, err := redis.Dial("tcp", "192.168.1.201:6379", optSelectDB, optTimeout)
		if err != nil {
			fmt.Println("连接失败")
			return ResourceConn{c}, err
		}
		if _, err := c.Do("AUTH", conf.Redis.Password); err != nil {
			return ResourceConn{c}, err
		}
		return ResourceConn{c}, err
	}, conf.Redis.MaxIdle, conf.Redis.MaxActive, time.Duration(conf.Redis.IdleTimeout) * time.Second)

	return &ResourcePool{ResourcePool:p, host:conf.Redis.Host, db:dbNum}
}

//set 操作
func (p *ResourcePool) SAdd(ctx context.Context, key string, arg interface{}) error {
	r, err := p.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Put(r)
	c := r.(ResourceConn)
	_, err = redis.Int64(c.Do("SADD", key, arg))
	return err
}

func (p *ResourcePool) HGetAll(ctx context.Context,key string) ([]interface{}, error) {
	r, err := p.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Put(r)
	c := r.(ResourceConn)
	return redis.Values(c.Do("HGETALL", key))
	//value, err := redis.Strings(c.Do("HGETALL", key))
	//return value, err
}

func (p *ResourcePool) HSet(ctx context.Context ,key, field, value string) error {
	r, err := p.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Put(r)
	c := r.(ResourceConn)
	_, err = redis.Int64(c.Do("HSET", key, field, value))
	return err
}

func (p *ResourcePool) HMset(ctx context.Context ,key string, value interface{}) error {
	r, err := p.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Put(r)
	c := r.(ResourceConn)
	_, err = c.Do("HMSET", redis.Args{}.Add(key).AddFlat(value)...)
	return err
}

var mutex sync.Mutex

//func main() {
//	e := echo.New()
//	p := pools.NewResourcePool(func() (pools.Resource, error) {
//		c, err := redis.Dial("tcp", "192.168.1.201:6379")
//		c.Do("AUTH", "redis666")
//		return ResourceConn{c}, err
//	}, 5, 30, time.Minute)
//	defer p.Close()
//	index := 1
//	e.GET("/", func(ctx echo.Context) error {
//		r, err := p.Get(ctx)
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer p.Put(r)
//		c := r.(ResourceConn)
//		//n, err := c.Do("INFO")
//		//if err != nil {
//		//	log.Fatal(err)
//		//}
//		mutex.Lock()
//		index ++
//		_, err = c.Do("SET", "color", index)
//		if err != nil {
//			index --
//			log.Fatal(err)
//		}
//		mutex.Unlock()
//		//log.Printf("info=%s", n)
//		//_ = n
//		fmt.Println("ok=", index)
//
//		return ctx.String(http.StatusOK, "Hello, World!")
//	})
//	e.Run(standard.New(":1323"))
//
//
//	//
//	//ctx := context.TODO()
//	//fmt.Println("什么情况")
//
//}

//func (p *RedisPool) Set(key, value string) error {
//	c := p.get()
//	defer c.Close()
//	_, err := redis.String(c.Do("SET", key, value))
//	return err
//}
//
//func (p *RedisPool) Get(key string) (string, error) {
//	c := p.get()
//	defer c.Close()
//	value, err := redis.String(c.Do("GET", key))
//	return value, err
//}