package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/gorilla/handlers"
	"github.com/jmlattanzi/itaic-backend/itaic/cc"
	"github.com/jmlattanzi/itaic-backend/itaic/pc"
	"github.com/jmlattanzi/itaic-backend/itaic/uc"
	"github.com/streadway/amqp"
	"goji.io"
	"goji.io/pat"
	"google.golang.org/api/option"
)

func main() {
	fmt.Println("[ * ] Starting API....")
	ctx := context.Background()

	// Use a service account
	sa := option.WithCredentialsFile("itaic-key.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	auth, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("[ ! ] Error getting Auth client: %v\n", err)
	}

	conn, err := amqp.Dial("amqp://176.24.0.9:5672")
	if err != nil {
		log.Fatal("[ ! ] Failed connecting to RabbitMQ: ", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("[ ! ] Error opening a channel: ", err)
	}

	q, err := ch.QueueDeclare(
		"test", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		log.Fatal("[ ! ] Error declaring a queue: ", err)
	}

	router := goji.NewMux()

	// post routes
	router.HandleFunc(pat.Get("/posts"), pc.HandleGetPosts(ctx, client))
	router.HandleFunc(pat.Post("/posts"), pc.HandleCreatePost(ctx, client, ch, q))
	router.HandleFunc(pat.Get("/posts/:id"), pc.HandleGetPostByID(ctx, client))
	router.HandleFunc(pat.Put("/posts/:id"), pc.HandleEditPost(ctx, client, ch, q))
	router.HandleFunc(pat.Delete("/posts/:id/:uid"), pc.HandleDeletePost(ctx, client))
	router.HandleFunc(pat.Put("/posts/like/:id/:uid"), pc.HandleLikePost(ctx, client, ch, q))

	// comment routes
	router.HandleFunc(pat.Post("/comment/:id"), cc.HandleAddComment(ctx, client))
	router.HandleFunc(pat.Delete("/comment/:id/:comment"), cc.HandleDeleteComment(ctx, client))
	router.HandleFunc(pat.Put("/comment/:id/:comment"), cc.HandleEditComment(ctx, client))
	router.HandleFunc(pat.Put("/comment/like/:post_id/:id/:uid"), cc.HandleLikeComment(ctx, client))

	// user routes
	router.HandleFunc(pat.Get("/user/:uid"), uc.HandleGetUser(ctx, client))
	router.HandleFunc(pat.Post("/user"), uc.HandleRegisterUser(ctx, client, auth))
	router.HandleFunc(pat.Put("/user/:uid"), uc.HandleEditUser(ctx, client))

	// MQProducer()
	fmt.Println("[ + ] API Started")
	http.ListenAndServe(":8000", handlers.LoggingHandler(os.Stdout, router))
}
