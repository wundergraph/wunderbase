<img src="/img/img.png" align="left"
alt="Size Limit logo by Anton Lovchikov" width="232" height="237">

is a Serverless SQLite database exposed through GraphQL. 
It can sleep for days and wakes up in less than a second.
During sleep, you only pay for storage. It is capable to serve a medium traffic website.

* **Firecracker (Fly.io Machines)** Runs on top of the same technology as AWS Fargate.
* **Scale to zero**. Cold starts in less than 600ms.
* **Fast**. 10k reads / 2k writes requests per second.
* **Cheap** Fly.io free tier allows your to run it almost for free.

<br>
<p align="center">Love the idea? :heartbeat: Read the announcement <a href="https://wundergraph.com/blog/wunderbase_serverless_graphql_database_on_top_of_sqlite_firecracker_and_prisma">blog post</a> to get more insights!</p>

## Getting Started

### 1. Install dependencies

- Mac
    ```
    sh install-prisma-darwin.sh
    ```

- Linux
    ```
    sh install-prisma-linux.sh
    ```

### 2. Start the server

- Run the application locally
    ```sh
    ENABLE_SLEEP_MODE=false go run main.go
    ```

- Running with Docker
    ```sh
    docker run -p 4466:4466 -e ENABLE_SLEEP_MODE=false wundergraph/wunderbase
    ```

### 3. Visit the playground

Open [http://0.0.0.0:4466](http://0.0.0.0:4466) in your browser.

## Running on fly Machines

Check out the fly.io [Machines documentation](https://fly.io/docs/reference/machines/) on how to deploy WunderBase to fly.io.
You can build the image yourself using the Dockerfile in this repo,
or just use a pre-built image from [Docker Hub](https://hub.docker.com/r/wundergraph/wunderbase).

## Testing

> **Note**: Requires Docker

`go test ./...`

## Firecracker & Fly Machines

- **Sleep Mode**
    - By default, sleep mode is enabled. This means, if there is no traffic, WunderBase will sleep.
    Sleep mode lets WunderBase exit with exit code 0.
    This indicated to fly that the application wants to sleep.
- **Cold Starts**
    - I've measured an avg cold start time of 400-600ms.
- **Backups**
    - We're looking at integrating Litestream or a similar service for backups.
      If you'd like to contribute, please reach out!

### Internals

WunderBase uses the Prisma Query Engine internally to translate GraphQL queries to SQL.
We've added a proxy written in Golang to add GraphiQL 2.0 as the frontend and handle the lifecycle to make the service sleep when it's not in use.
We're also using the Prisma Migration Engine to run migrations.

### Security

I do not recommend to expose WunderBase to the public internet.
The intended use case is to run it on a private network and expose it to your frontend via an API Gateway,
like [WunderGraph](https://github.com/wundergraph/wundergraph)!

> **Note**: Get [Early Access](https://wundergraph.com/#early-access) to WunderGraph Cloud - " The Vercel for Backend " 🪐
The Serverless API Developer Platform 🪄
