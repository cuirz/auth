package db

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	_ "net/http"
	_ "strconv"
	"strconv"
)

func main() {
	e := echo.New()
	reConf := redisConf{
		MaxIdle: 1,
		MaxActive: 1,
		IdleTimeout: 240,
		ConnTimeout: 5,
	}
	var conf Configure
	var mylog MyLog
	conf.Redis = reConf

	pool := newPool(&conf, "192.168.1.201:6379", "redis666", 0, &mylog)
	var index int = 1
	e.GET("/", func(c echo.Context) error {
		index ++
		pool.AddToSet("color" + strconv.Itoa(index), strconv.Itoa(index))
		_, err := pool.GetALLFromSet("color" + strconv.Itoa(index))
		if (err != nil) {
			return c.String(http.StatusOK, "Hello, World!")
		}
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Run(standard.New(":1323"))
}