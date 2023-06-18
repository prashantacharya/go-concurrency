package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Post struct {
	UserId   int       `json:"userId"`
	Id       int       `json:"id"`
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	Comments []Comment `json:"comments" bson:"comments"`
}

type Comment struct {
	PostId int    `json:"postId"`
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Body   string `json:"body"`
}

type PostWithComments struct {
	Post
	Comments []Comment `json:"comments" bson:"comments"`
}

type CommentResponseChannel struct {
	PostId   *int
	Comments *[]Comment
	Err      error
}

func fetchPosts() ([]Post, error) {
	req, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	posts := []Post{}
	err = json.Unmarshal(body, &posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func returnValueOrErr(userId string) (*string, error) {
	if userId == "1" {
		return nil, fmt.Errorf("User %s not found", userId)
	}

	return &userId, nil
}

func fetchComments(postId int) ([]Comment, error) {
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/posts/%d/comments", postId)
	req, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	comments := []Comment{}
	err = json.Unmarshal(body, &comments)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func fetchCommentsAsync(postId int, c chan CommentResponseChannel) {
	comments, err := fetchComments(postId)
	c <- CommentResponseChannel{&postId, &comments, err}
}

func createPostWithCommentMap(posts []Post) map[int]PostWithComments {
	postWithCommentsMap := make(map[int]PostWithComments)
	for _, post := range posts {
		postWithComments := PostWithComments{Post: post, Comments: []Comment{}}
		postWithCommentsMap[post.Id] = postWithComments
	}

	return postWithCommentsMap
}

func main() {
	t := time.Now()
	posts, err := fetchPosts()

	if err != nil {
		panic(err)
	}

	c := make(chan CommentResponseChannel)
	postWithComments := createPostWithCommentMap(posts)

	var wg sync.WaitGroup
	for _, post := range posts {
		wg.Add(1)
		go func(p Post) {
			defer wg.Done()
			fetchCommentsAsync(p.Id, c)
		}(post)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	for commentResponse := range c {
		postID := *commentResponse.PostId
		post := postWithComments[postID]
		post.Comments = *commentResponse.Comments
		postWithComments[postID] = post
	}

	for _, post := range postWithComments {
		fmt.Println(post.Id, post.Title)
		fmt.Println("===========================")
		fmt.Println("Comments:")
		for _, comment := range post.Comments {
			fmt.Println("\t", comment.Id, comment.Body)
		}
	}

	fmt.Println("Time taken:", time.Since(t))
}
