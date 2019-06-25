package confviper

//Conf 监听得配置文件
type Conf struct {
	configPath, configName string

	confMap map[string]interface{}
}
