package message

import (
	"reflect"
	"testing"

	"github.com/go-epaxos/message/example"
)

func TestRegister(t *testing.T) {
	register(0, reflect.TypeOf(example.PreAccept{}))
}
