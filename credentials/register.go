package credentials

type Factory func([]byte) Credential

var factoryRegistry = make(map[string]Factory)

// Register 注册凭据工厂
func Register(name string, factory Factory) {
	factoryRegistry[name] = factory
}

// GetFactory 获取凭据工厂
func GetFactory(name string) Factory {
	return factoryRegistry[name]
}
