package db

import (
	"github.com/garyburd/redigo/redis"
	"fmt"
	"time"
	_ "flag"
	_ "reflect"
	_ "strconv"
	"flag"
)

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

var (
	pool *redis.Pool
	redisServer  = flag.String("redisServer", "192.168.1.201:6379", "")
	redisPassword  = flag.String("redisPassword", "redis666", "")
)

func Open() (*redis.Pool){
	pool = newPool(*redisServer, *redisPassword)
	return pool
}

func main() {
	pool = newPool(*redisServer, *redisPassword)
	c := pool.Get()
	v, err := c.Do("SET", "name", "red")
	if err != nil {
		println(err)
	}
	fmt.Println(v)
	v, err = c.Do("GET", "name")
	if err != nil {
		fmt.Println(err)
	}
	//reflect.TypeOf(v)
	//if v2, err := int(v); err == nil {
	//	fmt.Println(v2)
	//
	//}
	switch value := v.(type) {
	case []byte:
		fmt.Println(string(value))
	}
	//fmt.Println(string(v))

}

//Protocol error from client: id=66816 addr=192.168.1.30:53789 fd=157 name= age=0 idle=0 flags=N db=0 sub=0 psub=0 multi=-1 qbuf=42 qbuf-free=32726 obl=44 oll=0 omem=0 events=r cmd=auth