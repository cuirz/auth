// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"

	"config"
	. "models"
)

// Engine represents a xorm engine or session.
type Engine interface {
	Delete(interface{}) (int64, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Find(interface{}, ...interface{}) error
	Get(interface{}) (bool, error)
	Insert(...interface{}) (int64, error)
	InsertOne(interface{}) (int64, error)
	Id(interface{}) *xorm.Session
	Sql(string, ...interface{}) *xorm.Session
	Where(string, ...interface{}) *xorm.Session
}

func sessionRelease(sess *xorm.Session) {
	if !sess.IsCommitedOrRollbacked {
		sess.Rollback()
	}
	sess.Close()
}

var (
	x         *xorm.Engine
	tables    []interface{}
	HasEngine bool

	DbCfg struct {
		Type, Host, Name, User, Passwd, Path, SSLMode string
		MaxIdle, MaxActive                            int
	}

	EnableSQLite3 bool
	EnableTidb bool
)

func GetSql() *xorm.Engine {
	if x == nil {
		x, _ = getEngine()
	}
	return x
}

func init() {
	tables = append(tables, new(Accounts))

	gonicNames := []string{"SSL"}
	for _, name := range gonicNames {
		core.LintGonicMapper[name] = true
	}

	LoadConfigs()

	if err := NewEngine(); err != nil {
		panic(err)
		fmt.Println("数据库连接报错！")
	}
}

func LoadConfigs() {
	c := config.GetConfig()
	sql := config.GetConfig().Mysql
	//var parm []string
	//parm = append(parm, sql.User, ":", sql.Password, "@tcp(", sql.Host, ")/", sql.Database, "?charset=utf8")
	//var err error
	//x, err = xorm.NewEngine("mysql", strings.Join(parm, ""))
	//if err != nil {
	//	panic(err)
	//}
	//x.SetMaxIdleConns(sql.MaxIdle)
	//x.SetMaxOpenConns(sql.MaxActive)
	////为了 表结构为下划线命名之间的转换，比默认 SnakeMapper支持的好
	//x.SetMapper(core.GonicMapper{})

	sec := c.Cfg.Section("database")
	DbCfg.Type = sec.Key("DB_TYPE").String()
	switch DbCfg.Type {
	case "sqlite3":
		config.UseSQLite3 = true
	case "mysql":
		config.UseMySQL = true
	case "postgres":
		config.UsePostgreSQL = true
	case "tidb":
		config.UseTiDB = true
	}
	DbCfg.Host = sql.Host
	DbCfg.Name = sql.Database
	DbCfg.User = sql.User
	if len(DbCfg.Passwd) == 0 {
		DbCfg.Passwd = sql.Password
	}
	DbCfg.SSLMode = sql.SslMode
	DbCfg.Path = sec.Key("PATH").MustString("data/goauth.db")
}

func getEngine() (*xorm.Engine, error) {
	cnnstr := ""
	switch DbCfg.Type {
	case "mysql":
		if DbCfg.Host[0] == '/' {
			// looks like a unix socket
			cnnstr = fmt.Sprintf("%s:%s@unix(%s)/%s?charset=utf8&parseTime=true",
				DbCfg.User, DbCfg.Passwd, DbCfg.Host, DbCfg.Name)
		} else {
			cnnstr = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true",
				DbCfg.User, DbCfg.Passwd, DbCfg.Host, DbCfg.Name)
		}
	case "postgres":
		var host, port = "127.0.0.1", "5432"
		fields := strings.Split(DbCfg.Host, ":")
		if len(fields) > 0 && len(strings.TrimSpace(fields[0])) > 0 {
			host = fields[0]
		}
		if len(fields) > 1 && len(strings.TrimSpace(fields[1])) > 0 {
			port = fields[1]
		}
		cnnstr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			url.QueryEscape(DbCfg.User), url.QueryEscape(DbCfg.Passwd), host, port, DbCfg.Name, DbCfg.SSLMode)
	case "sqlite3":
		if !EnableSQLite3 {
			return nil, fmt.Errorf("Unknown database type: %s", DbCfg.Type)
		}
		if err := os.MkdirAll(path.Dir(DbCfg.Path), os.ModePerm); err != nil {
			return nil, fmt.Errorf("Fail to create directories: %v", err)
		}
		cnnstr = "file:" + DbCfg.Path + "?cache=shared&mode=rwc"
	case "tidb":
		if !EnableTidb {
			return nil, fmt.Errorf("Unknown database type: %s", DbCfg.Type)
		}
		if err := os.MkdirAll(path.Dir(DbCfg.Path), os.ModePerm); err != nil {
			return nil, fmt.Errorf("Fail to create directories: %v", err)
		}
		cnnstr = "goleveldb://" + DbCfg.Path
	default:
		return nil, fmt.Errorf("Unknown database type: %s", DbCfg.Type)
	}
	fmt.Println("mysql:",cnnstr)
	return xorm.NewEngine(DbCfg.Type, cnnstr)
}

