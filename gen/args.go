package gen

type structConfig struct {
	Struct structDef
	Config config
}

func newStructConfig(s structDef, c config) *structConfig {
	return &structConfig{Struct: s, Config: c}
}

type interfaceConfig struct {
	Interface interfaceDef
	Config    config
}

func newInterfaceConfig(i interfaceDef, c config) *interfaceConfig {
	return &interfaceConfig{Interface: i, Config: c}
}
