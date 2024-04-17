package node

import (
	"github.com/MoonBaZZe/hauler/common"
)

var DefaultNodeConfig = common.Config{
	DataPath: common.DefaultDataDir(),

	GlobalState:               0,
	NoMEndpoints:              []string{"ws://127.0.0.1:35998"},
	ProducerKeyFileName:       "producer",
	ProducerKeyFilePassphrase: "",
	ProducerIndex:             0,
}
