package post

type Post struct {
	Date    string `json:"date"`
	Time    string `json:"time"`
	Channel string `json:"channel"`
	Message string `json:"message"`
}
