package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"./common"
	"./config"
	"github.com/zmb3/spotify"
)

var CONFIG_FILE = "key.ini"
var global_settings map[string]string

const redirectURI = "http://localhost:8080/callback"

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {

	common.SectionTitle("Spotify Migration Assistang")

	config_param_alias := []string{"Client ID", "Client Secret"}
	config_param_mapped := []string{"client_id", "client_secret"}

	global_settings = config.Load_config(CONFIG_FILE, config_param_alias, config_param_mapped)

	// client := authorizationHandler()

	common.SectionTitle("Step 1 : Gather Old Account Information")

	SavePlaylistsData()

	// for i := 0; i < 20; i++ {
	// 	lim := 20
	// 	off := 20 * i

	// 	opt := spotify.Options{
	// 		Limit:  &lim,
	// 		Offset: &off,
	// 	}

	// 	playlist_list, err := client.CurrentUsersPlaylistsOpt(&opt)

	// 	if err != nil {
	// 		fmt.Println("Playlist Error!")
	// 		fmt.Fprintf(os.Stderr, err.Error())
	// 		return
	// 	}

	// 	for _, playlist := range playlist_list.Playlists {
	// 		fmt.Println(playlist.Name, " ", playlist.ID)
	// 	}
	// }

	// fmt.Println("User ID:", user.ID)
	// fmt.Println("Display name:", user.DisplayName)
	// fmt.Println("Spotify URI:", string(user.URI))
	// fmt.Println("Endpoint:", user.Endpoint)
	// fmt.Println("Followers:", user.Followers.Count)
}

func authorizationHandler() spotify.Client {

	auth.SetAuthInfo(global_settings["client_id"], global_settings["client_secret"])

	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go http.ListenAndServe(":8080", nil)

	global_settings = config.Update_config(global_settings, "session_date", time.Now().String())

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

}
