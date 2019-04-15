package pc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"

	"cloud.google.com/go/firestore"
	"goji.io/pat"
	"google.golang.org/api/iterator"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fatih/structs"
	"github.com/jmlattanzi/itaic-backend/itaic/config"
	"github.com/jmlattanzi/itaic-backend/itaic/models"
)

// HandleGetPosts ... Gets all posts from the DB
func HandleGetPosts(ctx context.Context, client *firestore.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		posts := []models.Post{}
		iter := client.Collection("posts").Documents(ctx)
		for {
			post := models.Post{}
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				log.Fatal("[ ! ] Error iterating document: ", err)
			}

			err = doc.DataTo(&post)
			if err != nil {
				log.Fatal("[ ! ] Error mapping data to struct: ", err)
			}
			posts = append(posts, post)
		}

		json.NewEncoder(res).Encode(&posts)
	}
}

// HandleGetPostByID ... Gets a single post based on id
func HandleGetPostByID(ctx context.Context, client *firestore.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		post := models.Post{}
		id := pat.Param(req, "id")
		doc, err := client.Collection("posts").Doc(id).Get(ctx)

		if err != nil {
			fmt.Println("[ ! ] Document get returned an err: ", err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("500 - post not found"))
			return
		}

		err = doc.DataTo(&post)
		if err != nil {
			fmt.Println("[ ! ] Error mapping data into struct: ", err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("500 - post not found"))
			return
		}

		json.NewEncoder(res).Encode(&post)
	}
}

//HandleCreatePost ...Inserts a post to the DB
func HandleCreatePost(ctx context.Context, client *firestore.Client, ch *amqp.Channel, q amqp.Queue) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		// setup the new post
		newPost := models.Post{}
		user := models.User{}
		caption := req.FormValue("caption")
		uid := req.FormValue("uid")
		imageLocation := upload(req)

		// create a new document in the collection
		doc := client.Collection("posts").NewDoc()

		// using the new doc, set the id in the post to the doc's id
		newPost.ID = doc.ID
		newPost.UID = uid
		newPost.Caption = caption
		newPost.Created = time.Now().String()
		newPost.ImageURL = imageLocation

		query := client.Collection("users").Where("uid", "==", uid)
		iter := query.Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				fmt.Println("[ ! ] Error iterating documents: ", err)
				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("500 - post not found"))
				return
			}

			err = doc.DataTo(&user)
			if err != nil {
				fmt.Println("[ ! ] Error mapping data to struct: ", err)
				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("500 - post not found"))
				return
			}
		}
		user.Posts = append(user.Posts, doc.ID)
		newPost.Username = user.Username

		// write data to the doc
		_, err := doc.Create(ctx, newPost)
		if err != nil {
			fmt.Println("[ ! ] Error creating new document: ", err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("500 - post not found"))
			return
		}

		_, err = client.Collection("users").Doc(user.ID).Set(ctx, user)
		if err != nil {
			fmt.Println("[ ! ] Error assigning post to user: ", err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("500 - post not found"))
			return
		}

		// send message saying a post was updated
		body := doc.ID
		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Type:        "UPDATE",
				Body:        []byte(body),
			})
		if err != nil {
			log.Fatal("[ ! ] Error publishing message")
		}
		fmt.Println("[ + ] Message sent")
		json.NewEncoder(res).Encode(&newPost)
	}
}

// HandleDeletePost ...Deletes a document form the DB
func HandleDeletePost(ctx context.Context, client *firestore.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		id := pat.Param(req, "id")
		uid := pat.Param(req, "uid")
		user := models.User{}
		_, err := client.Collection("posts").Doc(string(id)).Delete(ctx)
		if err != nil {
			log.Fatal("[ ! ] Error deleting post: ", err)
		}

		query := client.Collection("users").Where("uid", "==", uid)
		iter := query.Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				fmt.Println("[ ! ] Error iterating documents: ", err)
				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("500 - post not found"))
				return
			}

			err = doc.DataTo(&user)
			if err != nil {
				fmt.Println("[ ! ] Error mapping data to struct: ", err)
				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("500 - post not found"))
				return
			}
		}

		for i := 0; i < len(user.Posts); i++ {
			if user.Posts[i] == id {
				user.Posts = append(user.Posts[:i], user.Posts[i+1:]...)
				i--
				break
			}
		}

		_, err = client.Collection("users").Doc(user.ID).Set(ctx, user)
		if err != nil {
			log.Fatal("[ ! ] Error setting user: ", err)
		}

		json.NewEncoder(res).Encode("Post deleted")
	}
}

