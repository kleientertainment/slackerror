package slack

type URL string
type Colour string

const (
	Good    = "good"
	Warning = "warning"
	Danger  = "danger"
)

type Message struct {
	Attachments []*Attachment `json:"attachments,omitempty"`
	Channel     string        `json:"channel,omitempty"`
	Markdown    bool          `json:"mrkdwn,omitempty"`
	Text        string        `json:"text,omitempty"`
	UserEmoji   string        `json:"icon_emoji,omitempty"`
	UserIcon    URL           `json:"icon_url,omitempty"`
	UserName    string        `json:"username,omitempty"`
}

type Attachment struct {
	AuthorName string   `json:"author_name,omitempty"`
	AuthorLink string   `json:"author_link,omitempty"`
	AuthorIcon string   `json:"author_icon,omitempty"`
	Colour     Colour   `json:"color,omitempty"` //// Can either be one of 'good', 'warning', 'danger', or any hex color code
	Fallback   string   `json:"fallback"`
	Fields     []*Field `json:"fields,omitempty"`
	Footer     string   `json:"footer,omitempty"`
	FooterIcon string   `json:"footer_icon,omitempty"`
	ImageURL   string   `json:"image_url,omitempty"`
	MarkdownIn []string `json:"mrkdwn_in,omitempty"` /// Can either be one of 'pretext', 'text' or 'fields'
	PreText    string   `json:"pretext,omitempty"`
	Text       string   `json:"text,omitempty"`
	Title      string   `json:"title,omitempty"`
	TitleLink  string   `json:"title_link,omitempty"`
	ThumbURL   string   `json:"thumb_url,omitempty"`
	Timestamp  uint64   `json:"ts,omitempty"`
}

type Field struct {
	Title string `json:"title"`           // The title may not contain markup and will be escaped for you
	Value string `json:"value"`           // Text value of the field. May contain standard message markup and must be escaped as normal.  May be multi-line
	Short bool   `json:"short,omitempty"` // Optional flag indicating whether the `value` is short enough to be displayed side-by-side with other values
}

type FieldList []*Field

func (f FieldList) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f FieldList) Len() int      { return len(f) }
func (f FieldList) Less(i, j int) bool {
	if f[i].Short != f[j].Short {
		return f[i].Short
	}
	return f[i].Title < f[j].Title
}
