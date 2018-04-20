package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"./common"
	"./config"
	"github.com/zmb3/spotify"
)

var CONFIG_FILE = "key.ini"
var global_settings map[string]string

const redirectURI = "http://localhost:8080/callback"

var global_client spotify.Client

var (
	auth = spotify.NewAuthenticator(redirectURI,
		spotify.ScopeUserReadPrivate,
		spotify.ScopePlaylistReadPrivate)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {

	global_settings = config.Load_config(CONFIG_FILE,
		[]string{"Client ID", "Client Secret", "Old User"},
		[]string{"client_id", "client_secret", "old_user"})

	common.SectionTitle("Spotify Migration Assistant")

	global_client = authorizationHandler()

	common.SectionTitle("Step 1 : Gather Old Account Information")
	SavePlaylistsData()

}

func authorizationHandler() spotify.Client {

	auth.SetAuthInfo(global_settings["client_id"], global_settings["client_secret"])

	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// wait for auth to complete
	client := <-ch

	// use the client to make calls that require authorization
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	return *client
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
}

func SavePlaylistsData() {

	user, err := global_client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	var playlist_arr []spotify.SimplePlaylist

	for i := 0; i < 3; i++ {
		lim := 50

		off := i * lim

		if i == 0 {
			off = 0
		}

		opt := spotify.Options{
			Limit:  &lim,
			Offset: &off,
		}

		fmt.Println("\tOffset: ", off)

		playlist_list, err := global_client.CurrentUsersPlaylistsOpt(&opt)

		if err != nil {
			fmt.Println("Playlist Error!")
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}

		if playlist_list != nil {
			for _, playlist := range playlist_list.Playlists {
				// fmt.Println(playlist.Name, " ", playlist.Tracks.Total)
				playlist_arr = append(playlist_arr, playlist)
			}
		} else {
			fmt.Println("All playlists gathered!")
		}
	}

	CreateDir("tmp")

	file, err := os.Create("tmp/master.json")
	writer := bufio.NewWriter(file)

	for _, playlist := range playlist_arr {

		if err != nil {
			panic(err)
		}
		defer file.Close()

		info, _ := json.MarshalIndent(playlist, "", "    ")

		fmt.Fprintln(writer, string(info))

		if playlist.Owner.ID == user.ID {
			track_filepath := fmt.Sprintf("tmp/%s_tracks.json", filenameHandler(playlist.Name))
			track_file, _ := os.Create(track_filepath)

			track_writer := bufio.NewWriter(track_file)

			opts := spotify.Options{}
			fields := "items(added_at,track(name,id))"

			tracks, err := global_client.GetPlaylistTracksOpt(user.ID, playlist.ID, &opts, fields)

			if err != nil {
				panic(err)
			}

			for _, track := range tracks.Tracks {

				track_listing, _ := json.MarshalIndent(track, "", "    ")
				fmt.Fprintln(track_writer, string(track_listing))
			}

			defer track_file.Close()
		}
	}
}

/* Written by Siong-Ui Te, siongui.github.io */
func CreateDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func filenameHandler(filename string) string {

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(filename, "")

	return processedString
}
