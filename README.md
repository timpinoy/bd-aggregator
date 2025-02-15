
# RSS Feed Aggregator

A command-line RSS feed aggregator written in Go for the [boot.dev](https://www.boot.dev/) "Build a Blog Aggregator" course.

## Prerequisites

- Go 1.23 or later
- PostgreSQL

## Installation

1. Clone the repository:

```bash
git clone https://github.com/timpinoy/bd-aggregator
cd bd-aggregator
```

2. Install dependencies:

```bash
go mod download
```

3. Set up the database:

* Create a database in PostgreSQL, e.g. gator
* Run the migration scripts with goose

```bash
cd sql/schema
goose postgres "postgres://postgres:postgres@localhost:5432/gator" up
```

4. Create configuration:
Create a .gatorconfig.json in your home directory:

```json
{
    "db_url": "postgres://username:password@localhost:5432/dbname"
}
```

5. Build the project:

```bash
go build
```

## Usage

```bash
./gator register <username>    # Create new user
./gator login <username>       # Login as user
./gator users                  # List all users
./gator reset                  # Delete all users
./gator addfeed <name> <url>   # Add new RSS feed, the current user will automatically follow
./gator feeds                  # List all feeds
./gator follow <url>           # Follow a feed
./gator unfollow <url>         # Unfollow a feed
./gator following              # List your followed feeds
./gator browse [limit]         # View posts (default limit: 2 posts)
./gator agg <interval>         # Start aggregating feeds
```

### Examples

```bash
# Register and login
./gator register tim
./gator login tim

# Add the Go blog
./gator addfeed "Go Blog" "https://go.dev/blog/feed.atom"

# View content
./gator agg 5m               # Aggregate feeds every 5 minutes
./gator browse 5             # Show 5 latest posts
```