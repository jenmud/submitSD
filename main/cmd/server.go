/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/jenmud/submitSD/registry/graph"
	"github.com/jenmud/submitSD/registry/graph/generated"
	"github.com/jenmud/submitSD/registry/graph/model"
	"github.com/jenmud/submitSD/registry/store"
	"github.com/spf13/cobra"
)

// Defining the Graphql handler
func graphqlHandler() gin.HandlerFunc {
	config := store.Config{
		TTL: store.DefaultConfig.TTL,
		// CleanupInterval: store.DefaultConfig.CleanupInterval,
		CleanupInterval: 5 * time.Second,
		Callback:        nil,
	}

	store := store.New(config)
	// defer store.Close()

	r := graph.NewResolver(store)

	/*
		set the a publishing callback so that subscribers get notified when
		something interesting happens on the store.
	*/
	store.SetCallback(
		func(event model.Event) {
			r.Publish(&event)
		},
	)

	h := handler.NewDefaultServer(
		generated.NewExecutableSchema(generated.Config{Resolvers: r}),
	)

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Defining the Playground handler
func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func RunGraphQLServer(addr string) {
	r := gin.Default()
	r.Any("/query", graphqlHandler())
	r.GET("/", playgroundHandler())
	r.Run(addr)
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "server is a registry which is used for storing services.",
	Long: `server is a registry which is used for storing services and
querying for services.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunGraphQLServer(cmd.Flags().Lookup("addr").Value.String())
		log.Printf("server shutdown")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringP("addr", "a", "localhost:8081", "Listen and accept client connection on this address")
}
