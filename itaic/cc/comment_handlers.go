package cc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	shortid "github.com/jasonsoft/go-short-id"
	"google.golang.org/api/iterator"

	"github.com/jmlattanzi/itaic-backend/itaic/models"
	"goji.io/pat"

	"cloud.google.com/go/firestore"
)

// HandleAddComment ... Adds a comment to the db
func HandleAddComment(ctx context.Context, client *firestore.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		id := pat.Param(req, "id")
		user := models.User{}
		type NewComment struct {
			Comment string `json:"comment"`
		}

		opt := shortid.Options{
			Number:        14,
			StartWithYear: false,
			EndWithHost:   false,
		}

		newComment := models.Comment{
			ID:       shortid.Generate(opt),
			Created:  time.Now().String(),
			Likes:    0,
			Comment:  "",
			UID:      "",
			Username: "",
		}
		currentPost := models.Post{}
		err := json.NewDecoder(req.Body).Decode(&newComment)
		if err != nil {
			log.Fatal("[ ! ] Error decoding request body: ", err)
		}

		query := client.Collection("users").Where("uid", "==", newComment.UID)
		iter := query.Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				log.Fatal("[ ! ] Error iterating documents: ", err)
			}

			fmt.Println("doc.Data: ", doc.Data())

			err = doc.DataTo(&user)
			if err != nil {
				log.Fatal("[ ! ] Error mapping data to struct: ", err)
			}
		}

		fmt.Println("username: ", user)
		newComment.Username = user.Username

		doc, err := client.Collection("posts").Doc(id).Get(ctx)
		if err != nil {
			log.Fatal("[ ! ] Error getting document: ", err)
		}
		err = doc.DataTo(&currentPost)
		if err != nil {
			log.Fatal("[ ! ] Error writing data to struct: ", err)
		}

		currentPost.Comments = append(currentPost.Comments, newComment)
		_, err = client.Collection("posts").Doc(id).Set(ctx, currentPost)
		if err != nil {
			log.Fatal("[ ! ] Error setting document: ", err)
		}

		json.NewEncoder(res).Encode(currentPost)
	}
}

// HandleDeleteComment ... Deletes a comment based on post id and comment id
func HandleDeleteComment(ctx context.Context, client *firestore.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		id := pat.Param(req, "id")
		commentID := pat.Param(req, "comment")
		currentPost := models.Post{}
		comments := []models.Comment{}

		doc, err := client.Collection("posts").Doc(id).Get(ctx)
		if err != nil {
			log.Fatal("[ ! ] Error getting document")
		}

		err = doc.DataTo(&currentPost)
		if err != nil {
			log.Fatal("[ ! ] Error mapping data into struct: ", err)
		}

		for _, comment := range currentPost.Comments {
			if comment.ID == commentID {
				fmt.Println("[ + ] Comment found")
			} else {
				comments = append(comments, comment)
			}
		}

		currentPost.Comments = comments
		_, err = client.Collection("posts").Doc(id).Set(ctx, currentPost)
		if err != nil {
			log.Fatal("[ ! ] Error setting document: ", err)
		}
		json.NewEncoder(res).Encode(currentPost)
	}
}

// HandleEditComment ... Edits a comment and submits to the db
func HandleEditComment(ctx context.Context, client *firestore.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		id := pat.Param(req, "id")
		commentID := pat.Param(req, "comment")
		currentPost := models.Post{}
		comments := []models.Comment{}

		type NewComment struct {
			Comment string `json:"comment"`
		}
		newComment := NewComment{}

		err := json.NewDecoder(req.Body).Decode(&newComment)
		if err != nil {
			log.Fatal("[ ! ] Error decoding the response body: ", err)
		}

		doc, err := client.Collection("posts").Doc(id).Get(ctx)
		if err != nil {
			log.Fatal("[ ! ] Error getting document")
		}

		err = doc.DataTo(&currentPost)
		if err != nil {
			log.Fatal("[ ! ] Error mapping data into struct: ", err)
		}

		for _, comment := range currentPost.Comments {
			if comment.ID == commentID {
				fmt.Println("[ + ] Comment found")
				comment.Comment = newComment.Comment
			}

			comments = append(comments, comment)
		}

		currentPost.Comments = comments
		_, err = client.Collection("posts").Doc(id).Set(ctx, currentPost)
		if err != nil {
			log.Fatal("[ ! ] Error setting document: ", err)
		}
		json.NewEncoder(res).Encode(currentPost)
	}
}

//HandleLikeComment ... Handles liking a comment
func HandleLikeComment(ctx context.Context, client *firestore.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		id := pat.Param(req, "id")
		postID := pat.Param(req, "post_id")
		uid := pat.Param(req, "uid")
		user := models.User{}
		post := models.Post{}

		query := client.Collection("users").Where("uid", "==", uid)
		iter := query.Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				log.Fatal("[ ! ] Error iterating documents: ", err)
			}

			err = doc.DataTo(&user)
			if err != nil {
				log.Fatal("[ ! ] Error mapping data to struct: ", err)
			}
		}

		// get the post containing the comment
		doc, err := client.Collection("posts").Doc(postID).Get(ctx)
		if err != nil {
			log.Fatal("[ ! ] Error getting post: ", err)
		}

		err = doc.DataTo(&post)
		if err != nil {
			log.Fatal("[ ! ] Error mapping data into post: ", err)
		}

		likes := user.CommentLikes
		for index, comment := range post.Comments {
			if comment.ID == id {
				addToLikes, i := remove(likes, id)
				if addToLikes == false && i == 0 {
					likes = append(likes, id)
					post.Comments[index].Likes++
				} else {
					likes = append(likes[:i], likes[i+1:]...)
					post.Comments[index].Likes--
				}
			}
		}

		user.CommentLikes = likes
		_, err = client.Collection("users").Doc(user.ID).Set(ctx, user)
		if err != nil {
			log.Fatal("[ ! ] Error adding comment to liked comments: ", err)
		}

		_, err = client.Collection("posts").Doc(postID).Set(ctx, post)
		if err != nil {
			log.Fatal("[ ! ] Error updating post: ", err)
		}

		json.NewEncoder(res).Encode(&post)

		// find the comment
		// check if the user has already liked itaic
		// if yes, remove the comment from their likes and decrement comment likes
		// if no, add comment to their likes and increment comment like
	}
}

func remove(likes []string, id string) (bool, int) {
	for i := 0; i < len(likes); i++ {
		likeID := likes[i]
		if likeID == id {
			return true, i
		}
	}
	return false, 0
}
