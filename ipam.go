package main

import (
	"ipam/routers"
	"ipam/utils/logging"
	"ipam/utils/options"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/pkg/errors"
)

// @title GCloud API
// @description  This is a sample server Petstore server.
// @version 1.0
// @BasePath /api/v1
// @HOST 172.0.0.1:1234
func main() {
	//初始化配置文件
	conf := options.InitConfig()
	//初始化log

	logging.ConfigInit()
	s := &http.Server{
		Addr:           conf.Http.Addr,
		Handler:        routers.InitRouter(),
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   3 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		logging.Error(errors.WithStack(err))
	}
}
