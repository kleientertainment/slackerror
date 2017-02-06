package slack

import (
	"fmt"
	"log"
	"runtime"
)

func (c *SlackChannel) Warning(format string, stuff ...interface{}) {
	var callingfuncline string
	log.Printf(format, stuff...)
	x := fmt.Errorf(format, stuff...)
	_, file, line, ok := runtime.Caller(1)
	if ok {
		callingfuncline = fmt.Sprintf("%s:%d", file, line)
	}
	go c.SendError(x, "warning", map[string]string{"Caller": callingfuncline}, nil)
	//"Server": serverName,
}
