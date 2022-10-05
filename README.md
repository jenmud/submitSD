# SubmitSD

SubmitSD is a proof of concept service discovery service/server.

**It is still under development and not feature complete.**

## Setup your environment

Before you do anything, always run the following command to generate all the necessary code used in the project.

```bash
$ make generate
```

## Building binaries

Running `make` will generated all the necessary code and build Linux and Windows binaries and place them under the `builds` directory.

```bash
$ make
```

## Running the service

Execute the binary which will start listening on address `localhost:8081`. Use `-h` for help.

```bash
$ cd builds
$ ./submitSD server
[GIN-debug] Listening and serving HTTP on localhost:8081
```

Once the server is running and accepting client connections, you can naviate to the address above in a browser or any other HTTP method.

Accessing the service via a browser will take you to the playground where you can start playing with the API.

## Features in the pipeline

Please note that this is still in early development and below is a list of features that I would like to still implement.

* Pubsub events via subscriptions
* Service queries by
  * type
  * name
  * version
  * etc....
* indexing for fast lookups and queries
* persistance (optional)
* gRPC service
* RestAPI
* Proper tests
* ...more to come

# Example

## Create a service

Query:

```json
mutation create {
  create(
    input: {id: "server", name: "submitSD", description: "graphQL server", version: "v0.0.0", address: "localhost:8081", ttl: "1m"}
  ) {
    id
    created_at
    name
    description
    address
    ttl
    expires_at
  }
}
```

Response:

```json
{
  "data": {
    "create": {
      "id": "server",
      "created_at": "2022-10-06T00:11:31.628603+11:00",
      "name": "submitSD",
      "description": "graphQL server",
      "address": "localhost:8081",
      "ttl": "1m",
      "expires_at": "2022-10-06T00:12:31.628606+11:00"
    }
  }
}
```

## Renew a service before it expires

Query:

```json
mutation renew {
  renew(input: {id: "server", ttl: "1m"}) {
    id
    created_at
    name
    description
    address
    ttl
    expires_at
  }
}
```

Response:

```json
{
  "data": {
    "renew": {
      "id": "server",
      "created_at": "2022-10-06T00:15:25.217102+11:00",
      "name": "submitSD",
      "description": "graphQL server",
      "address": "localhost:8081",
      "ttl": "1m0s",
      "expires_at": "2022-10-06T00:16:35.567788+11:00"
    }
  }
}
```

## Get all services

Query:

```json
query all {
  services {
    id
    created_at
    name
    description
    address
    ttl
    expires_at
  }
}
```

Response:

```json
{
  "data": {
    "services": [
      {
        "id": "server",
        "created_at": "2022-10-06T00:13:34.830831+11:00",
        "name": "submitSD",
        "description": "graphQL server",
        "address": "localhost:8081",
        "ttl": "1m",
        "expires_at": "2022-10-06T00:14:34.830832+11:00"
      }
    ]
  }
}
```