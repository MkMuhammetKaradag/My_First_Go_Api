package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/MKMuhammetKaradag/go-microservice/graphql-example/graph"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8084"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	resolver := &graph.Resolver{
		Messages: make(chan string), // Buffered channel ekleyerek bloklanmasını önlüyoruz
	}

	go func() {
		for msg := range resolver.Messages {
			fmt.Println("📩 Aboneye iletilen mesaj:", msg)
		}
	}()

	// Yeni mesajlar eklemek için test amaçlı goroutine
	// go func() {
	// 	for {
	// 		// time.Sleep(3 * time.Second)
	// 		// resolver.Messages <- fmt.Sprintf("Yeni mesaj: %s", time.Now().Format(time.RFC3339))
	// 	}
	// }()

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Tüm originlerden bağlantıyı kabul et
			},
		},
	})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
