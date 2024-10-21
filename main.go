package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"

	"github.com/JulianDavis/mattermost-bot/cmd/handlers"
	"github.com/JulianDavis/mattermost-bot/cmd/mattermost"
	"github.com/JulianDavis/mattermost-bot/cmd/post"
	"github.com/go-co-op/gocron/v2"
)

type MattermostBot struct {
	httpAddr  string
	scheduler gocron.Scheduler
	client    *mattermost.Mattermost
}

//go:embed templates
var templatesFs embed.FS

//go:embed templates/static
var staticFiles embed.FS

// Populated at build-time
var mmToken string

// Temporarily hardcoded configuration values
const mmDataFileName = "mattermost_bot.json"

const mmServerScheme = "http"
const mmServerIP = "127.0.0.1"
const mmServerPort = 8065
const mmTeamName = "test"

const httpServerIp = "127.0.0.1"
const httpServerPort = 8001

/* TODO: Fix hardcoded args and paths
 *       This file could probably just be opened at startup and held open, no one else should be reading or writing to it
 *	     Eventually this should just be using SQLite or whatever
 */
func loadPostData() ([]post.Post, error) {
	file, err := os.Open(mmDataFileName)
	if err != nil {
		fmt.Println("Failed to create static files:", err)
		return []post.Post{}, err
	}
	defer file.Close()

	var posts []post.Post
	if err := json.NewDecoder(file).Decode(&posts); err != nil {
		fmt.Println("Failed to create static files:", err)
		return []post.Post{}, err
	}

	return posts, nil
}

func main() {
	var err error

	// Create bot
	mmBot := MattermostBot{}
	mmBot.httpAddr = fmt.Sprintf("%s:%d", httpServerIp, httpServerPort)

	// Login to Mattermost
	mmBot.client = mattermost.New(mmToken, mmTeamName)
	mmServerURL := url.URL{
		Scheme: mmServerScheme,
		Host:   fmt.Sprintf("%s:%d", mmServerIP, mmServerPort),
	}
	mmBot.client.Login(mmServerURL)

	// Create scheduler
	mmBot.scheduler, err = gocron.NewScheduler()
	if err != nil {
		fmt.Println("Failed to create a new scheduler", err)
		os.Exit(1)
	}
	defer func() { _ = mmBot.scheduler.Shutdown() }()
	posts, err := loadPostData()
	if err != nil {
		fmt.Println("Failed to load post data:", err)
		os.Exit(1)
	}
	mmBot.client.SchedulePosts(mmBot.scheduler, posts)
	mmBot.scheduler.Start()

	// Start HTTP server
	http.Handle("/", handlers.MainHandler{
		TemplatesFs: templatesFs,
	})
	staticFs, err := fs.Sub(staticFiles, "templates/static")
	if err != nil {
		fmt.Println("Failed to create static files:", err)
		os.Exit(1)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFs))))

	http.HandleFunc("/save", handlers.SaveHandler)
	http.HandleFunc("/data", handlers.DataHandler)

	fmt.Println("HTTP Server Started...")
	if err := http.ListenAndServe(mmBot.httpAddr, nil); err != nil {
		panic(err)
	}
}