// HandleEditPost ...Edits a post in the DB
func HandleEditPost(ctx context.Context, client *firestore.Client, ch *amqp.Channel, q amqp.Queue) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		type Caption struct {
			Caption string `json:"caption"`
		}

		// setup some variable
		var newCaption Caption
		id := pat.Param(req, "id")
		currentPost := models.Post{}

		// get the document
		doc, err := client.Collection("posts").Doc(id).Get(ctx)
		if err != nil {
			log.Fatal("[ ! ] Error getting document: ", err)
		}

		// populate the current post
		err = doc.DataTo(&currentPost)
		if err != nil {
			log.Fatal("[ ! ] Error mapping data to the struct: ", err)
		}

		// decode the new caption string
		err = json.NewDecoder(req.Body).Decode(&newCaption)
		if err != nil {
			log.Fatal("[ ! ] Error decoding request body")
		}
		currentPost.Caption = newCaption.Caption

		// this is why I have to do in such a long manner
		mappedPost := structs.Map(currentPost)
		_, err = client.Collection("posts").Doc(id).Set(ctx, mappedPost)
		if err != nil {
			log.Fatal("[ ! ] Error setting document: ", err)
		}

		body := id
		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Type:        "UPDATE",
				Body:        []byte(body),
			})
		if err != nil {
			log.Fatal("[ ! ] Error publishing message")
		}
		fmt.Println("[ + ] Message sent")

		json.NewEncoder(res).Encode(currentPost)
	}
}

// HandleLikePost ... Handles liking a post
func HandleLikePost(ctx context.Context, client *firestore.Client, ch *amqp.Channel, q amqp.Queue) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		id := pat.Param(req, "id")
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

		doc, err := client.Collection("posts").Doc(id).Get(ctx)
		if err != nil {
			log.Fatal("[ ! ] Error finding post")
		}

		err = doc.DataTo(&post)
		if err != nil {
			log.Fatal("[ ! ] Error mapping data to struct: ", err)
		}

		likes := user.Likes
		addToLikes, i := remove(likes, id)
		if addToLikes == false && i == 0 {
			likes = append(likes, id)
			post.Likes++
		} else {
			likes = append(likes[:i], likes[i+1:]...)
			post.Likes--
		}

		user.Likes = likes
		_, err = client.Collection("posts").Doc(id).Set(ctx, post)
		if err != nil {
			log.Fatal("[ ! ] Error setting post: ", err)
		}

		_, err = client.Collection("users").Doc(user.ID).Set(ctx, user)
		if err != nil {
			log.Fatal("[ ! ] Error setting user: ", err)
		}

		// TODO:
		// 	+ this only sends a message about the post being updated
		// 		if the cache needs to update the user as well I'll need to fix this
		body := id
		err = ch.Publish(
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Type:        "UPDATE",
				Body:        []byte(body),
			})
		if err != nil {
			log.Fatal("[ ! ] Error publishing message")
		}
		fmt.Println("[ + ] Message sent")

		json.NewEncoder(res).Encode(&post)
	}
}

func upload(r *http.Request) string {
	config := config.LoadConfigurationFile("config.json")

	creds := credentials.NewStaticCredentials(config.S3AccessKey, config.S3SecretAccessKey, "")
	sesh := session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String("us-west-1"),
	}))
	uploader := s3manager.NewUploader(sesh)

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Fatal("[!] Error in FormFile: ", err)
	}
	defer file.Close()
	fmt.Println("[>] Filename: ", header.Filename)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.S3Bucket),
		Key:    aws.String(header.Filename),
		Body:   file,
	})

	if err != nil {
		log.Fatal("[!] Error uploading file: ", err)
	}

	fmt.Println("[+] File uploaded")
	fmt.Println("[+] File URL: ", result.Location)
	return result.Location
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
