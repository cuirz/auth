package models

type(
	LoginRecord struct {
		Id      int       `form:"id" xorm:"'accountid' autoincr"`
		Servers string    `form:"servers" xorm:"'servers' Text"`
	}

	SecurityCode struct {
		Id           int	`form:"id"`
		ChannelId    string `form:"channelid"`
		ChannelSub   string	`form:"channelsub"`
		ProductId    int	`form:"productid"`
		CodeVersion  int	`form:"codeversion"`
		UserId       string	`form:"userid"`
		RegisterTime string	`form:"registertime"`
		Belongto     int	`form:"belongto"`
	}
)