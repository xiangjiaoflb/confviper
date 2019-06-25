package confviper

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

//NewConf ...
func NewConf(configPath, configName string, callback func(map[string]interface{}), arg ...interface{}) (*Conf, error) {
	//创建实例
	conf := Conf{
		configPath: configPath,
		configName: configName,
		confMap:    map[string]interface{}{},
	}

	//创建配置实例
	vp := viper.New()
	//配置初始化
	vp.SetConfigName(configName)
	vp.AddConfigPath(configPath + "/")

	//设置默认配置
	//读取传入的参数
	if len(arg) != 0 {
		kvmap, _ := arg[0].(map[string]interface{})
		conf.confMap = kvmap
	}

	//读配置
	err := vp.ReadInConfig()
	if err != nil {
		if vp.ConfigFileUsed() != "" {
			//文件在使用当中则返回错误
			return nil, err
		}

		//没有配置文件则写配置文件
		err = conf.writeConfFile()
		if err != nil {
			return nil, err
		}
		return NewConf(configPath, configName, callback, arg...)
	}

	conf.confMap = vp.AllSettings()
	callback(conf.confMap)

	//注册回调函数
	vp.OnConfigChange(func(e fsnotify.Event) {
		if e.Op == fsnotify.Write {
			conf.confMap = vp.AllSettings()
			callback(conf.confMap)
		}
	})

	//监听配置文件
	vp.WatchConfig()

	return &conf, nil
}

//Get 获取配置
func (pointer *Conf) Get(key string) (value interface{}, ok bool) {
	value, ok = pointer.confMap[key]
	return
}

func (pointer *Conf) Write(mapkv map[string]interface{}) error {
	for k, v := range mapkv {
		pointer.confMap[k] = v
	}

	return pointer.writeConfFile()
}

func (pointer *Conf) writeConfFile() error {
	//创建文件夹
	err := os.MkdirAll(pointer.configPath, os.ModePerm)
	if err != nil {
		return err
	}

	confstr := ""

	for k, v := range pointer.confMap {
		confstr = fmt.Sprintf("%s%s=", confstr, k)

		switch inter := v.(type) {
		case string:
			confstr = fmt.Sprintf("%s\"%s\"", confstr, inter)
		case int:
			confstr = fmt.Sprintf("%s%d", confstr, inter)
		case int64:
			confstr = fmt.Sprintf("%s%d", confstr, inter)
		case bool:
			confstr = fmt.Sprintf("%s%t", confstr, inter)
		case float64:
			confstr = fmt.Sprintf("%s%f", confstr, inter)
		default:
			confstr = fmt.Sprintf("%s\"\" # This type is not supported", confstr)
			log.Println("This type is not supported:", k)
		}

		confstr = fmt.Sprintf("%s\n", confstr)
	}

	//创建配置文件并写入内容
	err = ioutil.WriteFile(path.Join(pointer.configPath, pointer.configName+".toml"), []byte(confstr), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