func NewTestEngine(x *xorm.Engine) (err error) {
	x, err = getEngine()
	if err != nil {
		return fmt.Errorf("Connect to database: %v", err)
	}

	x.SetMapper(core.GonicMapper{})
	return x.StoreEngine("InnoDB").Sync2(tables...)
}

func SetEngine() (err error) {
	x, err = getEngine()
	if err != nil {
		return fmt.Errorf("Fail to connect to database: %v", err)
	}
	x.SetMaxIdleConns(DbCfg.MaxIdle)
	x.SetMaxOpenConns(DbCfg.MaxActive)

	x.SetMapper(core.GonicMapper{})

	// WARNING: for serv command, MUST remove the output to os.stdout,
	// so use log file to instead print to stdout.
	logPath := path.Join(config.LogRootPath, "xorm.log")
	os.MkdirAll(path.Dir(logPath), os.ModePerm)

	f, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("Fail to create xorm.log: %v", err)
	}
	x.SetLogger(xorm.NewSimpleLogger(f))
	x.ShowSQL(true)
	return nil
}

func NewEngine() (err error) {
	if err = SetEngine(); err != nil {
		return err
	}

	if err = x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
		return fmt.Errorf("sync database struct error: %v\n", err)
	}

	return nil
}

type Statistic struct {
	Counter struct {
				User, Org, PublicKey,
				Repo, Watch, Star, Action, Access,
				Issue, Comment, Oauth, Follow,
				Mirror, Release, LoginSource, Webhook,
				Milestone, Label, HookTask,
				Team, UpdateTask, Attachment int64
			}
}

func GetStatistic() (stats Statistic) {
	//stats.Counter.User = CountUsers()
	//stats.Counter.Org = CountOrganizations()
	//stats.Counter.PublicKey, _ = x.Count(new(PublicKey))
	//stats.Counter.Repo = CountRepositories()
	//stats.Counter.Watch, _ = x.Count(new(Watch))
	//stats.Counter.Star, _ = x.Count(new(Star))
	//stats.Counter.Action, _ = x.Count(new(Action))
	//stats.Counter.Access, _ = x.Count(new(Access))
	//stats.Counter.Issue, _ = x.Count(new(Issue))
	//stats.Counter.Comment, _ = x.Count(new(Comment))
	//stats.Counter.Oauth = 0
	//stats.Counter.Follow, _ = x.Count(new(Follow))
	//stats.Counter.Mirror, _ = x.Count(new(Mirror))
	//stats.Counter.Release, _ = x.Count(new(Release))
	//stats.Counter.LoginSource = CountLoginSources()
	//stats.Counter.Webhook, _ = x.Count(new(Webhook))
	//stats.Counter.Milestone, _ = x.Count(new(Milestone))
	//stats.Counter.Label, _ = x.Count(new(Label))
	//stats.Counter.HookTask, _ = x.Count(new(HookTask))
	//stats.Counter.Team, _ = x.Count(new(Team))
	//stats.Counter.UpdateTask, _ = x.Count(new(UpdateTask))
	//stats.Counter.Attachment, _ = x.Count(new(Attachment))
	return
}

func Ping() error {
	return x.Ping()
}

// DumpDatabase dumps all data from database to file system.
func DumpDatabase(filePath string) error {
	return x.DumpAllToFile(filePath)
}
