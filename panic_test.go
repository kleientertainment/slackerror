package slack

import (
	"fmt"
	"testing"
)

const PANIC_OBJ = "TestPanic"

func panickingFunc() {
	panic(PANIC_OBJ)
}

func TestPanicPropagation(t *testing.T) {
	slackCh := &SlackChannel{URL: URL(""), UserName: "UnitTest", Channel: "TestChannel"}

	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Slack OnPanic haven't propagated the panic")
			}
			if r != PANIC_OBJ {
				t.Fatal("Slack OnPanic have panicked")
			}
			fmt.Println(r)
		}()
		defer slackCh.OnPanic("Test")

		panickingFunc()
	}()
}
