package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jenmud/submitSD/registry/graph/generated"
	"github.com/jenmud/submitSD/registry/graph/model"
)

// Create is the resolver for the create field.
func (r *mutationResolver) Create(ctx context.Context, input model.NewServiceInput) (*model.Service, error) {
	service, err := r.store.Add(ctx, input)
	if err != nil {
		return nil, err
	}
	return &service, err
}

// Renew is the resolver for the renew field.
func (r *mutationResolver) Renew(ctx context.Context, input model.RenewServiceInput) (*model.Service, error) {
	ttl := r.store.Config.TTL

	if input.TTL != nil {
		t, err := time.ParseDuration(*input.TTL)
		if err != nil {
			return nil, err
		}
		ttl = t
	}

	service, err := r.store.Renew(input.ID, ttl)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

// Expire is the resolver for the expire field.
func (r *mutationResolver) Expire(ctx context.Context, input *model.ExpireServiceInput) (bool, error) {
	err := r.store.Expire(input.ID)
	return err == nil, err
}

// Services is the resolver for the services field.
func (r *queryResolver) Services(ctx context.Context) ([]*model.Service, error) {
	found := r.store.Items()
	items := make([]*model.Service, 0, len(found))

	for i := range found {
		item := found[i]
		items = append(items, &item)
	}

	return items, nil
}

// Service is the resolver for the service field.
func (r *queryResolver) Service(ctx context.Context, id string) (*model.Service, error) {
	service, err := r.store.Get(id)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

// Events is the resolver for the events field.
func (r *subscriptionResolver) Events(ctx context.Context) (<-chan *model.Event, error) {
	ch := make(chan *model.Event, 1)
	id := uuid.NewString()

	go func() {
		<-ctx.Done()
		r.lock.Lock()
		delete(r.subscribers, id)
		log.Printf("removed subscriber %s", id)
		r.lock.Unlock()
	}()

	r.lock.Lock()
	r.subscribers[id] = ch
	log.Printf("added new subscriber %s", id)
	r.lock.Unlock()

	return ch, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type (
	mutationResolver     struct{ *Resolver }
	queryResolver        struct{ *Resolver }
	subscriptionResolver struct{ *Resolver }
)
