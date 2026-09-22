# Forum Audit Guide

## Authentication

### Are an email and a password asked for in the registration?
Browser → `http://localhost:8080/auth`

Code → `internal/handlers/auth.go` → `Register()`

### Does the project detect if the email or password are wrong?
Browser → Try to login with wrong credentials at `http://localhost:8080/auth`

Code → `internal/handlers/auth.go` → `Login()`, `findUser()`

Code → `internal/auth/password.go` → `CheckPassword()`

### Does the project detect if the email or user name is already taken in the registration?
Browser → Try to register twice with the same email or username.

Code → `internal/handlers/auth.go` → `Register()`, `isDuplicateUserError()`

Code → `internal/database/users.go` → `CreateUser()`

### Try to register as a new user in the forum. Is it possible to register?
Browser → `http://localhost:8080/auth` → Register a new user.

Code → `internal/handlers/auth.go` → `Register()`

### Try to login with the user you created. Can you login and have all the rights of a registered user?
Browser → `http://localhost:8080/auth` → Login with the new user.

Code → `internal/handlers/auth.go` → `Login()`

### Try to login without any credentials. Does it show a warning message?
Browser → `http://localhost:8080/auth` → Submit login with empty credentials.

Code → `internal/handlers/auth.go` → `Login()`

### Are sessions present in the project?
Code → `internal/session/manager.go` → `Create()`, `User()`, `Destroy()`

Code → `internal/database/sessions.go` → `CreateSession()`, `GetSessionByID()`, `GetSessionByUserID()`, `DeleteSessionByID()`, `DeleteSessionByUserID()`

### Open two browsers, login in one and refresh the other. Does the non-logged browser remain unregistered?
Browser → Open two different browsers. Login in only one and refresh both.

Code → `internal/session/manager.go` → `User()`

Code → `internal/middleware/auth.go` → `LoadUser()`

### Open two browsers, login in both and refresh both. Does only one have an active session?
Browser → Login with the same account in two different browsers and refresh both.

Code → `internal/session/manager.go` → `Create()`

Code → `internal/database/sessions.go` → `GetSessionByUserID()`, `DeleteSessionByUserID()`, `CreateSession()`

### Open two browsers, login in one, create a post or comment and refresh both. Does it appear on both?
Browser → Login in one browser → create a post/comment → refresh the second browser.

Code → `internal/handlers/posts.go` → `Create()`

Code → `internal/handlers/comments.go` → `Create()`

---

## SQLite

### Does the code contain at least one CREATE query?
Code → `internal/database/schema/schema.sql`

Code → `internal/database/database.go` → `InitSchema()`

### Does the code contain at least one INSERT query?
Code → `internal/database/users.go` → `CreateUser()`

Code → `internal/database/posts.go` → `CreatePost()`

Code → `internal/database/comments.go` → `CreateComment()`

### Does the code contain at least one SELECT query?
Code → `internal/database/users.go` → `GetUserByID()`, `GetUserByEmail()`, `GetUserByUsername()`

Code → `internal/database/posts.go` → `GetPostByID()`, `GetAllPosts()`

Code → `internal/database/comments.go` → `GetCommentsByPost()`

### Register a user and check if the user is present in the database.
Terminal →

```bash
sqlite3 ./data/forum.db "SELECT * FROM users;"
```

### Create a post and check if the post is present in the database.
Terminal →

```bash
sqlite3 ./data/forum.db "SELECT * FROM posts;"
```

### Create a comment and check if the comment is present in the database.
Terminal →

```bash
sqlite3 ./data/forum.db "SELECT * FROM comments;"
```

---

## Docker

### Does the project have Dockerfiles?
Code → `Dockerfile`

### Build the Docker image. Did all images build successfully?
Terminal →

```bash
docker image build -t gameforum .
docker images
```

### Run the Docker container. Is the container running?
Terminal →

```bash
make docker-run
```

Then in another terminal:

```bash
docker ps -a
```

### Does the project have no unused Docker objects?
Terminal →

```bash
docker system df
```

---

## Functional

### As a guest, try to create a post. Are you forbidden?
Browser → Logout → try to access/create a post.

Route → `/posts/create`

Code → `internal/server/server.go` → `Run()`

Code → `internal/middleware/auth.go` → `RequireAuthentication()`

Code → `internal/handlers/posts.go` → `Create()`

### As a guest, try to create a comment. Are you forbidden?
Browser → Logout → open a post → try to create a comment.

Route → `/posts/{id}/comments`

Code → `internal/middleware/auth.go` → `RequireAuthentication()`

Code → `internal/handlers/comments.go` → `Create()`

### As a guest, try to like a comment/post. Are you forbidden from liking a post?
Browser → Logout → open a post and try to react.

Route → `/posts/react`

Code → `internal/middleware/auth.go` → `RequireAuthentication()`

Code → `internal/handlers/reactions.go` → `Post()`

### As a guest, try to dislike a comment. Are you forbidden?
Browser → Logout → open a post and try to dislike a comment.

Route → `/comments/react`

Code → `internal/middleware/auth.go` → `RequireAuthentication()`

Code → `internal/handlers/reactions.go` → `Comment()`

### As a registered user, try to create a comment. Are you able to do so?
Browser → Login → open a post → create a comment.

Code → `internal/handlers/comments.go` → `Create()`

Code → `internal/database/comments.go` → `CreateComment()`

### As a registered user, try to create an empty comment. Are you forbidden?
Browser → Login → open a post → submit an empty comment.

Code → `internal/handlers/comments.go` → `Create()`

### As a registered user, try to create a post. Are you able to do so?
Browser → Login → create a post.

Route → `/posts/create`

Code → `internal/handlers/posts.go` → `Create()`

Code → `internal/database/posts.go` → `CreatePost()`

