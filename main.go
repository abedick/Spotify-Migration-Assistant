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

	common.SectionTitle("Step 2 : Login with new Account")

}

func SavePlaylistsData() {

	user, err := global_client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	/* Create an array of playlists */
	playlists := GrabPlaylists()

	/* Save a dataset of JSON information regarding each playlist */
	CreatePlaylistDataset(user.ID, playlists)

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

func GrabPlaylists() []spotify.SimplePlaylist {

	lim := 50
	offset := 0
	retrieved := 50
	count := 0
	var playlists []spotify.SimplePlaylist

	for i := 0; retrieved != 0; i++ {
		offset = i * lim

		opt := spotify.Options{
			Limit:  &lim,
			Offset: &offset,
		}

		payload, err := global_client.CurrentUsersPlaylistsOpt(&opt)

		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return nil
		}

		retrieved = len(payload.Playlists)

		if retrieved > 0 {
			for _, playlist := range payload.Playlists {
				playlists = append(playlists, playlist)
				count++
			}
		}
	}

	fmt.Println("Retrieved ", count, " playlists.")

	return playlists
}

func CreatePlaylistDataset(ownerID string, playlists []spotify.SimplePlaylist) {

	CreateDir("tmp")
	CreateDir("tmp/tracklistings")

	file, err := os.Create("tmp/master.json")
	writer := bufio.NewWriter(file)

	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, playlist := range playlists {

		info, _ := json.MarshalIndent(playlist, "", "    ")

		fmt.Fprintln(writer, string(info))

		if playlist.Owner.ID == ownerID {
			track_filepath := fmt.Sprintf("tmp/tracklistings/%s_tracks.json", filenameHandler(playlist.Name))
			track_file, _ := os.Create(track_filepath)

			track_writer := bufio.NewWriter(track_file)

			opts := spotify.Options{}
			fields := "items(added_at,track(name,id,album(name)))"

			tracks, err := global_client.GetPlaylistTracksOpt(ownerID, playlist.ID, &opts, fields)

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
