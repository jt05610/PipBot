package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"pipbot/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "5000"

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the GraphQL server",
	Long:  `Starts the GraphQL server`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

func runServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
