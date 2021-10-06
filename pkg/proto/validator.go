package proto

import "github.com/pkg/errors"

func (x *HelloRequest) Validate() error {
	if x.Name == "" {
		return errors.New("name can't be empty")
	}

	return nil
}
