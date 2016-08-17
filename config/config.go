package config

import (
	"gopkg.in/ini.v1"
	"sync"
	"os"
	"runtime"
	"os/exec"
	"path/filepath"
	"fmt"
)

type(
	Configure struct {
		Redis *redisConf
		Mysql *mysqlConf
		Cfg   *ini.File
	}

	redisConf struct {
		Host        string
		Password    string
		MaxIdle     int
		MaxActive   int
		IdleTimeout int
		ConnTimeout int
	}

	mysqlConf struct {
		Host      string
		User      string
		Password  string
		Database  string
		MaxIdle   int
		MaxActive int
		SslMode   string
	}


)

const (
	CONFIG_FILE string = "config/app.ini"

	DEV = "development"
	PROD = "production"
	TEST = "test"

)

var (
	cfg          *ini.File
	IsWindows bool

	Env = DEV
	envLock sync.Mutex
	Root string     // Path of work directory.

	// App settings
	AppId int
	AppVer string
	AppName string
	AppUrl string
	AppPath string
	AppDataPath string

	LogRootPath string

	HttpPort string

	UseSQLite3 bool
	UseMySQL bool
	UsePostgreSQL bool
	UseTiDB bool
)

func NewRedisConf() *redisConf {
	sec := cfg.Section("redis")
	return &redisConf{
		MaxIdle: sec.Key("MAXIDLE").MustInt(3),
		MaxActive: sec.Key("MAXACTIVE").MustInt(30),
		IdleTimeout: sec.Key("IDLETIMEOUT").MustInt(60),
		ConnTimeout: sec.Key("CONNTIMEOUT").MustInt(5),
		Host: sec.Key("HOST").MustString("192.168.1.201:6379"),
		Password: sec.Key("PASSWD").MustString("redis666"),
	}
}

func NewMysqlConf() *mysqlConf {
	sec := cfg.Section("database")
	return &mysqlConf{
		Host: sec.Key("HOST").MustString("192.168.1.201:3306"),
		User: sec.Key("USER").MustString("root"),
		Password: sec.Key("PASSWD").MustString("666666"),
		Database: sec.Key("NAME").MustString("db_account_doupo"),
		MaxIdle: sec.Key("MAXIDLE").MustInt(5),
		MaxActive: sec.Key("MAXACTIVE").MustInt(30),
		SslMode: sec.Key("SSL_MODE").String(),
	}
}

type MyLog interface {

}

var config *Configure

func init() {
	setENV(os.Getenv("PROJECT_ENV"))

	IsWindows = runtime.GOOS == "windows"
	var err error
	//if AppPath, err = execPath(); err != nil {
	//	log.Fatal(4, "fail to get app path: %v\n", err)
	//}
	//
	//// Note: we don't use path.Dir here because it does not handle case
	////	which path starts with two "/" in Windows: "//psf/Home/..."
	//AppPath = strings.Replace(AppPath, "\\", "/", -1)

	Root, err = os.Getwd()
	fmt.Println(Root)
	if err != nil {
		panic("error getting work directory: " + err.Error())
	}
	AppPath = Root

	Init()
}

// execPath returns the executable path.
func execPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	return filepath.Abs(file)
}

func Init() *Configure {

	config = new(Configure)
	config.Cfg, _ = setConfig(CONFIG_FILE)
	config.Redis = NewRedisConf()
	config.Mysql = NewMysqlConf()

	return config
}

func GetConfig() *Configure {
	if config == nil {
		return Init()
	}
	return config
}

func setENV(e string) {
	envLock.Lock()
	defer envLock.Unlock()

	if len(e) > 0 {
		Env = e
	}
}

func safeEnv() string {
	envLock.Lock()
	defer envLock.Unlock()

	return Env
}


// SetConfig sets data sources for configuration.
func setConfig(source interface{}, others ...interface{}) (_ *ini.File, err error) {
	cfg, err = ini.Load(source, others...)
	cfg = conf()
	cfg.NameMapper = ini.AllCapsUnderscore
	return cfg, err
}

// Config returns configuration convention object.
// It returns an empty object if there is no one available.
func conf() *ini.File {
	if cfg == nil {
		return ini.Empty()
	}
	return cfg
}


