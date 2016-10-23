package slack

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sort"
	"sync"
	"time"
)

type SlackChannel struct {
	start     sync.Once
	messages  chan *messageWithErrCh
	c         *http.Client
	URL       URL // The URL returned by slack's webhook integration thing
	UserName  string
	UserIcon  URL
	UserEmoji string
	Channel   string
}

func (c *SlackChannel) SendAttachment(a *Attachment) (err error) {
	var m Message

	m.UserName = c.UserName
	m.UserIcon = c.UserIcon
	m.UserEmoji = c.UserEmoji
	m.Channel = c.Channel

	m.Attachments = append(m.Attachments, a)

	return c.SendRawMessage(m)
}

func (c *SlackChannel) SendMessage(messageToSend string, colour Colour, shortFields map[string]string, longFields map[string]string) (err error) {
	var fields FieldList

	x := func(m map[string]string, short bool) {
		for k, v := range m {
			fields = append(fields, &Field{
				Title: k,
				Value: v,
				Short: short,
			})
		}
	}
	x(shortFields, true)
	x(longFields, false)
	sort.Sort(fields)
	if err = c.SendAttachment(&Attachment{
		Fallback: fmt.Sprintf("Error: %s", messageToSend),
		Text:     messageToSend,
		Colour:   colour,
		Fields:   fields,
	}); err != nil {
		return err
	}
	return nil
}

func (c *SlackChannel) SendRawMessage(m Message) (err error) {
	c.start.Do(c.Run)
	var timeout = time.NewTimer(5 * time.Second)
	var m2 = messageWithErrCh{errCh: make(chan error), message: m}
	select {
	case c.messages <- &m2:
		timeout.Stop()
		<-timeout.C
		return <-m2.errCh
	case <-timeout.C: // Took too long to send
		return fmt.Errorf("This daemon is queueing too many slack messages")
	}
}

func (c *SlackChannel) Run() {
	var err error
	c.messages = make(chan *messageWithErrCh)
	go func() {
		var sleepTime = 1 * time.Second // Slack doesn't like more than one message per second on average
		for message := range c.messages {
			err = jsonPost(c.URL, message)
			if responseDetails, ok := err.(*Non200ResponseError); ok && responseDetails.Code == 429 {
				// We're being told to back off, presumably because lots of other instances on the same webhook are also spamming
				if responseDetails.CountSecondAgo < 4 {
					sleepTime = time.Duration(4 * time.Second)
				}
				sleepTime = time.Duration(responseDetails.CountSecondAgo) * time.Second
			}
			time.Sleep(sleepTime)
			if sleepTime >= 2*time.Second {
				sleepTime -= 1 * time.Second
			}
		}
	}()
}

func (c *SlackChannel) OnPanic(hostname string) {
	r := recover()
	if r == nil {
		return
	}
	var errStr = fmt.Sprintf("%s panic: %s", hostname, r)
	var stackTrace = fmt.Sprintf("%s\n", debug.Stack())
	log.Printf("%s: %s\n", errStr, stackTrace)
	c.SendMessage(errStr, "danger",
		map[string]string{"Server": hostname},
		map[string]string{"Stack Trace": stackTrace},
	)
	panic(r)
}

func (c *SlackChannel) SendError(errorToSend error, colour Colour, shortFields map[string]string, longFields map[string]string) (err error) {
	return c.SendMessage(errorToSend.Error(), colour, shortFields, longFields)
}
