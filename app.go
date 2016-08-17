package main

import (
	"github.com/urfave/cli"
	"runtime"
	"appcmd"
	"config"
	"fmt"
	"flag"
)

const APP_VER = "0.0.1.0001"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config.AppVer = APP_VER
}

func main() {
	app := cli.NewApp()
	app.Name = "auth"
	app.Usage = "SDK Account Service"
	app.Version = APP_VER
	app.Commands = []cli.Command{
		appcmd.CmdWeb,
		//cmd.CmdServ,
		//cmd.CmdUpdate,
		//cmd.CmdDump,
		//cmd.CmdCert,
	}
	flag.Parse()
	flag.CommandLine.Parse([]string{"","web","--port","12006"})
	fmt.Println("arg:",flag.Args())

	//var args []cli.Flag
	//for i, name := range flag.Args() {
	//	args[i] = stringFlag("port", "12006", "Temporary port number to prevent conflict")
	//
	//}


	app.Flags = append(app.Flags, []cli.Flag{}...)
	//app.Flags = append(app.Flags, []cli.Flag{
	//	cli.StringFlag{
	//		Name:  "port",
	//		Value: "12006",
	//		Usage: "Temporary port number to prevent conflict",
	//	},
	//}...)



	//app.Run(app.Flags)
	app.Run(flag.Args())
}
