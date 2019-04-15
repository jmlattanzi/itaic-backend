package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"

	"github.com/go-redis/redis"
	"github.com/jmlattanzi/itaic-backend/itaic-cache/models"
	"github.com/streadway/amqp"

	"goji.io"
	"goji.io/pat"
)

func main() {
	fmt.Println("[ * ] Starting cache API....")
	client := redis.NewClient(&redis.Options{
		Addr:     "176.24.0.13:6379",
		Password: "",
		DB:       0,
	})

	defer client.Close()
	fmt.Println("[ * ] Initializing cache")
	res, err := http.Get("http://176.24.0.3:8000/posts")
	if err != nil {
		fmt.Println("[ ! ] Error making request to db api: ", err)
	}

	posts := []models.Post{}
	err = json.NewDecoder(res.Body).Decode(&posts)
	if err != nil {
		fmt.Println("[ ! ] Error decoding response body")
	}

	for _, post := range posts {
		mp, err := json.Marshal(post)
		if err != nil {
			fmt.Println("[ ! ] Error marshaling post: ", err)
		}

		_ = client.HSet("posts", post.ID, mp)
	}

	router := goji.NewMux()
	router.HandleFunc(pat.Get("/posts"), HandleGetAllPosts(client))
	router.HandleFunc(pat.Get("/posts/:id"), HandleGetPostByID(client))

	MQConsumer(client)
	http.ListenAndServe(":5000", handlers.LoggingHandler(os.Stdout, router))
}

// MQConsumer ... checks the queue for a message and prints it
func MQConsumer(client *redis.Client) {
	conn, err := amqp.Dial("amqp://176.24.0.9:5672")
	if err != nil {
		log.Fatal("[ ! ] Error connecting to RabbitMQ: ", err)
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

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("[ ! ] Error registering consumer: ", err)
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			fmt.Println("[ m ] Message Type: ", d.Type)
			fmt.Println("[ m ] Message received: ", string(d.Body))

			// check for updates
			if d.Type == "UPDATE" {
				fmt.Println("[ ! ] Need to update cache")
				fmt.Println("[ ! ] ID of post: ", string(d.Body))
				go UpdateCache(string(d.Body), client)
			}
		}
	}()

	fmt.Println("[ * ] Waiting to recieve messages")
	<-forever
}

// HandleGetAllPosts ... Test route
func HandleGetAllPosts(client *redis.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")

		var posts []string
		result := client.HGetAll("posts")
		for _, post := range result.Val() {
			posts = append(posts, post)
		}

		parsedResults := []models.Post{}
		for _, post := range posts {
			parsed := models.Post{}
			json.Unmarshal([]byte(post), &parsed)
			parsedResults = append(parsedResults, parsed)
		}

		json.NewEncoder(res).Encode(parsedResults)
	}
}

// HandleGetPostByID ... Handles getting a specific post
func HandleGetPostByID(client *redis.Client) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "application/json")
		id := pat.Param(req, "id")

		result := client.HGet("posts", id)
		fmt.Println(result)
		if result.Err() == redis.Nil {
			fmt.Println("[ ! ] Post not found: ", result)
		} else {
			parsed := models.Post{}
			err := json.Unmarshal([]byte(result.Val()), &parsed)
			if err != nil {
				fmt.Println("[ ! ] Error unmarshaling post: ", err)
			}
			json.NewEncoder(res).Encode(parsed)
		}
	}
}

// UpdateCache ... if the post isn't found in the cache this function will check the db
func UpdateCache(id string, client *redis.Client) {
	fmt.Println("[ - ] Checking DB for post with id: " + id)
	// for now we'll just simulate a miss
	result, _ := http.Get("http://176.24.0.3:8000/posts/" + id)
	if result.StatusCode == 500 {
		fmt.Println("Received error")
	}

	post := models.Post{}
	err := json.NewDecoder(result.Body).Decode(&post)
	if err != nil {
		fmt.Println("[ ! ] Error decoding result body: ", err)
	}

	mp, err := json.Marshal(post)
	if err != nil {
		fmt.Println("[ ! ] Error marshaling post: ", err)
	}

	_ = client.HSet("posts", post.ID, mp)
	fmt.Println("[ + ] Cache updated")
}
