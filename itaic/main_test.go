package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	firebase "firebase.google.com/go"
	"github.com/jmlattanzi/itaic-backend/itaic/pc"
	"github.com/stretchr/testify/assert"
	"goji.io"
	"goji.io/pat"
	"google.golang.org/api/option"
)

func Router() *goji.Mux {
	ctx := context.Background()
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

	router := goji.NewMux()
	router.HandleFunc(pat.Get("/posts"), pc.HandleGetPosts(ctx, client))
	return router
}

func TestHandleGetPosts(t *testing.T) {
	fmt.Println("[ t ] Testing HandleGetPosts....")
	req, _ := http.NewRequest("GET", "/posts", nil)
	res := httptest.NewRecorder()

	Router().ServeHTTP(res, req)
	assert.Equal(t, 200, res.Code, "OK response expected")
}
