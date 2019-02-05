package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type containerInfo struct {
	Id string
	Names []string
	Image string
	ImageID string
	Command string
	Created int
}

type containerFilter struct {
	Name []string `json:"name"`
}

func main() {
	auth := os.Getenv( "PLEXAUTH" )

	http.HandleFunc( "/", func( w http.ResponseWriter, r *http.Request ) {
		fmt.Fprintf( w, "Nothing to see here" )
	})

	http.HandleFunc( "/restartplex", func( w http.ResponseWriter, r *http.Request ) {
		body, _ := ioutil.ReadAll(r.Body)

		if string(body) != auth {
			fmt.Printf( "[%s] Unauthorized Restart Attempt!\n", time.Unix( time.Now().Unix(), 0 ) )
			fmt.Fprintf( w, "Unauthorized" )
			return
		}

		fmt.Printf( "[%s] Restarting Plex!\n", time.Unix( time.Now().Unix(), 0 ) )
		fmt.Fprintf( w, "Authorized Request. Restarting Plex." )
		containerId, _ := getContainerId()
		result, err := restartContainer( containerId ) ; if err != nil {
			fmt.Print( err )
			return
		}
		if result != true {
			fmt.Println( "Could not restart container" )
			return
		}
		fmt.Println( "Restarted Container" )
	})

	fmt.Println( "Serving on port 80" )
	fmt.Println( "Auth is: " + auth )

	err := http.ListenAndServe( ":80", nil )
	if err != nil {
		panic( err )
	}

}

func restartContainer( id string ) (bool, error) {
	httpc := getDockerHttpClient()

	response, err := httpc.Post( getSocketUrl( "/containers/" + id + "/restart", nil ), "", nil ) ; if err != nil {
		return false, err
	}
	defer response.Body.Close()

	if 204 != response.StatusCode {
		return false, nil
	}

	return true, nil
}

func getContainerId() (string, error) {
	filter := containerFilter{}
	filter.Name = append( filter.Name, "media_plex_1" )

	httpc := getDockerHttpClient()

	response, err := httpc.Get( getSocketUrl( "/containers/json", filter ) ) ; if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var containers []containerInfo
	decoder := json.NewDecoder( response.Body )
	err = decoder.Decode( &containers ) ; if err != nil {
		return "", err
	}

	if len( containers ) < 1 {
		return "", nil
	}

	return containers[0].Id, nil
}

func getDockerHttpClient() http.Client {
	// Create an HTTP client using the unix socket as the transport since docker does http via the socket
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func( _ context.Context, _, _ string ) (net.Conn, error) {
				return net.Dial( "unix", "/var/run/docker.sock" )
			},
		},
	}

	return httpc
}

func getSocketUrl( path string, filters interface{} ) string {
	finalurl := "http://unix/"

	finalurl += strings.TrimPrefix( path, "/" )

	if filters != nil {
		jsonstring, _ := json.Marshal(filters)
		finalurl += "?filters=" + url.QueryEscape( string(jsonstring) )
	}

	return finalurl
}
