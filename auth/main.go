package auth

import (
	"net/http"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/echo/engine/standard"
	//"github.com/labstack/echo/engine/fasthttp"
	_ "flag"
	_ "fmt"
	"strconv"
	"fmt"

	"runtime"
	"db"
)

// ServerHeader middleware adds a `Server` header to the response.
func ServerHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, "Echo/2.0")
		return next(c)
	}
}

//var infile *string = flag.String("i", "unsortedooo.dat", "File contains values for sorting")

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
func main() {
	//flag.Parse()
	//
	//if infile != nil {
	//	fmt.Println("infile =", *infile)
	//}

	//初始化配置
	//config.Init()

	e := echo.New()
	// Debug mode
	e.SetDebug(true)

	e.Use(middleware.Logger())
	//e.Use(func(c echo.Context) error {
	//	println(c.Path()) // Prints `/users/:name`
	//	return nil
	//})
	// Server header
	Router(e)

	e.GET("/users/:name", func(c echo.Context) error {
		// By name
		name := c.Param("name")
		return c.String(http.StatusOK, name)
	})
	var index int
	e.GET("/", func(c echo.Context) error {
		pool := db.GetRedis()
		index ++
		err := pool.AddToSet("color2" , strconv.Itoa(index))
		//_, err := pool.GetALLFromSet("color" + strconv.Itoa(index))
		if (err != nil) {
			fmt.Println(err)
			return c.String(http.StatusNoContent, "oooo")
		}

		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/users/:id", func(c echo.Context) error {
		//array := [10]int{1,2,3}
		var i int
		for index := 0; index < 10000; index ++ {
			i ++
			i --
			i *= 10
			i = i + 1000
			var ok = func(x, y int) int {
				return x + y
			}

			i = ok(i, i + 1)
		}
		println("这个打印会影响性能吗 = %d", i)
		return c.String(http.StatusOK, "/users/:id")
	})
	e.GET("/users/:name", func(c echo.Context) error {
		return c.String(http.StatusOK, "/users/new")
	})

	e.GET("/users", func(c echo.Context) error {
		name := c.QueryParam("name")
		return c.String(http.StatusOK, name)
	})

	e.POST("/users", func(c echo.Context) error {
		name := c.FormValue("name")
		return c.String(http.StatusOK, name)
	})

	//g := e.Group("/admin")
	//g.Use(middleware.BasicAuth(func(username, password string) bool {
	//	if username == "joe" && password == "secret" {
	//		return true
	//	}
	//	return false
	//}))
	//e.Run(fasthttp.New(":1323"))
	e.Run(standard.New(":12006"))
}