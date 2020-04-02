package chat

type Channel struct {
	ID       string             `json:"id"`
	UserID   string             `json:"user_id"`
	Clients  map[string]*Client `json:"-"`
	Messages []*Message         `json:"messages"`
}
