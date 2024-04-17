package block_header

var (
	// We hold here the account block height on which the hauler updated the blocks
	lastUpdatePrefix = []byte{0}

	blockHeaderPrefix = []byte{1}
)
