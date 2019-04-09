package main

import (
	"./config"
	"flag"
	"github.com/kardianos/service"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)


type services struct {
	log         service.Logger
	cfg         *service.Config
}

func (winSrv *services) Start(s service.Service) error {
	if winSrv.log != nil {
		winSrv.log.Info("Start running")
	}
	go winSrv.run()
	return nil

}

func (winSrv *services) run() error {
	var err error

	// 启动http模块
	go func() {
		winSrv.log.Info("starting proxy server")
		// service connections
		err = RunProxy()
	}()
	return err
}

func (winSrv *services) Stop(s service.Service) error {
	if winSrv.log != nil {
		winSrv.log.Info("stop server")
	}
	sigs <- syscall.SIGTERM
	return nil
}


func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	dir += "/"
	return strings.Replace(dir, "\\", "/", -1)
}


func main() {
	CurExePath := GetCurrentDirectory()
	//启动参数
	svcFlag := flag.String("service", "", "start stop restart install uninstall")
	flag.Parse()

	//加载配置
	if cfg,err := config.ParseCommonCfgFromIni(CurExePath + "/config.ini"); err != nil {
		log.Fatal(err)
		panic(err)
	}else{
		os.Setenv("FWD_FROM", cfg.FromAddr)
		os.Setenv("FWD_TO", cfg.ToAddr)
		os.Setenv("CLI_ADDR", cfg.CliAddr)
		os.Setenv("CONN_TIMEOUT", strconv.Itoa(cfg.ConnTimeout))
	}

	//构建服务
	var s = &services{
		cfg: &service.Config{
			Name:             "zjjfwd",
			DisplayName:      "yunwang proxy zjj",
			Description:      "port forward for pda",
			WorkingDirectory: CurExePath,
			ChRoot:           CurExePath,
		}}

	sys := service.ChosenSystem()
	srv, err := sys.New(s, s.cfg)
	if err != nil {
		log.Fatalf("Init service error:%s\n", err.Error())
	}

	s.log, err = srv.SystemLogger(nil)
	if err != nil {
		log.Println("Set logger error:%s\n", err.Error())
	}

	//带参数则安装、卸载、启动、停止
	if len(*svcFlag) != 0 {
		err := service.Control(srv, *svcFlag)
		if err != nil {
			//log.Errorf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}

	//否则就运行
	err = srv.Run()
	if err != nil {
		log.Fatalf("Run programe error:%s\n", err.Error())
	}
}
