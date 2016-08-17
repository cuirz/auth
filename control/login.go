package control

import (
	"github.com/labstack/echo"
	"net/http"
	"encoding/json"
	"time"
	"fmt"
	"gopkg.in/macaron.v1"
	. "models"
	"strings"
	"github.com/garyburd/redigo/redis"
	"db"
	"utils"
	"config"
	"sync"
)

var (
	STATUS = map[string]int{
		"OK": 0,
		"FAIL": 1,
		"FAIL_ACTIVE": 2,
		"NEED_ACTIVE": 3,
		"DEVICE_FROZEN": 4,
	}

	MSG = map[string]string{
		"OK": "OK",
		"NO": "未知",
		"DEVICE_FROZEN": "注册失败,设备被冻结",
		"FAIL_RETRY": "注册失败,请稍后再试",
	}
)

type (

	_ time.Time

	Common struct {
		uid string
	}

	//tokenDict map[string]interface{}

)

var tokenDict = make(map[string]interface{})
var mutex sync.Mutex
var Count int

func AddCount(){
	mutex.Lock()
	Count ++
	fmt.Println("完成请求数：",Count)
	mutex.Unlock()
}

func cache(pool *db.RedisPool, id string) (*Accounts, error) {
	values, err := pool.HGetAll(fmt.Sprintf("%s", id))
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, err
	}
	var person Accounts
	err = redis.ScanStruct(values, &person)
	return &person, err
}

//func cacheLogin(pool *db.ResourcePool, id string) (*Accounts, error) {
//	values, err := pool.HGetAll(ctx,fmt.Sprintf("%s", id))
//	if err != nil {
//		return nil, err
//	}
//	if len(values) == 0 {
//		return nil, err
//	}
//	var person Accounts
//	err = redis.ScanStruct(values, &person)
//	return &person, err
//}

//func cacheLogin(pool *db.RedisPool, id string) string {
//	values, err := pool.Get(fmt.Sprintf("%s", id))
//	if err != nil {
//		return ""
//	}
//	return values
//}

func checkToken(token interface{}) (bool, string) {
	var value []byte
	var ok bool
	if value, ok = token.([]byte); !ok {
		return false, "0"
	}
	fmt.Println(value)
	msg := &Common{}

	if err := json.Unmarshal(value, msg); err == nil {
		return true, msg.uid
	}
	return false, "0"
}

//func (ac Accounts) Error(ctx *macaron.Context, errs binding.Errors) {
//	// 自定义错误处理过程
//	ctx.JSON(200,errs)
//}


func LoginGet(ctx *macaron.Context, form Accounts) {
	result := make(map[string]interface{})

	userId := form.UserId
	if strings.EqualFold("anysdk", form.ChannelId) {
		if ok, id := checkToken(tokenDict[ctx.Query("token")]); ok {
			userId = id
			sendClient(ctx, result)
			return
		}
	}
	//fmt.Println(userId)

	if strings.EqualFold(strings.TrimSpace(userId), "") {
		sendClient(ctx, result)
		return
	}


	key := strings.Join([]string{form.ChannelId, form.ChannelSub, userId}, ":")


	//缓存中取数据
	redPool := db.GetRedis()
	x := db.GetSql()
	//redConn := db.Get()

	//if true{
	//	redPool.Set("color",strconv.Itoa(Count))
	//	return
	//}

	var user *Accounts
	var err error
	if user, err = cache(redPool, key); err != nil {
		fmt.Println("error:",err)
		//err = errors.New("index range out")
		ctx.PlainText(200, []byte(err.Error()))
		return
	} else if user != nil {
		//更新缓存、更新数据库
		if ok := updateCacheAndSql(ctx,user, key); !ok {
			sendClient(ctx, result)
			return
		}

	} else {
		fmt.Println("惊人的讨厌")
		//但缓存中没有数据时 从数据库中取数据   并更新到缓存中
		user = &Accounts{}

		opt := ""
		if form.ChannelSub == "" {
			opt = "channelsub IS NULL"
		} else {
			opt = fmt.Sprintf("channelsub = '%s'", form.ChannelSub)
		}
		if ok, _ := x.Where("userid = ?", userId).And("channelid = ?", form.ChannelId).And(opt).Get(user); ok {
			//数据库中有数据
			//更新缓存、更新数据库
			if ok := updateCacheAndSql(ctx,user, key); !ok {
				sendClient(ctx, result)
				return
			}
		} else {
			//创建用户
			if ok := createUser(&form); !ok {
				sendClient(ctx, result)
				return
			} else {
				//更新到缓存
				user = &form
				if err := redPool.HMset(key, user); err != nil {
					sendClient(ctx, result)
					return
				}
			}
		}
	}

	//记录登录历史
	//var history *LoginRecord
	//if user.LastLogin == ""{
	//	history = &LoginRecord{}
	//	if ok, _ := x.Id(user.Id).Get(history); ok {
	//		user.LastLogin = history.Servers
	//		redPool.HSet(key,"LastLogin",user.LastLogin)
	//	}
	//}

	//生成验证码
	securityCode := fmt.Sprintf("SC#%d#%d#%s", form.ProductId, config.AppId, getSecurityCode(user.Id))
	//fmt.Println("验证码：",securityCode)
	sKey := fmt.Sprintf("security:%s",securityCode)
	sCode := &SecurityCode{}

	utils.Copy(user,sCode)
	redPool.HMset(sKey, sCode)


	//返回给客户端消息
	result["status"] = STATUS["OK"]
	result["message"] = MSG["OK"]


	result["accountId"] = user.Id
	result["userid"] = user.UserId
	result["username"] = user.UserName
	result["sex"] = user.Sex
	result["accountFreeze"] = user.Freeze
	result["loginCount"] = user.LoginCount
	result["pwdStatus"] = user.PwdState
	result["securityCode"] = securityCode
	result["lastEnter"] = user.LastLogin
	suffix := ""
	if strings.EqualFold(user.DeviceOs,"ios"){
		suffix = "_ios"
	}
	result["notifyUri"] = config.NOTIFY_URI[fmt.Sprintf("notifyUri%s%s",user.ChannelId,suffix)]

	buf, _ := json.Marshal(&result)
	ctx.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	ctx.PlainText(200, buf)
	AddCount()
}

