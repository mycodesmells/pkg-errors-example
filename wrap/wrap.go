package wrap

import (
	"github.com/mycodesmells/pkg-errors-example/common"
	"github.com/pkg/errors"
)

func CallA() error {
	return errors.Wrap(CallB(), "Error from CallA")
}

func CallB() error {
	return errors.Wrap(CallC(), "Error from CallB")
}

func CallC() error {
	return common.MyError{Msg: "Error from CallC"}
}
