package slack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

type URL string
type Colour string

type SlackChannel struct {
	c         *http.Client
	URL       URL // The URL returned by slack's webhook integration thing
	UserName  string
	UserIcon  URL
	UserEmoji string
	Channel   string
}

type Message struct {
	UserName    string        `json:"username,omitempty"`
	UserIcon    URL           `json:"icon_url,omitempty"`
	UserEmoji   string        `json:"icon_emoji,omitempty"`
	Channel     string        `json:"channel,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
	Text        string        `json:"text,omitempty"`
}

type Attachment struct {
	Fallback string   `json:"fallback"`
	Text     string   `json:"text,omitempty"`
	PreText  string   `json:"pretext,omitempty"`
	Colour   Colour   `json:"color,omitempty"` //// Can either be one of 'good', 'warning', 'danger', or any hex color code
	Fields   []*Field `json:"fields,omitempty"`
}

type Field struct {
	Title string `json:"title"`           // The title may not contain markup and will be escaped for you
	Value string `json:"value"`           // Text value of the field. May contain standard message markup and must be escaped as normal.  May be multi-line
	Short bool   `json:"short,omitempty"` // Optional flag indicating whether the `value` is short enough to be displayed side-by-side with other values
}

func jsonPost(url URL, data interface{}) (err error) {
	var resp *http.Response
	var errCh = make(chan error)

	r, w := io.Pipe()
	go func() {
		var err error
		j := json.NewEncoder(w)
		err = j.Encode(data)
		if err != nil {
			errCh <- err
		}
		w.Close()
		close(errCh)
	}()
	if resp, err = http.Post(string(url), "application/json", r); err != nil {
		return err
	}
	defer resp.Body.Close()
	err = <-errCh
	return err
}

func (c *SlackChannel) SendAttachment(a *Attachment) (err error) {
	var m Message // FIXME Pool instead of GC?

	m.UserName = c.UserName
	m.UserIcon = c.UserIcon
	m.UserEmoji = c.UserEmoji
	m.Channel = c.Channel

	m.Attachments = append(m.Attachments, a)

	return jsonPost(c.URL, &m)
}

type FieldList []*Field

func (f FieldList) Len() int {
	return len(f)
}
func (f FieldList) Less(i, j int) bool {
	if f[i].Short != f[j].Short {
		return f[i].Short
	}
	return f[i].Title < f[j].Title
}
func (f FieldList) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
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

func (c *SlackChannel) SendError(errorToSend error, colour Colour, shortFields map[string]string, longFields map[string]string) (err error) {
	return c.SendMessage(errorToSend.Error(), colour, shortFields, longFields)
}
