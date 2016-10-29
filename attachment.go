package slack

import (
	"fmt"
	"sort"
)

func prepareAttachment(messageToSend string, colour Colour, shortFields, longFields map[string]string) (attachment *Attachment) {
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
	return &Attachment{
		Fallback: fmt.Sprintf("Error: %s", messageToSend),
		Text:     messageToSend,
		Colour:   colour,
		Fields:   fields,
	}
}
