package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mattermost/mattermost/server/public/model"
)

// Populated at build-time
var token string

func main() {
	client := model.NewAPIv4Client("http://127.0.0.1:8065")

	_, _, err := client.GetOldClientConfig(context.TODO(), "")
	if err != nil {
		fmt.Println("Failed to get config:", err)
		os.Exit(1)
	}

	client.SetToken(token)
	user, _, err := client.GetMe(context.TODO(), "")
	if err != nil {
		fmt.Println("Failed to log in:", err)
		os.Exit(1)
	}
	fmt.Printf("Logged in as %q with ID %q\n", user.Username, user.Id)

	channel, _, err := client.GetChannelByNameForTeamName(context.TODO(), "Announcements", "test", "")
	if err != nil {
		fmt.Println("Failed to get Announcements channel:", err)
		os.Exit(1)
	}
	fmt.Printf("Got Announcements channel ID: %q\n", channel.Id)

	post := model.Post{
		ChannelId: channel.Id,
		Message:   "Hello, World!",
	}

	_, _, err = client.CreatePost(context.TODO(), &post)
	if err != nil {
		fmt.Println("Failed to create post:", err)
	}

	fmt.Println("Post created!")
	os.Exit(0)
}
