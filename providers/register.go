package providers

type ProviderKey struct {
	Type string
	Name string
}

// String returns a "type/name" formatted string representation.
func (pk ProviderKey) String() string {
	return pk.Type + "/" + pk.Name
}

var registeredProviders = make(map[string]map[string]Provider)

// Register 注册一个 provider
func Register(provider Provider) {
	providerType, providerName := provider.Type(), provider.Name()
	if registeredProviders[providerType] == nil {
		registeredProviders[providerType] = make(map[string]Provider)
	}
	registeredProviders[providerType][providerName] = provider
}

func GetProvider(providerKey ProviderKey) Provider {
	return registeredProviders[providerKey.Type][providerKey.Name]
}

func Unregister(providerKey ProviderKey) {
	delete(registeredProviders[providerKey.Type], providerKey.Name)
}

// Reset 清空所有已注册的 Provider（主要用于测试）。
func Reset() {
	registeredProviders = make(map[string]map[string]Provider)
}
