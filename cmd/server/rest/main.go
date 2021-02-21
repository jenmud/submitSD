package main

import (
	"context"
	"flag"
	"net"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/jenmud/registry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// setupRoutes sets up all the routes for the Rest app.
func setupRoutes(app *fiber.App, client registry.RegistryServiceClient) {

	route := app.Group("/api/v1")

	route.Get(
		"/routes",
		func(c *fiber.Ctx) error { return c.JSON(app.Stack()) },
	)

	route.Get(
		"services",
		func(c *fiber.Ctx) error {
			nodes, err := client.Search(c.Context(), &registry.SearchReq{Name: "*"})
			if err != nil {
				return err
			}
			return c.JSON(nodes)
		},
	)
}

func main() {
	name := flag.String("service-name", "registry.rest.srv", "Service registry name.")
	addr := flag.String("addr", ":8080", "Address to listen and accept client connections.")
	srv := flag.String("registry", ":8000", "Registry service to connect to.")
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		logrus.Fatalf("Error starting registry RPC service %q: %s", *addr, err)
	}

	options := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(*srv, options...)
	if err != nil {
		logrus.Fatalf("Error connecting to registry service %s: %s", *srv, err)
	}

	client := registry.NewRegistryServiceClient(conn)
	app := fiber.New()

	node := &registry.Node{Name: *name, Address: listener.Addr().Network() + "://" + listener.Addr().String()}
	rnode, err := client.Register(context.Background(), node)
	if err != nil {
		logrus.Fatalf("Error registrying node with registry service %s: %s", *srv, err)
	}

	defer client.Unregister(context.Background(), rnode)
	logrus.Infof("REST service (%s, %s) is listening and accepting client connections on %s", rnode.GetName(), rnode.GetUid(), listener.Addr())

	setupRoutes(app, client)
	app.Listener(listener)
}
