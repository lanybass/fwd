package config

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	ini "github.com/vaughan0/go-ini"
)

type CommonConf struct {
	FromAddr        string              `json:"from_addr"`
	ToAddr        string                 `json:"to_addr"`
	CliAddr         string              `json:"cli_addr"`
	ConnTimeout 	int               `json:"conn_timeout"`
}


func GetDefault() *CommonConf {
	return &CommonConf{
		FromAddr:        "127.0.0.1:9000",
		ToAddr:         "www.baidu.com:80",
		CliAddr:          "",//connect to ToAddr from a specified IP and Port when multi-ip exist
		ConnTimeout:3,//seconds
	}
}


func ParseCommonCfgFromIni(filePath string) (cfg *CommonConf, err error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	content := string(b)

	cfg, err = ParseCommonCfgFromString(content)
	if err != nil {
		return
	}

	return
}

func ParseCommonCfgFromString(content string) (cfg *CommonConf, err error) {
	cfg = GetDefault()

	conf, err := ini.Load(strings.NewReader(content))
	if err != nil {
		err = fmt.Errorf("parse ini conf file error: %v", err)
		return nil, err
	}

	var (
		tmpStr string
		ok     bool
		v		int64
	)
	if tmpStr, ok = conf.Get("common", "from_addr"); ok {
		cfg.FromAddr = tmpStr
	}


	if tmpStr, ok = conf.Get("common", "to_addr"); ok {
		cfg.ToAddr = tmpStr
	}

	if tmpStr, ok = conf.Get("common", "cli_addr"); ok {
		cfg.CliAddr = tmpStr
	}


	if tmpStr, ok = conf.Get("common", "conn_timeout"); ok {
		v, err = strconv.ParseInt(tmpStr, 10, 64)
		if err != nil {
			err = fmt.Errorf("Parse conf error: invalid conn_timeout")
			return
		}
		cfg.ConnTimeout = int(v)
	}

	err = cfg.Check()
	if err != nil {
		return
	}

	return
}

func (cfg *CommonConf) Check() (err error) {
	
	return
}
