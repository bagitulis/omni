package lazadasdk

import "fmt"

func ErrBadRequest(msg string) error {
	return fmt.Errorf("bad request: %s", msg)
}
