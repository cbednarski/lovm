package unknown

import (
	"errors"
	"net"

	"github.com/cbednarski/lovm/core"
)

var ErrNoConfiguration = errors.New("no configuration found; you need to clone first")

type Unknown struct{}

func New(config *core.MachineConfig) *Unknown {
	return &Unknown{}
}

func (u *Unknown) Clone(source string) error {
	if source != "" {
		return errors.New("unrecognized virtualization format; specify a path to .vmx or .vbox")
	}
	return ErrNoConfiguration
}

func (u *Unknown) Start() error {
	return ErrNoConfiguration
}

func (u *Unknown) Stop() error {
	return ErrNoConfiguration
}

func (u *Unknown) Restart() error {
	return ErrNoConfiguration
}

func (u *Unknown) Delete() error {
	return ErrNoConfiguration
}

func (u *Unknown) IP() (net.IP, error) {
	return nil, ErrNoConfiguration
}

func (u *Unknown) Mount() error {
	return ErrNoConfiguration
}

func (u *Unknown) Found() bool {
	return false
}
