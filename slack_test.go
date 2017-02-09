package slack

import (
	"fmt"
	"testing"
)

func TestSlack(t *testing.T) {
	var err error
	var c = SlackChannel{URL: "https://hooks.slack.com/services/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", UserName: "UnitTestBot", Channel: "@nome"}

	if err = c.SendError(fmt.Errorf("Error: Mike is too awesome"), "warning", []*Field{
		{
			Title: "Server",
			Value: "sp2.nome.ca",
			Short: true,
		},
		{
			Title: "Client",
			Value: "sp2.nome.ca",
			Short: true,
		},
	}); err != nil {
		t.Fatalf("Error sending attachment to slack: %s\n", err)
	}
}
