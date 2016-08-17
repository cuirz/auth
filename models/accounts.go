package models

import (
	"gopkg.in/macaron.v1"
	"github.com/go-macaron/binding"
	"strings"
)

type (
	Accounts struct {
		Id           int        `form:"id" xorm:"'id' pk autoincr"`
		UserId       string     `form:"userid" xorm:"'userid' varchar(256) index"`
		ReUserId     string     `form:"reuserid" xorm:"'reuserid' varchar(256)"`
		UserName     string     `form:"username" xorm:"'username' varchar(128)"`
		Password     string     `form:"password" xorm:"'password' varchar(50)"`
		Sex          byte       `form:"sex" xorm:"'sex'"`
		PhoneCode    string     `form:"phonecode" xorm:"'phonecode' varchar(30)"`
		Email        string     `form:"email" xorm:"'mail' varchar(50)"`
		ProductId    int        `binding:"Default(2)" form:"product_id" xorm:"'productid'"`
		ChannelId    string     `form:"channelid" xorm:"'channelid' varchar(64)"`
		Belongto     int        `form:"belongto" xorm:"'belongto'"`
		Permit       string     `form:"permit" xorm:"'permit' varchar(50)"`
		Rtime        string     `form:"rtime" xorm:"'rtime' DateTime" `
		Ltime        string     `form:"ltime" xorm:"'ltime' DateTime" `
		RegisterIp   string     `form:"remoteAddress" xorm:"'registerip' varchar(30)"`
		State        int        `form:"state" xorm:"'state'"`
		RegisterFrom int        `form:"registerfrom" xorm:"'registerfrom'"`
		Money        int        `form:"money" xorm:"'money'"`
		Freeze       int64      `form:"freeze" xorm:"'freeze'"`
		DeviceOs     string     `form:"deivceos" xorm:"'deivceos' varchar(50)"`
		Osversion    string     `form:"osversion" xorm:"'osversion' varchar(50)"`
		DeviceSerial string     `form:"deviceserial" xorm:"'deviceserial' varchar(50)"`
		DeviceUDID   string     `form:"deviceudid" xorm:"'deviceudid' varchar(128)"`
		DeviceMAC    string     `form:"devicemac" xorm:"'devicemac' varchar(50)"`
		DeviceUA     string     `form:"deviceua" xorm:"'deviceua' varchar(300)"`
		ScreenWidth  int        `form:"screenWidth" xorm:"'screenwidth'"`
		ScreenHeight string     `form:"screenHeight" xorm:"'screenheight' varchar(255)"`
		CodeVersion  string     `form:"codeversion" xorm:"'codeversion' varchar(50)"`
		PwdState     byte       `form:"pwdstate" xorm:"'pwdstate'"`
		FreezeDay    int64      `form:"freezeday" xorm:"'freezeday'"`
		LoginCount   int        `form:"logincount" xorm:"'logincount'"`
		ChannelSub   string     `form:"channelsub" xorm:"'channelsub' varchar(256)"`
		ActiveCode   string     `form:"active_code" xorm:"'active_code' varchar(64)"`

		LastLogin    string		`xorm:"-"`
	}



)

func (ac *Accounts) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	//额外处理功能
	ac.ActiveCode = strings.ToUpper(ac.ActiveCode)
	ac.RegisterIp = ac.calAddress(ctx)
	if strings.Contains(ac.ActiveCode, "1") {
		errs = append(errs, binding.Error{
			FieldNames:     []string{"message"},
			Classification: "FormValidateError",
			Message:        "ActiveCode numberic error",
		})
	}
	return errs
}

func (ac *Accounts)calAddress(ctx *macaron.Context) string {
	ip := ctx.Query("X-Real-IP")
	if len(ip) < 6 {
		ip = "/0.0.0.0:"
		if ac.RegisterIp != "" {
			ip = ac.RegisterIp
		}
		if start, end := strings.Index(ip, "/"), strings.Index(ip, ":"); start > -1 && end > -1 {
			ip = ip[start + 1:end]
		}
	}
	return ip
}