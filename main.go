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

const CONFIG_FILE = "key.ini"
const redirectURI = "http://localhost:8080/callback"

var global_settings map[string]string

var (
	auth = spotify.NewAuthenticator(redirectURI,
		spotify.ScopeUserReadPrivate,
		spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistModifyPublic)
	ch    = make(chan *spotify.Client)
	state = "abc123"
)

func main() {

	global_settings = config.Load_config(CONFIG_FILE,
		[]string{"Client ID", "Client Secret", "Old User", "New User"},
		[]string{"client_id", "client_secret", "old_user", "new_user"})

	common.SectionTitle("Spotify Migration Assistant")

	client := authorizationHandler("/callback", "/")
	client.AutoRetry = true

	common.SectionTitle("Step 1 : Gather Old Account Information")
	SavePlaylistsData(&client)

	old_playlists := SavePlaylistsData(&client)

	common.SectionTitle("Step 2 : Populate new Account")

	newUserClient := authorizationHandler("/secondcallback", "/second/")
	newUserClient.AutoRetry = true

	PopulateNewAcct(&client, &newUserClient, old_playlists)

}

func PopulateNewAcct(oldclient *spotify.Client, newclient *spotify.Client, playlists []spotify.SimplePlaylist) {

	DeleteAll(newclient)

	file, _ := os.Create("tmp/error.txt")
	writer := bufio.NewWriter(file)

	user, _ := oldclient.CurrentUser()

	for _, playlist := range playlists {

		/* User created playlists else follwing playlists */
		if playlist.Owner.ID == user.ID {

			/* Grab the tracks from the old playlist */
			tracks, err := oldclient.GetPlaylistTracksOpt(user.ID, playlist.ID, &spotify.Options{}, "items(track(name,id))")

			if err != nil {
				panic(err)
			}

			/* Add each track to the new user or save an error */
			fmt.Println("Attempting to add: ", playlist.Name)

			newPlaylist, err := newclient.CreatePlaylistForUser(global_settings["new_user"], playlist.Name, true)

			if err != nil {
				log.Fatal(err)
			}
			for _, track := range tracks.Tracks {

				_, err := newclient.AddTracksToPlaylist(global_settings["new_user"], newPlaylist.ID, track.Track.ID)

				if err != nil {
					err_msg := fmt.Sprintf("Could not add %s to playlist %s.", track.Track.Name, playlist.Name)
					fmt.Fprintln(writer, err_msg)
					fmt.Println(err_msg)
				}
			}

		} else {

			err := newclient.FollowPlaylist(spotify.ID(playlist.Owner.ID), playlist.ID, true)

			if err != nil {
				err_msg := fmt.Sprintf("Could not follow %s.", playlist.Name)
				fmt.Fprintln(writer, err_msg)
				fmt.Println(err_msg)
			}
		}
	}
}

/*
 * Save Playlist Data
 *
 * Desc: Checks for logged in. If the user has been logged in, an attempt is
 *	     made to grab all of the playlists from the user. If that is successful,
 *		 then the data collected is stored in JSON format in an tmp directory
 *		 to be used for later processing and or backing up of Spotify
 *		 information
 *
 *
 */
func SavePlaylistsData(client *spotify.Client) []spotify.SimplePlaylist {

	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)

	fmt.Println("Grabbing all playlists")

	/* Create an array of playlists */
	playlists := GrabPlaylists(client)

	fmt.Println("Found ", len(playlists), " playlists.")

	/* Save a dataset of JSON information regarding each playlist */
	CreatePlaylistDataset(user.ID, playlists)

	return playlists
}

func authorizationHandler(callbk string, path string) spotify.Client {

	auth.SetAuthInfo(global_settings["client_id"], global_settings["client_secret"])

	http.HandleFunc(callbk, completeAuth)

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})

	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)

	fmt.Println("Please log in to Spotify by visiting the following page in your browser:\n", url)

	// wait for auth to complete
	client := <-ch

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
	var client spotify.Client

	client = auth.NewClient(tok)

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

/*
 * Quickly use a regular expression to compress file names and remove special
 * characters.
 */
func filenameHandler(filename string) string {

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(filename, "")

	return processedString
}

func GrabPlaylists(client *spotify.Client) []spotify.SimplePlaylist {

	lim, offset, retrieved := 50, 0, 50
	var playlists []spotify.SimplePlaylist

	for i := 0; retrieved != 0; i++ {
		offset = i * lim

		opt := spotify.Options{
			Limit:  &lim,
			Offset: &offset,
		}

		payload, err := client.CurrentUsersPlaylistsOpt(&opt)

		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return nil
		}

		retrieved = len(payload.Playlists)

		if retrieved > 0 {
			for _, playlist := range payload.Playlists {
				playlists = append(playlists, playlist)
			}
		}
	}

	return playlists
}

func CreatePlaylistDataset(ownerID string, playlists []spotify.SimplePlaylist) {

	CreateDir("tmp")
	// CreateDir("tmp/tracklistings")

	file, err := os.Create("tmp/master.json")
	writer := bufio.NewWriter(file)

	if err != nil {
		panic(err)
	}

	for _, playlist := range playlists {

		info, _ := json.MarshalIndent(playlist, "", "    ")
		fmt.Fprintln(writer, string(info))

		// if playlist.Owner.ID == ownerID {

		// 	track_filepath := fmt.Sprintf("tmp/tracklistings/%s_tracks.json", filenameHandler(playlist.Name))
		// 	track_file, _ := os.Create(track_filepath)

		// 	track_writer := bufio.NewWriter(track_file)

		// 	opts := spotify.Options{}
		// 	fields := "items(added_at,track(name,id,album(name)))"

		// 	tracks, err := client.GetPlaylistTracksOpt(ownerID, playlist.ID, &opts, fields)

		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	for _, track := range tracks.Tracks {

		// 		track_listing, _ := json.MarshalIndent(track, "", "    ")
		// 		fmt.Fprintln(track_writer, string(track_listing))
		// 	}
		// }
	}
}

func DeleteAll(client *spotify.Client) {

	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Unfollowing all playlists for user: ", user.ID)
	var value string
	fmt.Scanln(&value)

	playlists, err := client.GetPlaylistsForUser(global_settings["new_user"])

	if err != nil {
		panic(err)
	}

	for _, playlist := range playlists.Playlists {

		err := client.UnfollowPlaylist(spotify.ID(playlist.Owner.ID), playlist.ID)

		if err != nil {
			fmt.Println("Could not remove ", playlist.Name)
			panic(err)
		}
	}

}
