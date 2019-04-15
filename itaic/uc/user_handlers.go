package uc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/jmlattanzi/itaic-backend/itaic/models"
	"goji.io/pat"
	"google.golang.org/api/iterator"
)

// HandleGetUser ... Gets a single user based on uid
func HandleGetUser(ctx context.Context, client *firestore.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		uid := pat.Param(req, "uid")
		user := models.User{}
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

			// fmt.Println(doc.Data())
			err = doc.DataTo(&user)
			if err != nil {
				log.Fatal("[ ! ] Error mapping data to struct: ", err)
			}
		}

		json.NewEncoder(res).Encode(user)
	}
}

// HandleRegisterUser ... Handles registering a user to the auth system and adding them to the db
func HandleRegisterUser(ctx context.Context, client *firestore.Client, authClient *auth.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		// register user
		// add to the db
		newUser := models.User{}
		err := json.NewDecoder(req.Body).Decode(&newUser)
		if err != nil {
			log.Fatal("[ ! ] Error decoding the request body: ", err)
		}

		params := (&auth.UserToCreate{}).Email(newUser.Email).DisplayName(newUser.Username)

		user, err := authClient.CreateUser(ctx, params)
		if err != nil {
			log.Fatal("[ ! ] Error creating user: ", err)
		}

		fmt.Println("[ + ] User created")
		doc := client.Collection("users").NewDoc()
		newUser.UID = user.UID
		newUser.ID = doc.ID

		_, err = doc.Create(ctx, newUser)
		if err != nil {
			log.Fatal("[ ! ] Error adding document to users collection: ", err)
		}

		json.NewEncoder(res).Encode(&newUser)
	}
}

// HandleEditUser ... Handles editing the user's bio
func HandleEditUser(ctx context.Context, client *firestore.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		type NewBio struct {
			Bio string `json:"bio"`
		}

		uid := pat.Param(req, "uid")
		user := models.User{}
		newBio := NewBio{}

		err := json.NewDecoder(req.Body).Decode(&newBio)
		if err != nil {
			log.Fatal("[ ! ] Error decoding request body: ", err)
		}
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

		user.Bio = newBio.Bio
		_, err = client.Collection("users").Doc(user.ID).Set(ctx, user)
		if err != nil {
			log.Fatal("[ ! ] Error setting document: ", err)
		}

		json.NewEncoder(res).Encode(&user)
	}
}
