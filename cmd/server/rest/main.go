package main

import (
	"context"
	"encoding/json"
	"flag"
	"net"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	route.Get(
		"services/:name",
		func(c *fiber.Ctx) error {
			nodes, err := client.Search(c.Context(), &registry.SearchReq{Name: c.Params("name")})
			if err != nil {
				return err
			}

			return c.JSON(nodes)
		},
	)

	route.Get(
		"node/:uid",
		func(c *fiber.Ctx) error {
			node, err := client.Get(c.Context(), &registry.GetReq{Uid: c.Params("uid")})
			if err != nil {
				return err
			}

			return c.JSON(node)
		},
	)

	route.Post(
		"node",
		func(c *fiber.Ctx) error {
			node := new(registry.Node)
			payload := c.Body()

			if err := json.Unmarshal(payload, node); err != nil {
				return err
			}

			resp, err := client.Register(c.Context(), node)
			if err != nil {
				return err
			}

			return c.JSON(resp)
		},
	)

	route.Delete(
		"node/:uid",
		func(c *fiber.Ctx) error {
			node, err := client.Get(c.Context(), &registry.GetReq{Uid: c.Params("uid")})
			if err != nil {
				return err
			}

			nodes, err := client.Unregister(c.Context(), node)
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
	expiry := flag.String("expire-duration", "5s", "Node expiry duration")
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

	app := fiber.New(
		fiber.Config{
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
	)

	app.Use(cors.New(cors.ConfigDefault))

	node := &registry.Node{Name: *name, Address: listener.Addr().Network() + "://" + listener.Addr().String(), ExpiryDuration: *expiry}
	rnode, err := client.Register(context.Background(), node)
	if err != nil {
		logrus.Fatalf("Error registrying node with registry service %s: %s", *srv, err)
	}

	// Run a background updating worker
	go func() {
		d, err := time.ParseDuration(*expiry)
		if err != nil {
			logrus.Fatal(err)
		}

		xd := d - time.Second
		ticker := time.NewTicker(xd)
		for {
			<-ticker.C
			n, err := client.Register(context.Background(), rnode)
			if err != nil {
				logrus.Errorf("Error updating node %s: %s", n, err)
				continue
			}
			rnode = n
		}
	}()

	defer client.Unregister(context.Background(), rnode)
	logrus.Infof("REST service (%s, %s) is listening and accepting client connections on %s", rnode.GetName(), rnode.GetUid(), listener.Addr())

	setupRoutes(app, client)
	app.Listener(listener)
}
