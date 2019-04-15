package models

// post json
// {
//     "id": "369",
//     "uid": "test",
//     "username": "new post",
//     "caption": "new post",
//     "image": "test",
//     "likes": 0,
//     "created": "test",
//     "comments": [
//         {
//             "id": "test",
//             "uid": "test",
//             "comment": "test",
//             "time": "test",
//             "likes": 0
//         }
//     ]
// }

// user json
// {
// 	"uid": "test",
// 	"id": "test",
// 	"username": "test",
// 	"email": "test@gmail.com",
// 	"bio": "test",
//  "profile_pic": "nice",
// 	"posts": ["test"],
// 	"likes": ["test"],
//	"comment_likes": ["test"]
// 	"following": ["test"],
// 	"followers": ["test"]
// }

// Comment ... Defines the structure of a comment in the post
type Comment struct {
	ID       string `firestore:"id"`
	UID      string `firestore:"uid"`
	Comment  string `firestore:"comment"`
	Created  string `firestore:"created"`
	Likes    int    `firestore:"likes"`
	Username string `firestore:"username"`
}

// Post ... Defines the structure of our post in firestore
type Post struct {
	ID       string    `firestore:"id"`
	UID      string    `firestore:"uid"`
	Username string    `firestore:"username"`
	Caption  string    `firestore:"caption"`
	ImageURL string    `firestore:"imageURL"`
	Likes    int       `firestore:"likes"`
	Created  string    `firestore:"created"`
	Comments []Comment `firestore:"comments"`
}

// User ... Defines what will be stored in the user object
type User struct {
	UID          string   `firestore:"uid"`
	ID           string   `firestore:"id"`
	Username     string   `firestore:"username"`
	Email        string   `firestore:"email"`
	Bio          string   `firestore:"bio"`
	ProfilePic   string   `firestore:"profile_pic"`
	Posts        []string `firestore:"posts"`
	Likes        []string `firestore:"likes"`
	CommentLikes []string `firestore:"comment_likes"`
	Following    []string `firestore:"following"`
	Followers    []string `firestore:"followers"`
}

// Config ... Defines the shape of our config
type Config struct {
	S3AccessKey       string `json:"S3_ACCESS_KEY"`
	S3SecretAccessKey string `json:"S3_SECRET_ACCESS_KEY"`
	S3Bucket          string `json:"S3_BUCKET"`
}
