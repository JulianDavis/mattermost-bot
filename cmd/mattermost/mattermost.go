package mattermost

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/JulianDavis/mattermost-bot/cmd/post"
	"github.com/go-co-op/gocron/v2"
	"github.com/mattermost/mattermost/server/public/model"
)

// YYYY-MM-DD hh:mm
const postDateTimeLayout = "2006-01-02 15:04"

type Mattermost struct {
	token     string
	team      string
	serverUrl url.URL
	client    *model.Client4
	user      *model.User
	ctx       context.Context
}

func New(token string, team string) *Mattermost {
	return &Mattermost{token: token, team: team, ctx: context.TODO()}
}

func (m *Mattermost) Login(url url.URL) error {
	m.client = model.NewAPIv4Client(url.String())
	m.client.SetToken(m.token)

	if err := m.pingServer(); err != nil {
		return err
	}
	m.serverUrl = url

	var err error
	m.user, _, err = m.client.GetMe(m.ctx, "")
	if err != nil {
		return err
	}

	return nil
}

func (m Mattermost) Post(channel string, message string) error {
	fmt.Printf("Creating a new post: %q\n", message)

	ch, _, err := m.client.GetChannelByNameForTeamName(m.ctx, channel, m.team, "")
	if err != nil {
		fmt.Println("Failed to get Announcements channel:", err)
		return err
	}
	fmt.Printf("Got Announcements channel ID: %q\n", ch.Id)

	post := model.Post{
		ChannelId: ch.Id,
		Message:   message,
	}

	_, _, err = m.client.CreatePost(m.ctx, &post)
	if err != nil {
		fmt.Println("Failed to create post:", err)
		return err
	}

	return nil
}

func (m Mattermost) pingServer() error {
	// According to the examples online this is how you ping the server
	_, _, err := m.client.GetOldClientConfig(m.ctx, "")
	return err
}

// TODO: This needs to do more, like check if a post is already scheduled and updated it if anything has changed, or add
// the post if it's missing. We also need to remove posts that don't exist in the latest version, meaning the user has
// removed them recently.
func (m Mattermost) SchedulePosts(scheduler gocron.Scheduler, posts []post.Post) error {
	for i, post := range posts {
		fmt.Printf("Adding post #%d to the schedule: %q\n", i, post.Message)

		postDateTimeString := fmt.Sprintf("%s %s", post.Date, post.Time)
		postDateTime, err := time.ParseInLocation(postDateTimeLayout, postDateTimeString, time.Local)
		if err != nil {
			fmt.Println("Failed to parse post time:", err)
			return err
		}
		//fmt.Printf("postDateTime: %v\n", postDateTime)

		job, err := scheduler.NewJob(
			gocron.OneTimeJob(
				gocron.OneTimeJobStartDateTime(postDateTime),
			),
			gocron.NewTask(
				m.Post, post.Channel, post.Message,
			),
		)
		if err != nil {
			fmt.Println("Error when scheduling new job:", err)
			return err
		}
		fmt.Printf("Job successfully created! [%v]\n", job.ID())
	}

	return nil
}
