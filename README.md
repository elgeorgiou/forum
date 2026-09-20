# GameForum

GameForum is a full-stack gaming discussion forum built in Go.

Users can create an account, authenticate locally or through OAuth, create discussions, upload images, comment, react to content, receive notifications, search the forum, track their activity, and interact with gaming categories.

The application also includes a role-based moderation system with users, moderators, and administrators.

---

## Features

### Authentication

- User registration with username, email, and password
- Login and logout
- Password hashing with bcrypt
- Cookie-based sessions
- One active session per user
- Session expiration
- GitHub OAuth authentication
- Google OAuth authentication
- Persistent OAuth accounts

### Discussions

Registered users can:

- Create posts
- Edit their own posts
- Delete their own posts
- Add a post to one or multiple categories
- Upload an image with a post
- Comment on posts
- Edit their own comments
- Delete their own comments

Guests can browse the forum without creating or modifying content.

### Reactions

Registered users can like or dislike:

- Posts
- Comments

A user cannot simultaneously like and dislike the same content.

Reaction counts are visible to all visitors.

### Categories

Posts can belong to one or multiple gaming categories.

The forum includes category pages and category-based filtering.

Administrators can:

- Create categories
- Edit categories
- Delete categories

### Search

The navigation bar contains forum search with live suggestions.

Search supports:

- Posts
- Categories
- Category-related discussions

Pressing Enter opens the full search results page.

### Filters

Users can filter discussions by:

- Category
- Posts created by the current user
- Posts liked by the current user

### Activity

Authenticated users have an activity page containing:

- Created posts
- Liked posts
- Disliked posts
- Comments

### Notifications

Users receive notifications when another user:

- Comments on their post
- Likes their post
- Dislikes their post
- Likes their comment
- Dislikes their comment

Moderators can also receive notifications when an administrator responds to a report.

Notifications can be marked individually as read or all marked as read.

### Image Uploads

Registered users can attach images to posts.

Supported formats:

- PNG
- JPEG
- GIF

Maximum upload size:

```text
20 MB
```

Uploaded files are validated using their detected content type and stored with generated filenames.

### Moderation

The forum supports the following access levels:

#### Guest

- Browse public forum content

#### User

- Create posts and comments
- Edit/delete own content
- React to content
- Receive notifications
- Request moderator status

#### Moderator

Includes normal user permissions plus:

- Moderate forum content
- Delete inappropriate posts and comments
- Report posts and comments to administrators

#### Administrator

Includes moderation capabilities plus:

- Review moderator requests
- Promote users to moderators
- Demote moderators
- Review reports
- Respond to reports
- Manage categories
- Access the administration dashboard

---

## Technologies

### Backend

- Go 1.22.2
- Go `net/http`
- SQLite
- `github.com/mattn/go-sqlite3`
- `golang.org/x/crypto/bcrypt`

### Frontend

- HTML templates
- CSS
- Vanilla JavaScript
- Inline SVG icons

### Infrastructure

- Docker
- Make
- SQLite persistent storage

---

## Project Structure

```text
forum/
├── cmd/
│   └── main.go
│
├── internal/
│   ├── auth/
│   │   ├── password.go
│   │   └── password_test.go
│   │
│   ├── database/
│   │   ├── schema/
│   │   │   └── schema.sql
│   │   ├── categories.go
│   │   ├── comments.go
│   │   ├── dashboard.go
│   │   ├── database.go
│   │   ├── moderation.go
│   │   ├── notifications.go
│   │   ├── oauth.go
│   │   ├── posts.go
│   │   ├── post_views.go
│   │   ├── reactions.go
│   │   ├── report_views.go
│   │   ├── search.go
│   │   ├── seed.go
│   │   ├── sessions.go
│   │   └── users.go
│   │
│   ├── handlers/
│   │   ├── admin_categories.go
│   │   ├── admin_moderators.go
│   │   ├── auth.go
│   │   ├── comments.go
│   │   ├── github_oauth.go
│   │   ├── google_oauth.go
│   │   ├── moderation.go
│   │   ├── notifications.go
│   │   ├── pages.go
│   │   ├── posts.go
│   │   ├── reactions.go
│   │   └── search.go
│   │
│   ├── middleware/
│   │   └── auth.go
│   │
│   ├── models/
│   │   └── page.go
│   │
│   ├── server/
│   │   └── server.go
│   │
│   ├── session/
│   │   └── manager.go
│   │
│   └── upload/
│       └── image.go
│
├── static/
│   ├── css/
│   ├── images/
│   ├── js/
│   └── uploads/
│
├── templates/
│   ├── layouts/
│   ├── pages/
│   └── partials/
│
├── data/
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

# Running the Project

GameForum can be started in several ways.

## Prerequisites

For native execution:

- Go 1.22.2 or compatible version
- GCC / CGO support for `go-sqlite3`

For containerized execution:

- Docker
- Docker Desktop when using WSL2 on Windows

---

## 1. Run with Make

This is the recommended development command.

From the project root:

```bash
make run
```

This executes:

```bash
go run ./cmd
```

The server will be available at:

```text
http://localhost:8080
```

---

## 2. Run Directly with Go

The project can also be started without Make:

```bash
go run ./cmd
```

Then open:

```text
http://localhost:8080
```

---

## 3. Run with Docker and Make

Build the Docker image:

```bash
make docker-build
```

Start the application:

```bash
make docker-run
```

Then open:

```text
http://localhost:8080
```

The Docker setup mounts:

```text
./data
```

to:

```text
/app/data
```

and:

```text
./static/uploads
```

to:

```text
/app/static/uploads
```

This keeps the SQLite database and uploaded images outside the container.

Stop a running container from another terminal with:

```bash
make docker-stop
```

Remove the container and Docker image with:

```bash
make docker-clean
```

---

## 4. Run with Docker Commands Directly

The project can also be run without the Makefile.

Build:

```bash
docker build -t gameforum .
```

Then run the image while passing the OAuth environment variables from the current shell:

```bash
docker run --rm \
    --name gameforum \
    -p 8080:8080 \
    -e GITHUB_CLIENT_ID \
    -e GITHUB_CLIENT_SECRET \
    -e GITHUB_REDIRECT_URL \
    -e GOOGLE_CLIENT_ID \
    -e GOOGLE_CLIENT_SECRET \
    -e GOOGLE_REDIRECT_URL \
    -v "$(pwd)/data:/app/data" \
    -v "$(pwd)/static/uploads:/app/static/uploads" \
    gameforum
