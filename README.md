# WunderBase: Serverless SQLite-based GraphQL Database built on top of Firecracker (fly.io)

WunderBase is a Serverless Database with an HTTP Interface (GraphQL) and SQLite as the storage.

WunderBase only runs when it is processing requests.
It can sleep for days and wakes up in less than a second.
During sleep, you only pay for storage.

## Firecracker / fly Machines

It's built on top of Firecracker / fly.io Machines.
Fly gives you a generous free tier with 3GB of storage and 160GB of outbound traffic.
Beyond that, 1GB of storage costs $0.15 per month.

This means, for small to medium workloads, you can run WunderBase more or less for free!
WunderBase is Open Source, so you're not locked into a proprietary solution.
It's just a Docker container with a volume attached, so you can run it anywhere,
even on your local machine.

## Read the announcement blog post!

## Running locally

Install dependencies:

`sh install-prisma-darwin.sh` (MacOS)

or 

`sh install-prisma-linux.sh` (Linux)

then 

`ENABLE_SLEEP_MODE=false go run main.go`

## Running in Docker

`docker run -p 4466:4466 -e ENABLE_SLEEP_MODE=false wundergraph/wunderbase`

## Running on fly Machines

Check out the fly.io [Machines documentation](https://fly.io/docs/reference/machines/) on how to deploy WunderBase to fly.io.
You can build the image yourself using the Dockerfile in this repo,
or just use a pre-built image from Docker Hub: `wundergraph/wunderbase`.

## Testing

Requires Docker

`go test ./...`

## Firecracker & Fly Machines

## Sleep Mode

By default, sleep mode is enabled. This means, if there is no traffic, WunderBase will sleep.
Sleep mode lets WunderBase exit with exit code 0.
This indicated to fly that the application wants to sleep.

## Cold Starts

I've measured an avg cold start time of 400-600ms.

## Backups

We're looking at integrating Litestream or a similar service for backups.
If you'd like to contribute, please reach out!

## Internals

WunderBase uses the Prisma Query Engine internally to translate GraphQL queries to SQL.
We've added a proxy written in Golang to add GraphiQL 2.0 as the frontend and handle the lifecycle to make the service sleep when it's not in use.
We're also using the Prisma Migration Engine to run migrations.

## Security

I do not recommend to expose WunderBase to the public internet.
The intended use case is to run it on a private network and expose it to your frontend via an API Gateway,
like [WunderGraph](https://github.com/wundergraph/wundergraph)!