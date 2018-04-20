package utils

import (
	"github.com/pkg/errors"
)

func GetString(m map[string]interface{}, key string) string {
	raw, ok := m[key]

	if !ok {
		return ""
	}

	r, _ := raw.(string)

	return r
}

func ErrorFromArray(errs []error) error {
	err := errs[0]
	errs = errs[1:]

	for _, v := range errs {
		err = errors.Wrap(err, v.Error())
	}

	return err
}
