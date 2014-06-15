// usage: go run demo.go OR go run demo.go <number>
package main

import (
	"os"
	"fmt"
	netUrl "net/url"
	fb "github.com/huandu/facebook"
	"net/http"
)

var globalApp = fb.New(clientId, os.Args[1])

var leToken = ""

const clientId = "509571419145066"

type Post struct {
	Id string `facebook:",required"`
	Message string
	CreatedTime string
}

type Paging struct {
	Previous, Next string
}

func processPosts (url string, total int) int {
	fmt.Println(url)
	res, err := fb.Get(url, fb.Params{
		// "access_token": ACCESS_TOKEN,
	})

	if (err != nil) {
		panic(err)
	}

	var posts []Post
	var paging Paging
	res.DecodeField("data", &posts)
	res.DecodeField("paging", &paging)

	if (len(posts) == 0) {
		return total
	}

	u, _ := netUrl.Parse(paging.Next)
	path_without_version_prefix := u.Path[len("/v2.0"):]
	next := path_without_version_prefix + "?" + u.RawQuery
	return processPosts(next, total + len(posts))

}

func authOut(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w,
	 r,
	 "https://www.facebook.com/dialog/oauth?client_id=509571419145066&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fauth%2Fin%2F",
	 http.StatusMovedPermanently)
}


func authIn(w http.ResponseWriter, r *http.Request) {
	token, err := globalApp.ParseCode(r.FormValue("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	leToken = token
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func index(w http.ResponseWriter, r *http.Request) {
	if leToken == "" {
		fmt.Fprintf(w, "<a href=\"/auth/out\">OAuth</a>")
	} else {
		session := globalApp.Session(leToken)
		res, err := session.Get("/me", nil)
		if err != nil {
			fmt.Println(err)
		}

		var firstName string
		res.DecodeField("first_name", &firstName)
		fmt.Fprintf(w, "%s\n%s", leToken, firstName)
	}
}

func main () {
	fb.Version = "v2.0"

	globalApp.RedirectUri = "http://localhost:8080/auth/in/"

	// total := processPosts("/163844093817909/feed?limit=" + limit, 0)
	// fmt.Println("Total was", total);

	http.HandleFunc("/auth/out/", authOut)
	http.HandleFunc("/auth/in/", authIn)
	http.HandleFunc("/", index)	
	http.ListenAndServe(":8080", nil)
}