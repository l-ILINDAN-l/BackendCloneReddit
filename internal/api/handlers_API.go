package api

import (
	"errors"
	"time"

	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/l-ILINDAN-l/BackendCloneReddit/internal/models"
	"github.com/l-ILINDAN-l/BackendCloneReddit/internal/repository"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PostData struct {
	Category string `json:"category"`
	Text     string `json:"text"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	Type     string `json:"type"`
}
type CommentData struct {
	Comment string `json:"comment"`
}

func (server *Server) RegisterHandler(w http.ResponseWriter, r *http.Request) {

	// Getting data from the Request Payload
	var creds Credentials
	errJSONDecode := json.NewDecoder(r.Body).Decode(&creds)
	if errJSONDecode != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	// Getting a password that has already been hashed and checked for hash generation errors
	password, errHashPassword := HashPassword(creds.Password)
	if errHashPassword != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Printf("RegisterHandler HashPassword err: %s", errHashPassword)
		return
	}
	// Getting a username
	username := creds.Username
	// Checking for the existence of such a user
	if _, errGetByUsername := server.MemServ.UserRepo.GetByUsername(username); !errors.Is(errGetByUsername, repository.ErrUserNotFound) {
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := struct {
			Errors []struct {
				Location string `json:"location"`
				Param    string `json:"param"`
				Value    string `json:"value"`
				Msg      string `json:"msg"`
			} `json:"errors"`
		}{
			Errors: []struct {
				Location string `json:"location"`
				Param    string `json:"param"`
				Value    string `json:"value"`
				Msg      string `json:"msg"`
			}{
				{
					Location: "body",
					Param:    "username",
					Value:    username,
					Msg:      "already exists",
				},
			},
		}
		if errJSONEncode := json.NewEncoder(w).Encode(errorResponse); errJSONEncode != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("RegisterHandler UserRepo Encode errorResponse err: %s", errJSONEncode)
		}
		return
	}
	// Generating a unique ID
	genID, errID := GenerateID()
	if errID != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("RegisterHandler GenerateID err: %s", errID)
		return
	}
	// Entering such a user into the database and checking for the presence, if any, an error is returned
	if errUserRepoCreate := server.MemServ.UserRepo.Create(&models.User{
		ID:       genID,
		Username: username,
		Password: password,
	}); errUserRepoCreate != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("RegisterHandler UserRepo Create err: %s", errUserRepoCreate)
		return
	}

	tokenString, errGenToken := GenerateToken(username, genID, []byte(server.KeyJWT))

	if errGenToken != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("RegisterHandler GenerateToken err: %s", errGenToken)
		return
	}

	// Creating a session
	session := models.Session{
		Token:  tokenString,
		UserID: genID,
	}
	// Entering the created session into the database with a check for existence, if there has already been one, an error is thrown
	if errSessionRepoCreate := server.MemServ.SessionRepo.Create(&session); errSessionRepoCreate != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("RegisterHandler SessionRepo Create err: %s", errSessionRepoCreate)
		return
	}
	// Session recording of the response
	w.WriteHeader(http.StatusCreated)
	if errJSONEncode := json.NewEncoder(w).Encode(session); errJSONEncode != nil {
		log.Printf("RegisterHandler Encode session err: %s", errJSONEncode)
	}
}

func (server *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Getting data from the Request Payload
	var creds Credentials
	errCredentials := json.NewDecoder(r.Body).Decode(&creds)
	if errCredentials != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Getting a user by name and verifying their existence
	username := creds.Username
	var user *models.User
	user, errGetByUsername := server.MemServ.UserRepo.GetByUsername(username)
	if errors.Is(errGetByUsername, repository.ErrUserNotFound) {
		w.WriteHeader(http.StatusUnauthorized)
		if errJSONEncode := json.NewEncoder(w).Encode(
			struct {
				Message string `json:"message"`
			}{
				Message: "user not found",
			}); errJSONEncode != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("LoginHandler Encode Message err: %s", errJSONEncode)
		}
		return
	}

	// Password matching check
	if !CheckPassword(user.Password, creds.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		if errJSONEncode := json.NewEncoder(w).Encode(
			struct {
				Message string `json:"message"`
			}{
				Message: "invalid password",
			}); errJSONEncode != nil {
			log.Printf("LoginHandler Encode Message err: %s", errJSONEncode)
		}
		return
	}

	// Checking for the existence of a session and creating a session
	if session, errSession := server.MemServ.SessionRepo.GetByUserID(user.ID); errSession == nil {
		// If the session has already been created
		if errJSONEncode := json.NewEncoder(w).Encode(session); errJSONEncode != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("LoginHandler Encode session err: %s", errJSONEncode)
		}
		w.WriteHeader(http.StatusCreated)
		return
	}

	// Otherwise, create a new one
	token, errCreateJWTToken := GenerateToken(username, user.ID, []byte(server.KeyJWT))
	if errCreateJWTToken != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("LoginHandler CreateJWTToken err: %s", errCreateJWTToken)
	}

	session := models.Session{
		Token:  token,
		UserID: user.ID,
	}
	// Entry into the database with verification
	if errSessionRepoCreate := server.MemServ.SessionRepo.Create(&session); errSessionRepoCreate != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("LoginHandler SessionRepo Create err: %s", errSessionRepoCreate)
	}
	// Sending data
	if errJSONEncode := json.NewEncoder(w).Encode(session); errJSONEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("LoginHandler Encode session err: %s", errJSONEncode)
	}
	w.WriteHeader(http.StatusCreated)
}

func (server *Server) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Getting all posts from the database
	posts, err := server.MemServ.PostRepo.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GetPostsHandler PostRepo GetAll err: %s", err)
		return
	}
	// Sending data
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GetPostsHandler PostRepo Encode posts err: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (server *Server) PostPostsHandler(w http.ResponseWriter, r *http.Request) {
	var data PostData

	// Decoding the JSON from the request body
	errJSONDecode := json.NewDecoder(r.Body).Decode(&data)
	if errJSONDecode != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	category := data.Category
	text := data.Text
	url := data.URL
	title := data.Title
	typePost := data.Type

	genID, errGenID := GenerateID()
	if errGenID != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("PostPostsHandler GenerateID err: %s", errGenID)
	}

	token, errToken := getJWTByRequest(r)
	if errToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("PostPostsHandler getJWTByRequest err: %s", errToken)
	}
	user, errGetUserToken := getUserByJWT(token, []byte(server.KeyJWT))
	if errGetUserToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("PostPostsHandler getUserByJWT err: %s", errGetUserToken)
	}

	post := models.Post{
		ID:       genID,
		Score:    1,
		Views:    0,
		Type:     typePost,
		Title:    title,
		Author:   *user,
		Category: category,
		Text:     text,
		URL:      url,
		Votes: []models.Vote{
			{
				UserID: user.ID,
				Vote:   1,
			},
		},
		Comments:         []models.Comment{},
		Created:          time.Now(),
		UpvotePercentage: 100,
	}
	errPostRepoCreate := server.MemServ.PostRepo.Create(&post)
	if errPostRepoCreate != nil {
		log.Printf("PostPostsHandler PostRepo Create err: %s", errPostRepoCreate)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if errJSONEncode := json.NewEncoder(w).Encode(post); errJSONEncode != nil {
		log.Printf("PostPostsHandler PostRepo Encode post: %s", errJSONEncode)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (server *Server) GetPostsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	categoryName, ok := vars["CATEGORY_NAME"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	categoryPosts, errPostRepoGetByCategory := server.MemServ.PostRepo.GetByCategory(categoryName)
	if errPostRepoGetByCategory != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GetPostsByCategory GetByCategory categoryName %s", errPostRepoGetByCategory)
		return
	}

	if errJSONEncode := json.NewEncoder(w).Encode(categoryPosts); errJSONEncode != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GetPostsByCategory Encode categoryPost %s", errJSONEncode)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (server *Server) GetPostsByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	postID, ok := vars["POST_ID"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idPost, err := server.MemServ.PostRepo.GetByID(postID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GetPostsByID PostRepo GetByID %s", err)
		return
	}

	idPost.Views += 1
	if errUpdate := server.MemServ.PostRepo.Update(idPost); errUpdate != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GetPostsByID PostRepo.Update err:%s", err)
		return
	}

	if err := json.NewEncoder(w).Encode(idPost); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GetPostsByID Encode idPost %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (server *Server) AddCommentPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	postID, ok := vars["POST_ID"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var data CommentData

	// Decoding the JSON from the request body
	errDecode := json.NewDecoder(r.Body).Decode(&data)
	if errDecode != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	bodyText := data.Comment

	token, errToken := getJWTByRequest(r)
	if errToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("AddCommentPost getJWTByRequest err: %s", errToken)
	}
	user, errGetUserToken := getUserByJWT(token, []byte(server.KeyJWT))
	if errGetUserToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("AddCommentPost getUserByJWT err: %s", errGetUserToken)
	}

	idPost, errGetByID := server.MemServ.PostRepo.GetByID(postID)
	if errGetByID != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	genIDComment, errGenIDComment := GenerateID()
	if errGenIDComment != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	newComment := models.Comment{
		ID:      genIDComment,
		Author:  *user,
		Body:    bodyText,
		Created: time.Now(),
	}

	idPost.Comments = append(idPost.Comments, newComment)

	if err := server.MemServ.PostRepo.Update(idPost); err != nil {
		log.Printf("AddCommentPost UserRepo Update err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(idPost); err != nil {
		log.Printf("AddCommentPost Encode idPost err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (server *Server) DeleteCommentPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	postID, okPostID := vars["POST_ID"]
	if !okPostID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	commentID, okCommentID := vars["COMMENT_ID"]
	if !okCommentID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	post, err := server.MemServ.PostRepo.GetByID(postID)
	if err != nil {
		log.Printf("DeleteCommentPost PostRepo GetByID err: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for indexComment, comment := range post.Comments {
		if comment.ID == commentID {
			post.Comments = append(post.Comments[:indexComment], post.Comments[indexComment+1:]...)
			break
		}
	}

	if errUpdPost := server.MemServ.PostRepo.Update(post); errUpdPost != nil {
		log.Printf("DeleteCommentPost PostRepo Update err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(post); err != nil {
		log.Printf("DeleteCommentPost Encode post err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (server *Server) UpvotePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	postID, ok := vars["POST_ID"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, errToken := getJWTByRequest(r)
	if errToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("UpvotePost getJWTByRequest err: %s", errToken)
	}
	user, errGetUserToken := getUserByJWT(token, []byte(server.KeyJWT))
	if errGetUserToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("UpvotePost getUserByJWT err: %s", errGetUserToken)
	}

	post, err := server.MemServ.PostRepo.GetByID(postID)
	if err != nil {
		log.Printf("UpvotePost PostRepo GetByID err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Checking for ratings user.ID
	for index, vote := range post.Votes {
		if vote.UserID == user.ID {
			// If it's upvote, then we don't do anything and just delete the same post.
			if vote.Vote == 1 {
				if errMarshal := json.NewEncoder(w).Encode(post); errMarshal != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("UpvotePost NewEncoder Encode post If Vote == 1 err: %s", errMarshal)
					return
				}
				w.WriteHeader(http.StatusOK)
				// If the score is negative
			} else {
				post.Votes[index].Vote = 1
				post.Score += 2
				post.UpvotePercentage = post.Score / len(post.Votes) * 100
				if errUpate := server.MemServ.PostRepo.Update(post); errUpate != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("UpvotePost PostRepo Update post If Vote == -1 err: %s", errUpate)
					return
				}
				if errMarshal := json.NewEncoder(w).Encode(post); errMarshal != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("UpvotePost NewEncoder Encode post If Vote == -1 err: %s", errMarshal)
					return
				}
				w.WriteHeader(http.StatusOK)
			}
			return
		}
	}
	// The case when the estimate was not found
	newVote := models.Vote{
		UserID: user.ID,
		Vote:   1,
	}
	post.Votes = append(post.Votes, newVote)
	post.Score += 1
	post.UpvotePercentage = (post.Score + len(post.Votes)) / (len(post.Votes) * 2) * 100
	if errUpate := server.MemServ.PostRepo.Update(post); errUpate != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("UpvotePost PostRepo Update post Not Have Until err: %s", errUpate)
		return
	}
	if errMarshal := json.NewEncoder(w).Encode(post); errMarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("UpvotePost NewEncoder Encode post Not Have until err: %s", errMarshal)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (server *Server) DownvotePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	postID, ok := vars["POST_ID"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	post, err := server.MemServ.PostRepo.GetByID(postID)
	if err != nil {
		log.Printf("DownvotePost PostRepo GetByID err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, errToken := getJWTByRequest(r)
	if errToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("DownvotePost getJWTByRequest err: %s", errToken)
	}
	user, errGetUserToken := getUserByJWT(token, []byte(server.KeyJWT))
	if errGetUserToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("DownvotePost getUserByJWT err: %s", errGetUserToken)
	}

	// Checking for ratings user.ID
	for index, vote := range post.Votes {
		if vote.UserID == user.ID {
			// If it's a downvote, then we don't do anything and just delete the same post.
			if vote.Vote == -1 {
				if errMarshal := json.NewEncoder(w).Encode(post); errMarshal != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("DownvotePost NewEncoder Encode post If Vote == -1 err: %s", errMarshal)
					return
				}
				w.WriteHeader(http.StatusOK)
				// If the rating is positive
			} else {
				post.Votes[index].Vote = -1
				post.Score -= 2
				post.UpvotePercentage = post.Score / len(post.Votes) * 100
				if errUpate := server.MemServ.PostRepo.Update(post); errUpate != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("DownvotePost PostRepo Update post If Vote == 1 err: %s", errUpate)
					return
				}
				if errMarshal := json.NewEncoder(w).Encode(post); errMarshal != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("DownvotePost NewEncoder Encode post If Vote == 1 err: %s", errMarshal)
					return
				}
				w.WriteHeader(http.StatusOK)
			}
			return
		}
	}
	// The case when the estimate was not found
	newVote := models.Vote{
		UserID: user.ID,
		Vote:   -1,
	}
	post.Votes = append(post.Votes, newVote)
	post.Score -= 1
	post.UpvotePercentage = (post.Score + len(post.Votes)) / (len(post.Votes) * 2) * 100
	if errUpate := server.MemServ.PostRepo.Update(post); errUpate != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("DownvotePost PostRepo Update post Not Have Until err: %s", errUpate)
		return
	}
	if errMarshal := json.NewEncoder(w).Encode(post); errMarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("DownvotePost NewEncoder Encode post Not Have until err: %s", errMarshal)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (server *Server) UnvotePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	postID, ok := vars["POST_ID"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	post, err := server.MemServ.PostRepo.GetByID(postID)
	if err != nil {
		log.Printf("UnvotePost PostRepo GetByID err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, errToken := getJWTByRequest(r)
	if errToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("UnvotePost getJWTByRequest err: %s", errToken)
	}
	user, errGetUserToken := getUserByJWT(token, []byte(server.KeyJWT))
	if errGetUserToken != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("UnvotePost getUserByJWT err: %s", errGetUserToken)
	}
	// Checking for ratings user.ID
	for index, vote := range post.Votes {
		if vote.UserID == user.ID {
			if vote.Vote == 1 {
				post.Score -= 1
			} else {
				post.Score += 1
			}
			post.Votes = append(post.Votes[:index], post.Votes[index+1:]...)
			post.UpvotePercentage = (post.Score + len(post.Votes)) / (len(post.Votes) * 2) * 100
			if errUpate := server.MemServ.PostRepo.Update(post); errUpate != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("UnvotePost PostRepo Update Have Vote post err: %s", errUpate)
				return
			}
			if errMarshal := json.NewEncoder(w).Encode(post); errMarshal != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("UnvotePost NewEncoder Encode Have Vote post err: %s", errMarshal)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	// The case when the estimate was not found
	if errMarshal := json.NewEncoder(w).Encode(post); errMarshal != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("UnvotePost NewEncoder Encode post Not Have Vote err: %s", errMarshal)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (server *Server) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	postID, ok := vars["POST_ID"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := server.MemServ.PostRepo.Delete(postID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("DeletePost PostRepo Delete postID err: %s", err)
		return
	}

	if err := json.NewEncoder(w).Encode(
		struct {
			Message string `json:"message"`
		}{
			Message: "success",
		}); err != nil {
		log.Printf("DeletePost Encode Message err: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (server *Server) GetPostsByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	userLogin, ok := vars["USER_LOGIN"]

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := server.MemServ.UserRepo.GetByUsername(userLogin)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("GetPostsByUser UserRepo GetByUsername userLogin err: %s", err)

		return
	}

	userPosts, err := server.MemServ.PostRepo.GetByUserID(user.ID)

	if errors.Is(err, repository.ErrNoPostsUser) {
		userPosts = make([]models.Post, 0)
	} else if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("GetPostsByUser PostRepo GetByUserID user.ID err: %s", err)

		return
	}

	if err := json.NewEncoder(w).Encode(userPosts); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GetPostsByUser Encode userPosts %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
