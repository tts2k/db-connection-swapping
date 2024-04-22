# SQLite runtime connection swap
Fiddling with swapping sqlite database connection during runtime

## The Problems
1. You're working on an app that requires mostly reading from disk, and you 
want to leverage SQLite low-latency output for that.
2. You want to push a large amount of data periodically to said database
to keep your content up-to-date, but the concurrent writes performance of
SQLite triggers your OCD, even though your app will probably never have
enough user for that to be a problem.

## The Solutions
1. Create a new database (or clone your old one) and push your data there,
as creating a new SQLite database is as easy as `touch new.db` 
(it's actually so easy that 
[one database per user](https://turso.tech/database-per-tenant)
is viable).
2. Swap the connection during runtime when you have pushed the new data to
the database. This will minimize the impact created by writing to the database
while still reading it.

## Running the app
### 1. Start the server
It's gonna start on port 5000.
```sh
CGO_ENABLED=1 GOOS=linux go run .
```
Note: The golang SQLite driver is a cgo package so you'll probably have trouble
building it on Windows

### 2. Connect to the websocket
Connect your websocket client to the `/ws` endpoint. You should see the
current database name being send to you each second once connected

### 3. Swap your database
Send a `POST` request to the `/createdb` endpoint and a new database will
be created with a timestamp postfix. After that, you should see the change
reflected on you existing websocket connection
