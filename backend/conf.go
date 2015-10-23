package neoeditor

import (
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
var config = Conf{
	"expandtab": {Name: "expandtab", Abbr: "et", Value: false, Doc: "In Insert mode: Use the appropriate number of spaces to insert a <Tab>"},
	"tabstop":   {Name: "tabstop", Abbr: "ts", Value: 2, Doc: "Number of spaces that a <Tab> in the file counts for."},
}

// config based on filetype
var ftConfig = map[string]Conf{}

func confInit() error {
	confFile := "./ned.lua"
	if err := confParse(confFile, config); err != nil {
		log.Warn("parse config file:%s err:%s", confFile, err)
		return err
	}
	log.Debug("parsed config:%v", config)

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
