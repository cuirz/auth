package repository

import (
	"config"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"strings"
	"time"
	"github.com/go-xorm/core"
)

type User struct {
	Id         int64
	Userid     string    `xorm:"varchar(256) index"`
	Reuserid   string    `xorm:"varchar(200)"`
	Sex        int
	Age        int
	Rtime      time.Time `xorm:"deleted"`
	Inserttime time.Time `xorm:"created"`
	Updatetime time.Time `xorm:"updated"`
	Version    int       `xorm:"version"`
}

//func (u *User)TableName() string {
//	return "user"
//}

var engine *xorm.Engine

func init() {
	c := config.GetConfig().Mysql
	var parm []string
	parm = append(parm, c.User, ":", c.Password, "@tcp(", c.Host, ")/", c.Database, "?charset=utf8")
	var err error
	engine, err = xorm.NewEngine("mysql", strings.Join(parm, ""))
	if err != nil {
		panic(err)
	}
	engine.SetMaxIdleConns(c.MaxIdle)
	engine.SetMaxOpenConns(c.MaxActive)
	//为了 表结构为下划线命名之间的转换，比默认 SnakeMapper支持的好
	engine.SetMapper(core.GonicMapper{})
	//engine.DropTables(&User{})
	if err := engine.Ping(); err != nil {
		fmt.Println(err)
	}

	if err := engine.Sync2(new(User)); err != nil {
		panic(err)
	}

	affected, err := engine.Insert(&User{5, "8912", "8912", 2, 30, time.Now(), time.Time{}, time.Time{},0})
	fmt.Println(affected)
	var user User
	//engine.Id(4).Delete(&user)
	ok,err := engine.Id(5).Get(&user)
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println(ok)
}

func main() {
	//var err error
	//engine, err = xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
}
