package slack

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

const (
	MAX_ERRORS         = 80
	NOTIFY_ERROR_COUNT = 10
)

type errorDropCount struct {
	sync.Mutex
	dropCount uint
}

type SlackChannel struct {
	start    sync.Once
	messages chan *messageWithErrCh
	errors   chan *Attachment
	errorDropCount
	c         *http.Client
	URL       URL // The URL returned by slack's webhook integration thing
	UserName  string
	UserIcon  URL
	UserEmoji string
	Channel   string
}

func (c *SlackChannel) errorGrabLoop(startWith *Attachment) (ret []*Attachment) {
	if startWith != nil {
		ret = append(ret, startWith)
	}
	for {
		select {
		case a := <-c.errors:
			ret = append(ret, a)
			if len(ret) >= MAX_ERRORS {
				return ret
			}
		default:
			return ret
		}
	}
}

func (c *SlackChannel) getAllErrors(startWith *Attachment) (ret []*Attachment) {
	ret = c.errorGrabLoop(startWith)
	c.errorDropCount.Lock()
	dropCount := c.errorDropCount.dropCount
	c.errorDropCount.dropCount = 0
	c.errorDropCount.Unlock()
	if len(ret) > NOTIFY_ERROR_COUNT || dropCount != 0 {
		var newList = make([]*Attachment, 0, len(ret)+1)
		var summaryAttachment = &Attachment{
			Fallback: fmt.Sprintf("Too many errors! Dropped %d", dropCount),
			Text:     fmt.Sprintf("<!channel|@channel> Too many errors! Dropped %d", dropCount),
			Colour:   Danger,
		}
		newList = append(newList, summaryAttachment)
		return append(newList, ret...)
	}
	return ret
}

func (c *SlackChannel) getNext() (m *Message, errCh chan<- error) {
	m = &Message{
		UserName:  c.UserName,
		UserIcon:  c.UserIcon,
		UserEmoji: c.UserEmoji,
		Channel:   c.Channel,
	}
	if m.Attachments = c.getAllErrors(nil); m.Attachments != nil {
		return m, nil
	}
	select {
	case a := <-c.errors:
		m.Attachments = c.getAllErrors(a)
		return m, nil
	case mwe := <-c.messages:
		return &mwe.message, mwe.errCh
	}
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
	if err = c.SendAttachment(prepareAttachment(messageToSend, colour, shortFields, longFields)); err != nil {
		return err
	}
	return nil
}

func (c *SlackChannel) SendRawMessage(m Message) (err error) {
	c.start.Do(c.Run)
	var timeout = time.NewTimer(5 * time.Second)
	var m2 = messageWithErrCh{errCh: make(chan error, 1), message: m}
	select {
	case c.messages <- &m2:
		if !timeout.Stop() {
			<-timeout.C
		}
		return <-m2.errCh
	case <-timeout.C: // Took too long to send
		return fmt.Errorf("This daemon is queueing too many slack messages")
	}
}

func (c *SlackChannel) Run() {
	var err error
	c.messages = make(chan *messageWithErrCh, 1)
	c.errors = make(chan *Attachment, MAX_ERRORS)
	go func() {
		var sleepTime = 1 * time.Second // Slack doesn't like more than one message per second on average
		for {
			var message, errCh = c.getNext()
			err = jsonPost(c.URL, message)
			if responseDetails, ok := err.(*Non200ResponseError); ok && responseDetails.Code == 429 {
				// We're being told to back off, presumably because lots of other instances on the same webhook are also spamming
				sleepTime = maxDuration(
					4*time.Second,
					time.Duration(responseDetails.CountSecondAgo)*time.Second,
					time.Duration(responseDetails.CountMinuteAgo/60)*time.Second,
				)
			}
			if errCh != nil {
				errCh <- err
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
	if errorToSend == nil {
		errorToSend = fmt.Errorf("SendError called with a nil error")
	}
	c.start.Do(c.Run)
	var a = prepareAttachment(errorToSend.Error(), colour, shortFields, longFields)
	select {
	case c.errors <- a:
		return nil
	default:
		c.errorDropCount.Lock()
		c.errorDropCount.dropCount++
		dc := c.errorDropCount.dropCount
		c.errorDropCount.Unlock()
		return fmt.Errorf("Too many errors, %d thrown away", dc)
	}
}
