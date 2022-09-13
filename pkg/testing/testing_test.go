package testing

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func TestWunderBase(t *testing.T) {
	if os.Getenv("SKIP_BUILD") != "true" {
		cmd := exec.Command("docker", "build", "--platform", "linux/amd64", "-t", "wundergraph/wunderbase", ".")
		workingDir, err := os.Getwd()
		assert.NoErrorf(t, err, "get current path")
		cmd.Dir = path.Clean(path.Join(workingDir, "../../"))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		assert.NoErrorf(t, err, "docker build wunderbase")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "run", "-p", "4466:4466", "-e", "SLEEP_AFTER_SECONDS=1", "wundergraph/wunderbase")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	assert.NoErrorf(t, err, "docker run wunderbase")

	baseURL := "http://localhost:4466"
	waitUntilReady(t, baseURL, 30*time.Second)

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL: baseURL,
		Client: &http.Client{
			Jar:     httpexpect.NewJar(),
			Timeout: time.Second * 2,
		},
		Reporter: httpexpect.NewRequireReporter(t),
	})

	e.GET(baseURL).Expect().Status(http.StatusOK).Body().Contains("GraphQL Contributors").Contains(baseURL)

	createUser := e.POST(baseURL).
		WithHeader("Content-Type", "application/json").
		WithJSON(map[string]interface{}{
			"query":         createUser,
			"operationName": "CreateUser",
			"variables":     map[string]interface{}{},
		}).
		Expect().
		Status(http.StatusOK).Body().Raw()
	assert.Equal(t, `{"data":{"createOneUser":{"id":1,"email":"jens@wundergraph.com","name":"Jens"}}}`, createUser)

	allUsers := e.POST(baseURL).
		WithHeader("Content-Type", "application/json").
		WithJSON(map[string]interface{}{
			"query":         allUsers,
			"operationName": "AllUsers",
			"variables":     map[string]interface{}{},
		}).
		Expect().
		Status(http.StatusOK).Body().Raw()
	assert.Equal(t, `{"data":{"findManyUser":[{"id":1,"email":"jens@wundergraph.com"}]}}`, allUsers)

	createdPost := e.POST(baseURL).
		WithHeader("Content-Type", "application/json").
		WithJSON(map[string]interface{}{
			"query":         createPost,
			"operationName": "CreatePost",
			"variables":     map[string]interface{}{},
		}).
		Expect().
		Status(http.StatusOK).Body().Raw()
	assert.Equal(t, `{"data":{"createOnePost":{"id":1,"title":"myPost","author":{"id":1,"email":"jens@wundergraph.com","name":"Jens"}}}}`, createdPost)

	createdSecondPost := e.POST(baseURL).
		WithHeader("Content-Type", "application/json").
		WithJSON(map[string]interface{}{
			"query":         createPost,
			"operationName": "CreatePost",
			"variables":     map[string]interface{}{},
		}).
		Expect().
		Status(http.StatusOK).Body().Raw()
	assert.Equal(t, `{"data":{"createOnePost":{"id":2,"title":"myPost","author":{"id":1,"email":"jens@wundergraph.com","name":"Jens"}}}}`, createdSecondPost)

	allPosts := e.POST(baseURL).
		WithHeader("Content-Type", "application/json").
		WithJSON(map[string]interface{}{
			"query":         allPosts,
			"operationName": "AllPosts",
			"variables":     map[string]interface{}{},
		}).
		Expect().
		Status(http.StatusOK).Body().Raw()
	assert.Equal(t, `{"data":{"findManyPost":[{"id":1,"title":"myPost","author":{"id":1,"email":"jens@wundergraph.com","name":"Jens"}},{"id":2,"title":"myPost","author":{"id":1,"email":"jens@wundergraph.com","name":"Jens"}}]}}`, allPosts)
}

const (
	createUser = `mutation CreateUser {
  createOneUser(data: {name: "Jens" email: "jens@wundergraph.com"}){
    id
    email
    name
  }
}`
	allUsers = `query AllUsers {
  findManyUser {
    id
    email
  }
}`
	createPost = `mutation CreatePost {
  createOnePost(data: {title: "myPost" author: {connect: {email: "jens@wundergraph.com"}}}){
    id
    title
    author {
      id
      email
      name
    }
  }
}`
	allPosts = `query AllPosts {
  findManyPost {
    id
    title
    author {
      id
      email
      name
    }
  }
}`
)

func waitUntilReady(t *testing.T, url string, duration time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			assert.Fail(t, "timeout waiting for wunderbase to be ready")
			return
		default:
			_, err := http.Get(url)
			if err == nil {
				return
			}
			time.Sleep(time.Second)
		}
	}
}
