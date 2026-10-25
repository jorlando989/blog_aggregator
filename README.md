# Prerequisites
- Postgres
- Go

# Setup
1. Install gator CLI using go install ??
2. Create a config file called ".gatorconfig.json" in the root folder of your computer
3. Add the following to the gatorconfig file:
```
{"DbUrl":"<your_database_url>?sslmode=disable","CurrentUserName":""}
```
4. Run the program using go run . <command_name>

# Available Commands
- login <username>
- register <username>
- reset: unregisters all users
- users: returns a list of all registered users
- agg <time_between_requests>: periodically collects the RSS feeds for new posts
- addfeed <url>: adds a new feed and subscribes the current user to it
- feeds: lists all feeds followed by any user
- follow <url>: adds the current user as a follower of an existing feed
- following: lists all the feeds the current user is following
- unfollow <url>: unfollows the current user from the given feed 
- browse: shows all the posts from the feeds the current user is following