package db

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"config"
	"fmt"

)

type RedisPool struct {
	c    chan struct{}
	*redis.Pool
	host string
	db   int
}


type waitConn struct {
	c chan struct{}
	redis.Conn
}

func (conn *waitConn) Close() error {
	//fmt.Println("close :",len(conn.c))
	defer func() {
		<-conn.c
	}()
	return conn.Conn.Close()
}


func (pool *RedisPool)get() *waitConn {
	//fmt.Println("get :",len(pool.c))
	pool.c <- struct{}{}
	return &waitConn{c: pool.c, Conn: pool.Pool.Get()}
}
//============================================
var dbRedis *RedisPool

//func (p *RedisPool)get() redis.Conn {
//	return p.p.Get()
//}

//func (p *RedisPool) Close() {
//	p.p.Close()
//	//p.mylog.log.Infof("[redis_pool] close redis pool(%s:%d) success.", p.host, p.db)
//}

func GetRedis() *RedisPool {
	if dbRedis == nil {
		return InitRedis()
	}
	return dbRedis
}

func InitRedis() *RedisPool {
	var mylog config.MyLog
	c := config.GetConfig()
	dbRedis = newPool(c, 0, make(chan struct{},c.Redis.MaxActive), &mylog)
	return dbRedis
}

func newPool(conf *config.Configure, dbnum int, c chan struct{},mylog *config.MyLog) *RedisPool {
	pool := &redis.Pool{
		MaxIdle:     conf.Redis.MaxIdle,
		MaxActive:   conf.Redis.MaxActive,
		IdleTimeout: time.Duration(conf.Redis.IdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			opt_timeout := redis.DialConnectTimeout(time.Duration(conf.Redis.ConnTimeout) * time.Second)
			opt_selectdb := redis.DialDatabase(dbnum)
			c, err := redis.Dial("tcp", conf.Redis.Host, opt_timeout, opt_selectdb)
			if err != nil {
				fmt.Println("连接失败")
				return nil, err
			}
			if _, err := c.Do("AUTH", conf.Redis.Password); err != nil {
				return nil, err
			}

			return c, err
		},
		//TestOnBorrow: func(c redis.Conn, t time.Time) error {
		//	_, err := c.Do("PING")
		//	return err
		//},
	}
	p := &RedisPool{Pool: pool, db: dbnum, host: conf.Redis.Host,c: c}
	return p
}





//set 操作
func (p *RedisPool) SAdd(key string, arg interface{}) error {
	c := p.get()
	defer c.Close()
	_, err := redis.Int64(c.Do("SADD", key, arg))
	return err
}

func (p *RedisPool) SRem(key string, arg interface{}) error {
	c := p.get()
	defer c.Close()
	_, err := redis.Int64(c.Do("SREM", key, arg))
	return err
}

func (p *RedisPool) SMembers(key string) ([]string, error) {
	c := p.get()
	defer c.Close()
	array, err := redis.Strings(c.Do("SMEMBERS", key))
	return array, err
}

func (p *RedisPool) SDiff(set1, set2 string) ([]string, error) {
	c := p.get()
	defer c.Close()
	array, err := redis.Strings(c.Do("SDIFF", set1, set2))
	return array, err
}

//hash 操作
func (p *RedisPool) HSet(key, field, value string) error {
	c := p.get()
	defer c.Close()
	_, err := redis.Int64(c.Do("HSET", key, field, value))
	return err
}

func (p *RedisPool) HMset(key string, value interface{}) error {
	c := p.get()
	defer c.Close()
	_, err := c.Do("HMSET", redis.Args{}.Add(key).AddFlat(value)...)
	return err
}

func (p *RedisPool) HDel(key, field string) error {
	c := p.get()
	defer c.Close()
	_, err := redis.Int64(c.Do("HDEL", key, field))
	return err
}

func (p *RedisPool) HGet(key, field string) (string, error) {
	c := p.get()
	defer c.Close()
	value, err := redis.String(c.Do("HGET", key, field))
	return value, err
}

func (p *RedisPool) HGetAll(key string) ([]interface{}, error) {
	c := p.get()
	defer c.Close()
	return redis.Values(c.Do("HGETALL", key))
	//value, err := redis.Strings(c.Do("HGETALL", key))
	//return value, err
}

//string 操作
func (p *RedisPool) Set(key, value string) error {
	c := p.get()
	defer c.Close()
	_, err := redis.String(c.Do("SET", key, value))
	return err
}

func (p *RedisPool) Get(key string) (string, error) {
	c := p.get()
	defer c.Close()
	value, err := redis.String(c.Do("GET", key))
	return value, err
}

func (p *RedisPool) MGet(keys []interface{}) ([]string, error) {
	c := p.get()
	defer c.Close()
	value, err := redis.Strings(c.Do("MGET", keys...))
	return value, err
}

func (p *RedisPool) SetPExpire(key string, t string) error {
	c := p.get()
	defer c.Close()
	_, err := redis.Int(c.Do("PEXPIRE", key, t))
	return err
}

//key 操作
//1:exist  0:not exist
func (p *RedisPool) Exists(key string) (int, error) {
	c := p.get()
	defer c.Close()
	value, err := redis.Int(c.Do("EXISTS", key))
	return value, err
}

func (p *RedisPool) DelKey(key string) (int, error) {
	c := p.get()
	defer c.Close()
	value, err := redis.Int(c.Do("DEL", key))
	return value, err
}

func (p *RedisPool) FindKeys(key string) ([]string, error) {
	c := p.get()
	defer c.Close()
	//keys, err := redis.Strings(c.Do("KEYS", key))
	var keys []string
	iter := 0
	for {
		if arr, err := redis.MultiBulk(c.Do("SCAN", iter)); err != nil {
			return keys, err
		} else {
			iter, _ = redis.Int(arr[0], nil)
			tmp_keys, _ := redis.Strings(arr[1], nil)
			keys = append(keys, tmp_keys...)
		}
		if iter == 0 {
			break
		}
	}
	return keys, nil
}
