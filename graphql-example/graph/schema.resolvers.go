package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"context"
	"fmt"

	"github.com/MKMuhammetKaradag/go-microservice/graphql-example/graph/model"
)

// CreateTodo is the resolver for the createTodo field.
func (r *mutationResolver) CreateTodo(ctx context.Context, input model.NewTodo) (*model.Todo, error) {
	panic(fmt.Errorf("not implemented: CreateTodo - createTodo"))
}

// AddMessage is the resolver for the addMessage field.
func (r *mutationResolver) AddMessage(ctx context.Context, msg string) (string, error) {
	go func() {
		r.Messages <- msg
	}()

	message := "Mesaj gönderildi: " + msg
	fmt.Println(message) // Artık çalışacak!
	return message, nil
}

// Todos is the resolver for the todos field.
func (r *queryResolver) Todos(ctx context.Context) ([]*model.Todo, error) {
	panic(fmt.Errorf("not implemented: Todos - todos"))
}

// Message is the resolver for the message field.
func (r *queryResolver) Message(ctx context.Context) (string, error) {
	fmt.Println("hello query")
	return "Merhaba GraphQL Dünyası!", nil
	// panic(fmt.Errorf("not implemented: Message - message"))
}

// MessageAdded is the resolver for the messageAdded field.
func (r *subscriptionResolver) MessageAdded(ctx context.Context) (<-chan string, error) {
	// fmt.Println("subscription")
	msgChan := make(chan string,10)

	go func() {
		defer close(msgChan)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Abonelik iptal edildi.")
				return
			case msg := <-r.Messages:
				fmt.Println("Yeni mesaj abonelere gönderildi:", msg)
				msgChan <- msg
			}
		}
	}()

	return msgChan, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
