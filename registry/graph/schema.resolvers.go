package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"time"

	"github.com/jenmud/submitSD/registry/graph/generated"
	"github.com/jenmud/submitSD/registry/graph/model"
	"github.com/jenmud/submitSD/registry/store"
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
	ttlStr := store.DefaultTTL

	if input.TTL != nil {
		ttlStr = *input.TTL
	}

	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return nil, err
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

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)
