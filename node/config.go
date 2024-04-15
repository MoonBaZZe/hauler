package node

import (
	"github.com/MoonBaZZe/hauler/common"
)

var DefaultNodeConfig = common.Config{
	DataPath: common.DefaultDataDir(),

	ProducerKeyFileName:       "producer",
	ProducerKeyFilePassphrase: "",
	ProducerIndex:             0,
}
