package config

import (
	"fmt"
	"strconv"
)

// Version defines the version of the Config.
type Version int

const (
	// V1 represents config version 1
	V1 Version = iota + 1
	// V2 represents config version 2
	V2
)

// ErrInvalidConfigVersion - if the config version is not valid
var ErrInvalidConfigVersion error = fmt.Errorf("invalid config version")

// NewConfigVersionValue returns Version set with default value
func NewConfigVersionValue(val Version, p *Version) *Version {
	*p = val
	return p
}

// Set sets the value of the named command-line flag.
func (c *Version) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*c = Version(v)
	if err != nil {
		return err
	}
	if !c.IsValid() {
		return ErrInvalidConfigVersion
	}
	return nil
}

// Type returns a string that uniquely represents this flag's type.
func (c *Version) Type() string {
	return "int"
}

func (c *Version) String() string {
	return strconv.Itoa(int(*c))
}

// IsValid returns if its a valid config version
func (c Version) IsValid() bool {
	return c == V1 || c == V2
}
