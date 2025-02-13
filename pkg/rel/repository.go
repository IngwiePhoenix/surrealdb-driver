package rel

import (
	"context"

	"github.com/go-rel/rel"
)

var _ (rel.Repository) = (*Repository)(nil)

type Repository struct {
	rel.Repository
}

func NewRepo(adapter rel.Adapter) rel.Repository {
	return &Repository{
		Repository: rel.New(adapter),
	}
}

// Adapter implements rel.Repository.
func (r Repository) Adapter(ctx context.Context) rel.Adapter {
	return r.Repository.Adapter(ctx)
}

// Aggregate implements rel.Repository.
func (r Repository) Aggregate(ctx context.Context, query rel.Query, aggregate string, field string) (int, error) {
	return r.Repository.Aggregate(ctx, query, aggregate, field)
}

// Count implements rel.Repository.
func (r Repository) Count(ctx context.Context, collection string, queriers ...rel.Querier) (int, error) {
	return r.Repository.Count(ctx, collection, queriers...)
}

// Delete implements rel.Repository.
func (r Repository) Delete(ctx context.Context, entity any, mutators ...rel.Mutator) error {
	return r.Repository.Delete(ctx, entity, mutators...)
}

// DeleteAll implements rel.Repository.
func (r Repository) DeleteAll(ctx context.Context, entities any) error {
	return r.Repository.DeleteAll(ctx, entities)
}

// DeleteAny implements rel.Repository.
func (r Repository) DeleteAny(ctx context.Context, query rel.Query) (int, error) {
	return r.Repository.DeleteAny(ctx, query)

}

// Exec implements rel.Repository.
func (r *Repository) Exec(ctx context.Context, statement string, args ...any) (int, int, error) {
	return r.Repository.Exec(ctx, statement, args...)
}

// Find implements rel.Repository.
func (r Repository) Find(ctx context.Context, entity any, queriers ...rel.Querier) error {
	return r.Repository.Find(ctx, entity, queriers...)
}

// FindAll implements rel.Repository.
func (r Repository) FindAll(ctx context.Context, entities any, queriers ...rel.Querier) error {
	return r.Repository.FindAll(ctx, entities, queriers...)
}

// FindAndCountAll implements rel.Repository.
func (r Repository) FindAndCountAll(ctx context.Context, entities any, queriers ...rel.Querier) (int, error) {
	return r.Repository.FindAndCountAll(ctx, entities, queriers...)
}

// Insert implements rel.Repository.
func (r Repository) Insert(ctx context.Context, entity any, mutators ...rel.Mutator) error {
	return r.Repository.Insert(ctx, entity, mutators...)
}

// InsertAll implements rel.Repository.
func (r Repository) InsertAll(ctx context.Context, entities any, mutators ...rel.Mutator) error {
	return r.Repository.InsertAll(ctx, entities, mutators...)
}

// Instrumentation implements rel.Repository.
func (r Repository) Instrumentation(instrumenter rel.Instrumenter) {
	r.Repository.Instrumentation(instrumenter)
}

// Iterate implements rel.Repository.
func (r Repository) Iterate(ctx context.Context, query rel.Query, option ...rel.IteratorOption) rel.Iterator {
	return r.Repository.Iterate(ctx, query, option...)
}

// MustAggregate implements rel.Repository.
func (r Repository) MustAggregate(ctx context.Context, query rel.Query, aggregate string, field string) int {
	return r.Repository.MustAggregate(ctx, query, aggregate, field)
}

// MustCount implements rel.Repository.
func (r Repository) MustCount(ctx context.Context, collection string, queriers ...rel.Querier) int {
	return r.Repository.MustCount(ctx, collection, queriers...)
}

// MustDelete implements rel.Repository.
func (r Repository) MustDelete(ctx context.Context, entity any, mutators ...rel.Mutator) {
	r.Repository.MustDelete(ctx, entity, mutators...)
}

// MustDeleteAll implements rel.Repository.
func (r Repository) MustDeleteAll(ctx context.Context, entities any) {
	r.Repository.MustDeleteAll(ctx, entities)
}

// MustDeleteAny implements rel.Repository.
func (r Repository) MustDeleteAny(ctx context.Context, query rel.Query) int {
	return r.Repository.MustDeleteAny(ctx, query)

}

// MustExec implements rel.Repository.
func (r Repository) MustExec(ctx context.Context, statement string, args ...any) (int, int) {
	return r.Repository.MustExec(ctx, statement, args...)
}

// MustFind implements rel.Repository.
func (r Repository) MustFind(ctx context.Context, entity any, queriers ...rel.Querier) {
	r.Repository.MustFind(ctx, entity, queriers...)
}

// MustFindAll implements rel.Repository.
func (r Repository) MustFindAll(ctx context.Context, entities any, queriers ...rel.Querier) {
	r.Repository.MustFindAll(ctx, entities, queriers...)
}

// MustFindAndCountAll implements rel.Repository.
func (r Repository) MustFindAndCountAll(ctx context.Context, entities any, queriers ...rel.Querier) int {
	return r.Repository.MustFindAndCountAll(ctx, entities, queriers...)
}

// MustInsert implements rel.Repository.
func (r Repository) MustInsert(ctx context.Context, entity any, mutators ...rel.Mutator) {
	r.Repository.MustInsert(ctx, entity, mutators...)
}

// MustInsertAll implements rel.Repository.
func (r Repository) MustInsertAll(ctx context.Context, entities any, mutators ...rel.Mutator) {
	r.Repository.MustInsertAll(ctx, entities, mutators...)
}

// MustPreload implements rel.Repository.
func (r Repository) MustPreload(ctx context.Context, entities any, field string, queriers ...rel.Querier) {
	r.Repository.MustPreload(ctx, entities, field, queriers...)
}

// MustUpdate implements rel.Repository.
func (r Repository) MustUpdate(ctx context.Context, entity any, mutators ...rel.Mutator) {
	r.Repository.MustUpdate(ctx, entity, mutators...)
}

// MustUpdateAny implements rel.Repository.
func (r Repository) MustUpdateAny(ctx context.Context, query rel.Query, mutates ...rel.Mutate) int {
	return r.Repository.MustUpdateAny(ctx, query, mutates...)
}

// Ping implements rel.Repository.
func (r Repository) Ping(ctx context.Context) error {
	return r.Repository.Ping(ctx)
}

// Preload implements rel.Repository.
func (r Repository) Preload(ctx context.Context, entities any, field string, queriers ...rel.Querier) error {
	return r.Repository.Preload(ctx, entities, field, queriers...)
}

// Transaction implements rel.Repository.
func (r Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.Repository.Transaction(ctx, fn)
}

// Update implements rel.Repository.
func (r Repository) Update(ctx context.Context, entity any, mutators ...rel.Mutator) error {
	return r.Repository.Update(ctx, entity, mutators...)
}

// UpdateAny implements rel.Repository.
func (r Repository) UpdateAny(ctx context.Context, query rel.Query, mutates ...rel.Mutate) (int, error) {
	return r.Repository.UpdateAny(ctx, query, mutates...)
}
