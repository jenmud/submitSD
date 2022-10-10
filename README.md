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

- [x] Add services
- [x] Renew services TTL
- [x] Auto expiry of services
- [x] GraphQL and served up via HTTP router  
- [x] Pubsub events via subscriptions
- [ ] Auto expiry and cleanups
  - [ ] Backend store should periodically clean up expired services
  - [ ] Backend store should take clean up callbacks and publish cleaned/expired services
- [ ] Service queries by
  - [ ] type
  - [ ] name
  - [ ] version
  - [ ] etc....
- [ ] Indexing for fast lookups and queries
- [ ] Persistance (optional)
- [ ] gRPC service
- [ ] RestAPI
- [ ] Web based GUI with realtime updates
- [x] Extend the message to include service config
- [ ] Add Update API calls to updated services
- [ ] Proper tests
...more to come

# Example

## Create a service

Query:

```graphql
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

```graphql
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

```graphql
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

## Events subscription

Query:

```grapql
subscription pubsub {
  events {
    timestamp
    event
    service {
      name
    }
  }
}
```

Response:

```json
{
  "data": {
    "events": {
      "timestamp": "2022-10-10T22:53:37.908826+11:00",
      "event": "RENEWED",
      "service": {
        "name": "submitSD"
      }
    }
  }
}
```