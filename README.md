# SubmitSD

SubmitSD is a proof of concept server discovery service.

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

## NOTE

This is still under development and not feature complete.