### As a registered user, try to create an empty post. Are you forbidden?
Browser → Login → try to submit a post without the required content.

Code → `internal/handlers/posts.go` → `Create()`

### As a registered user, create a post and choose several categories. Are you able to do so?
Browser → Login → create a post → select multiple categories.

Code → `internal/handlers/posts.go` → `Create()`, `parseCategoryIDs()`, `createPostWithCategories()`

Code → `internal/database/posts.go` → `CreatePost()`, `AddPostCategory()`

### As a registered user, choose a category. Are you able to do so?
Browser → `http://localhost:8080/categories` → open a category.

Code → `internal/handlers/pages.go` → `Categories()`, `Category()`

Code → `internal/database/post_views.go` → `GetPostViewsByCategory()`

### As a registered user, like and dislike a post. Can you do so?
Browser → Login → open a post → Like → Dislike.

Code → `internal/handlers/reactions.go` → `Post()`

Code → `internal/database/reactions.go` → `SetPostReaction()`, `RemovePostReaction()`

### As a registered user, like and dislike a comment. Can you do so?
Browser → Login → open a post → react to a comment.

Code → `internal/handlers/reactions.go` → `Comment()`

Code → `internal/database/reactions.go` → `SetCommentReaction()`, `RemoveCommentReaction()`

### React to a post and refresh the page. Does the reaction count change correctly?
Browser → Login → react to a post → refresh the page.

Code → `internal/database/reactions.go` → `GetPostReactionCounts()`, `GetPostReactionByUser()`

Code → `internal/database/post_views.go` → `GetPostViewByID()`

### Like and then dislike the same post. Can the post be both liked and disliked at the same time?
Browser → Login → Like a post → Dislike the same post.

Code → `internal/handlers/reactions.go` → `Post()`

Code → `internal/database/reactions.go` → `SetPostReaction()`, `GetPostReactionByUser()`

### Can a registered user see all posts they created?
Browser → Login → `http://localhost:8080/activity`

Code → `internal/handlers/pages.go` → `Activity()`

Code → `internal/database/post_views.go` → `GetPostViewsByUser()`

### Can a registered user see all posts they liked?
Browser → Login → `http://localhost:8080/activity`

Code → `internal/handlers/pages.go` → `Activity()`

Code → `internal/database/post_views.go` → `GetLikedPostViewsByUser()`

### Can all users see the like and dislike counts of comments?
Browser → Open any post and check the reactions displayed on its comments.

Code → `internal/database/comments.go` → `GetCommentViewsByPost()`

Code → `internal/database/reactions.go` → `GetCommentReactionCounts()`

### Filter all posts from one category. Are all displayed posts from that category?
Browser → `http://localhost:8080/categories` → select one category.

Code → `internal/handlers/pages.go` → `Category()`

Code → `internal/database/post_views.go` → `GetPostViewsByCategory()`

### Did the server behave as expected and not crash?
Terminal →

```bash
make run
```

### Does the server use the right HTTP method?
Code → `internal/server/server.go` → `Run()`

Code → `internal/handlers/posts.go` → `Create()`, `Update()`, `Delete()`

Code → `internal/handlers/comments.go` → `Create()`, `Update()`, `Delete()`

Code → `internal/handlers/reactions.go` → `Post()`, `Comment()`

### Are all pages working? Is there a custom 404 page?
Browser → Open:

`http://localhost:8080/this-page-does-not-exist`

Code → `internal/handlers/pages.go` → `renderNotFound()`, `renderError()`

Code → `internal/handlers/errors.go` → `RenderErrorPage()`

Template → `templates/pages/error.html`

### Does the project handle HTTP 400?
Terminal →

```bash
curl -i -X POST http://localhost:8080/register \
  -d "username=" \
  -d "email=" \
  -d "password="
```

Code → `internal/handlers/pages.go` → `renderBadRequest()`

Code → `internal/handlers/errors.go` → `RenderErrorPage()`

### Does the project handle HTTP 500?
Code → `internal/handlers/pages.go` → `renderInternalServerError()`

Code → `internal/handlers/errors.go` → `RenderErrorPage()`

Template → `templates/pages/error.html`

### Does the project use only the allowed packages?
Terminal →

```bash
cat go.mod
```

File → `go.mod`

### As an auditor, is the project up to every standard?
Terminal →

```bash
gofmt -l .
go vet ./...
go test ./...
```

---

## General

### Does the project present a script to build images and containers?
Code → `Makefile` → `docker-build`, `docker-run`, `docker-stop`, `docker-clean`

Terminal →

```bash
make docker-build
make docker-run
```

### Is the password encrypted in the database?
Code → `internal/auth/password.go` → `HashPassword()`, `CheckPassword()`

Terminal →

```bash
sqlite3 ./data/forum.db "SELECT id, username, email, password_hash FROM users;"
```

---

## Basic

### Does the project run quickly and effectively?
Terminal →

```bash
curl -s -o /dev/null \
  -w "HTTP: %{http_code}\nTotal time: %{time_total}s\n" \
  http://localhost:8080/
```

Code → `internal/database/post_views.go` → `GetAllPostViews()`, `scanPostViews()`

### Is there a test file for this code?
Terminal →

```bash
find . -name "*_test.go" -type f | sort
```

Tests →

`internal/auth/password_test.go`

`internal/database/category_management_test.go`

`internal/database/database_test.go`

`internal/database/moderation_test.go`

`internal/database/moderator_management_test.go`

`internal/database/notifications_test.go`

`internal/database/post_views_test.go`

`internal/handlers/auth_test.go`

`internal/handlers/posts_test.go`

`internal/middleware/auth_test.go`

`internal/session/manager_test.go`

`internal/upload/image_test.go`

Run all tests:

```bash
go test ./...
```