func updateCacheAndSql(ctx *macaron.Context,user *Accounts, key string) bool {

	//从缓存里取到数据
	user.LoginCount += 1

	//更新缓存
	redPool := db.GetRedis()
	if err := redPool.HMset(key, user); err != nil {
		fmt.Println("error:",err)
		return false
	} else {
		//更新数据库
		x := db.GetSql()
		//fmt.Println("id:",user.Id)
		if _, err := x.Id(user.Id).Cols("logincount").Update(user); err != nil {
			//报错时清除缓存并纠错
			fmt.Println("error,mysql:",err)
			return false
		}
	}
	return true
}

func sendClient(ctx *macaron.Context, result map[string]interface{}) {
	result["status"] = STATUS["FAIL"]
	result["message"] = MSG["FAIL_RETRY"]
	buf, _ := json.Marshal(&result)
	ctx.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	ctx.PlainText(200, buf)
}

func createUser(form *Accounts) bool {
	form.Password = utils.GetSHA1("") //默认没有空密码
	form.Permit = fmt.Sprintf("%s@%s", form.UserId, form.Permit)
	form.Rtime = time.Now().Format("2006-01-02 15:04:05")
	form.Ltime = form.Rtime
	form.LoginCount = 1

	//更新数据库
	x := db.GetSql()

	if _, err := x.InsertOne(form); err != nil {
		//报错时清除缓存并纠错
		fmt.Println(err)
		return false
	}
	return true
}

/**
 * 当前账号服计算的签名数据非常鸡肋，仅为了账号缓存而用，无防作弊功能
 * */
func getSecurityCode(id int)string {
	return fmt.Sprintf("%x%x", (time.Now().Unix() & 0x7fff), (0x111111 * id));
}

func Login(c echo.Context) error {
	user := new(Accounts)

	result := make(map[string]interface{})
	if len(c.Request().Header().Get(echo.HeaderContentType)) == 0 {
		c.Request().Header().Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	}
	//c.Request().Header().Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	if err := c.Bind(user); err != nil {
		fmt.Println(err)
		result["status"] = STATUS["FAIL"]
		result["message"] = MSG["FAIL_RETRY"]
		js, err := json.Marshal(result)
		if err != nil {
			return c.String(http.StatusOK, fmt.Sprint(err))
		}
		//c.JSON(http.StatusOK, )
		return c.String(http.StatusOK, string(js))
	}

	result["status"] = STATUS["FAIL"]
	result["message"] = MSG["FAIL_RETRY"]
	js, err := json.Marshal(result)
	if err != nil {
		return c.String(http.StatusOK, fmt.Sprint(err))
	}
	//c.JSON(http.StatusOK, )

	return c.String(http.StatusOK, string(js))

}
//type Person struct {
//	Name string `redis:"name"`
//	Age  int    `redis:"age"`
//}

