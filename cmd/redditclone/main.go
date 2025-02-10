package main

import (
	"log"
	"net/http"

	"github.com/l-ILINDAN-l/BackendCloneReddit/internal/api"
)

const (
	pathToIndexHTML   = "../../static/html/index.html"
	pathToStaicSource = "../../static/"
)

func main() {
	// Creating a server object with data storage initialization
	// You can do this via the environment (.env)
	server := api.NewServer(":3000", "../../configs/config_server.json")

	// Connecting api methods to the server object
	server.Router.HandleFunc("/api/register", server.RegisterHandler).Methods("POST")                        // registration
	server.Router.HandleFunc("/api/login", server.LoginHandler).Methods("POST")                              // login
	server.Router.HandleFunc("/api/posts/", server.GetPostsHandler).Methods("GET")                           // list of all posts
	server.Router.HandleFunc("/api/posts", server.PostPostsHandler).Methods("POST")                          // adding a post
	server.Router.HandleFunc("/api/posts/{CATEGORY_NAME}", server.GetPostsByCategory).Methods("GET")         // a list of posts in a specific category
	server.Router.HandleFunc("/api/post/{POST_ID}", server.GetPostsByID).Methods("GET")                      // details of the post with comments
	server.Router.HandleFunc("/api/post/{POST_ID}", server.AddCommentPost).Methods("POST")                   //  adding a comment
	server.Router.HandleFunc("/api/post/{POST_ID}/{COMMENT_ID}", server.DeleteCommentPost).Methods("DELETE") // deleting a comment
	server.Router.HandleFunc("/api/post/{POST_ID}/upvote", server.UpvotePost).Methods("GET")                 // the rating of the post is up
	server.Router.HandleFunc("/api/post/{POST_ID}/downvote", server.DownvotePost).Methods("GET")             // the rating of the post is down
	server.Router.HandleFunc("/api/post/{POST_ID}/unvote", server.UnvotePost).Methods("GET")                 // voice cancellation
	server.Router.HandleFunc("/api/post/{POST_ID}", server.DeletePost).Methods("DELETE")                     // deleting a post
	server.Router.HandleFunc("/api/user/{USER_LOGIN}", server.GetPostsByUser)                                // getting all the posts of a specific user

	// Handler for issuing index.html on the root route "/"
	server.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, pathToIndexHTML)
	})

	// Handler for static files (JS, CSS)
	staticFiles := http.StripPrefix("/static/", http.FileServer(http.Dir(pathToStaicSource)))
	server.Router.PathPrefix("/static/").Handler(staticFiles)

	if err := http.ListenAndServe(server.Addr, server.Router); err != nil {
		log.Fatalf("Error ListenAndServe err: %s", err)
	}
}
