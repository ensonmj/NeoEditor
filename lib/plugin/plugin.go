package plugin

type PluginManager map[string]Plugin

type PluginInput struct {
	Text  [][]rune
	Chars []rune
}

type PluginOutput struct {
	Chars []rune
}

type Plugin interface {
	Init(name, guid string)
	Register(PluginManager)
	Handle(*PluginInput) (*PluginOutput, error)
	Release(PluginManager)
}

type DummyPlugin struct {
	Name, Guid string
}

func (dp *DummyPlugin) Init(name, guid string) {
	dp.Name, dp.Guid = name, guid
}

func (dp *DummyPlugin) Register(pm PluginManager) {
	pm[dp.Guid] = dp
}

func (dp *DummyPlugin) Handle(pi *PluginInput) (*PluginOutput, error) {
	return &PluginOutput{pi.Chars}, nil
}

func (dp *DummyPlugin) Release(pm PluginManager) {
	delete(pm, dp.Guid)
}
