package main

import (
	"app/internal/inits"

	"github.com/mszlu521/thunder/config"
	"github.com/mszlu521/thunder/logs"
	"github.com/mszlu521/thunder/server"
)

func main() {
	config.Init()
	conf := config.GetConfig()
	logs.Init(conf.Log)
	s := server.NewServer(conf)
	inits.Init(s, conf)
	s.Start()
}
