package handlers

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/JulianDavis/mattermost-bot/cmd/post"
)

// TODO: Fix this dupe
const mmDataFileName = "mattermost_bot.json"

type MainHandler struct {
	TemplatesFs embed.FS
}

func (mh MainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(mh.TemplatesFs, "templates/index.html")
	if err != nil {
		fmt.Printf("failed to parse the template: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		fmt.Printf("failed to execute the template: %s", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func SaveHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] Save handler called\n", time.Now().Format("2006-01-02 15:04:05"))
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the JSON data from the request body
	var data interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(mmDataFileName); err == nil {
		err := os.Rename(mmDataFileName, mmDataFileName+".old")
		if err != nil {
			// TODO: This is a non-critical warning, continue on
			fmt.Printf("[%s] Failed to rename file, overriding it...\n", time.Now().Format("2006-01-02 15:04:05"))
		}
	}

	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}
	err = os.WriteFile(mmDataFileName, file, 0644)
	if err != nil {
		http.Error(w, "Error writing to file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data saved successfully!"))
	fmt.Printf("[%s] Data saved successfully\n", time.Now().Format("2006-01-02 15:04:05"))

	//posts, err := loadPostData()
	//if err != nil {
	//	fmt.Println("Failed to load post data:", err)
	//  return
	//}
	//mmBot.client.SchedulePosts(mmBot.scheduler, posts)
}

func DataHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] Data handler called\n", time.Now().Format("2006-01-02 15:04:05"))
	w.Header().Set("Content-Type", "application/json")

	file, err := os.Open(mmDataFileName)
	if err != nil {
		http.Error(w, "Could not open data file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var rows []post.Post
	if err := json.NewDecoder(file).Decode(&rows); err != nil {
		http.Error(w, "Could not decode data", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(rows); err != nil {
		http.Error(w, "Could not encode data", http.StatusInternalServerError)
	}
}
