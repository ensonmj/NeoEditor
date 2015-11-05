package neoeditor

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/yuin/gopher-lua"
)

type ConfItem struct {
	Name  string
	Abbr  string
	Value interface{} // string or int or bool
	Doc   string
	File  string
	line  int
}

// TODO: support config overlap: buffer > filetype > global
type Conf map[string]ConfItem

// global config
var globalConfig = Conf{
	"expandtab": {Name: "expandtab", Abbr: "et", Value: false, Doc: "In Insert mode: Use the appropriate number of spaces to insert a <Tab>"},
	"tabstop":   {Name: "tabstop", Abbr: "ts", Value: 2, Doc: "Number of spaces that a <Tab> in the file counts for."},
}

func confInit() error {
	confFile := "./ned.lua"
	if err := confParse(confFile, globalConfig); err != nil {
		log.Warn("parse config file:%s err:%s", confFile, err)
		return err
	}
	log.Debug("parsed config:%v", globalConfig)

	return nil
}

func confParse(fPath string, conf Conf) error {
	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile(fPath); err != nil {
		return err
	}

	lv := L.GetGlobal("config").(*lua.LTable)
	lv.ForEach(func(key lua.LValue, value lua.LValue) {
		convertedKey := reflect.ValueOf(key)
		if convertedKey.Kind() != reflect.String {
			log.Warn("config key[%v] type is not string")
			return
		}
		strKey := convertedKey.String()

		convertedValue := reflect.ValueOf(value).Interface()
		if item, ok := conf[strKey]; ok {
			item.Value = convertedValue
			item.File = fPath
			// make sure update the original item
			conf[strKey] = item
			log.Debug("update [%s]%v", strKey, item)
		}
	})

	return nil
}

// config based on filetype
var ftConfig = map[string]Conf{}

func getFileTypeConf(filetype string) Conf {
	if filetype == "" {
		return globalConfig
	}

	if conf, ok := ftConfig[filetype]; ok {
		return conf
	}

	// try to find filetype config from disk
	fPath := fmt.Sprintf("./pkgs/%s.lua", filetype)
	_, err := os.Stat(fPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debug("filetype config:%s not exist", fPath)
		}
		return globalConfig
	}

	ftConf := globalConfig
	if err := confParse(fPath, ftConf); err != nil {
		log.Debug("parse filetype config:% err:%s", fPath, err)
		return globalConfig
	}

	// cached the filetype config
	ftConfig[filetype] = ftConf

	return ftConf
}

func getFileTypeConfItem(filetype, key string) ConfItem {
	conf := getFileTypeConf(filetype)
	return conf[key]
}

func getConfValueBool(conf ConfItem) bool {
	v := conf.Value.(lua.LBool)
	return bool(v)
}

func getConfValueInt(conf ConfItem) int {
	v := conf.Value.(lua.LNumber)
	return int(v)
}

func getConfValueString(conf ConfItem) string {
	v := conf.Value.(lua.LString)
	return string(v)
}