```

Then open:

```text
http://localhost:8080
```

---

# Environment Variables

OAuth authentication requires environment variables.

Create a `.env` file in the project root:

```bash
touch .env
```

The project Makefile supports the following format:

```bash
export GITHUB_CLIENT_ID=your_github_client_id
export GITHUB_CLIENT_SECRET=your_github_client_secret
export GITHUB_REDIRECT_URL=http://localhost:8080/auth/github/callback

export GOOGLE_CLIENT_ID=your_google_client_id
export GOOGLE_CLIENT_SECRET=your_google_client_secret
export GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
```

The `.env` file is ignored by Git and must never be committed.

When using:

```bash
make run
```

or:

```bash
make docker-run
```

the Makefile loads these variables automatically.

When using `go run ./cmd` directly, load the variables into the shell first:

```bash
source .env
go run ./cmd
```

When using Docker directly, load them first:

```bash
source .env
```

and then run the Docker command shown above.

---

# OAuth Setup

## GitHub

Create a GitHub OAuth application and configure its callback URL as:

```text
http://localhost:8080/auth/github/callback
```

Add the generated client ID and client secret to `.env`.

## Google

Create Google OAuth credentials and configure the authorized redirect URI as:

```text
http://localhost:8080/auth/google/callback
```

Add the generated client ID and client secret to `.env`.

---

# Database

GameForum uses SQLite.

The schema is defined in:

```text
internal/database/schema/schema.sql
```

The database contains tables for:

- Users
- Sessions
- Categories
- Posts
- Post/category relationships
- Comments
- Post reactions
- Comment reactions
- OAuth accounts
- Moderator requests
- Reports
- Notifications

Foreign keys are used to maintain relationships between records.

The application also creates indexes for frequently queried relationships.

Runtime database files are stored under:

```text
data/
```

Database files are ignored by Git.

---

# Testing

Run all tests with:

```bash
make test
```

or directly:

```bash
go test ./...
```

Tests cover important areas including:

- Password handling
- Authentication
- Sessions
- Authentication middleware
- Database operations
- Post views
- Notifications
- Moderation
- Moderator management
- Category management
- Post handlers
- Image validation and uploads

---

# Formatting

Format all Go files:

```bash
make fmt
```

Equivalent command:

```bash
gofmt -w .
```

---

# Static Analysis

Run Go vet:

```bash
make vet
```

Equivalent command:

```bash
go vet ./...
```

---

# Full Development Check

Run formatting, tests, and static analysis together:

```bash
make check
```

This runs:

```text
gofmt -w .
go test ./...
go vet ./...
```

---

# Useful Make Commands

```text
make run           Start the forum locally
make test          Run all Go tests
make fmt           Format Go source files
make vet           Run Go static analysis
make check         Format, test, and vet the project
make docker-build  Build the Docker image
make docker-run    Run the forum in Docker
make docker-stop   Stop the Docker container
make docker-clean  Remove the Docker container and image
```

---

# Application Pages

The forum includes:

```text
/                       Home
/auth                   Login / registration
/categories             Categories
/recent                 Recent posts
/about                  About
/search                  Search
/profile                 User profile
/activity                User activity
/notifications           Notifications
/dashboard               Administrator dashboard
/moderation/request      Moderator request
/admin/moderators        Moderator management
/admin/categories        Category management
```

Individual categories and discussions use dynamic routes.

---

# Security

The project includes several security and validation measures:

- Passwords are hashed with bcrypt
- Session authentication uses cookies
- Session records expire
- A user has one active session
- Protected routes require authentication
- Moderator and administrator actions use role checks
- SQL operations use parameterized queries
- Uploaded images have a 20 MB request limit
- Uploaded file types are validated using detected content type
- Uploaded files receive generated filenames
- OAuth secrets are stored outside the repository
- Database foreign keys are enabled
- User input is rendered through Go HTML templates

---

# Docker Persistence

The Docker container itself is disposable.

Persistent application data is mounted from the host:

```text
data/               -> SQLite database
static/uploads/      -> User uploaded images
```

This means rebuilding or removing the container does not remove the forum database or uploaded images.

---

# Authors

- elgeorgiou
- gpapadaki