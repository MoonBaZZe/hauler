package node

import (
	"github.com/pkg/errors"
	"syscall"
)

var (
	ErrDataDirUsed     = errors.New("dataDir already used by another process")
	ErrNodeStopped     = errors.New("node not started")
	datadirInUseErrnos = map[uint]bool{11: true, 32: true, 35: true}
)

func convertFileLockError(err error) error {
	if errno, ok := err.(syscall.Errno); ok && datadirInUseErrnos[uint(errno)] {
		return ErrDataDirUsed
	}
	return err
}
