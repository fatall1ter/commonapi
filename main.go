package main

//go:generate sqlboiler --wipe --add-global-variants --no-context mssql

import (
	"git.countmax.ru/countmax/commonapi/infra"

	"github.com/sethvargo/go-signalcontext"
)

var (
	version = "1.2.23"
	build   = "-1"
	githash = "hash"
)

// @title CountMax Common API
// @version 2.0
// @contact.name CountMax
// @contact.url https://git.countmax.ru/countmax/commonapi
// @contact.email 1020@watcom.ru
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	ctx, cancel := signalcontext.OnInterrupt()
	defer cancel()
	serv := infra.NewServer(ctx, version, build, githash)
	serv.Run()
	<-ctx.Done()
	serv.Stop()
}
