// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/oauth"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/user"
)

// OAuthCreate is the builder for creating a OAuth entity.
type OAuthCreate struct {
	config
	mutation *OAuthMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (oc *OAuthCreate) SetCreatedAt(t time.Time) *OAuthCreate {
	oc.mutation.SetCreatedAt(t)
	return oc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (oc *OAuthCreate) SetNillableCreatedAt(t *time.Time) *OAuthCreate {
	if t != nil {
		oc.SetCreatedAt(*t)
	}
	return oc
}

// SetUpdatedAt sets the "updated_at" field.
func (oc *OAuthCreate) SetUpdatedAt(t time.Time) *OAuthCreate {
	oc.mutation.SetUpdatedAt(t)
	return oc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (oc *OAuthCreate) SetNillableUpdatedAt(t *time.Time) *OAuthCreate {
	if t != nil {
		oc.SetUpdatedAt(*t)
	}
	return oc
}

// SetProvider sets the "provider" field.
func (oc *OAuthCreate) SetProvider(s string) *OAuthCreate {
	oc.mutation.SetProvider(s)
	return oc
}

// SetSub sets the "sub" field.
func (oc *OAuthCreate) SetSub(s string) *OAuthCreate {
	oc.mutation.SetSub(s)
	return oc
}

// SetID sets the "id" field.
func (oc *OAuthCreate) SetID(u uuid.UUID) *OAuthCreate {
	oc.mutation.SetID(u)
	return oc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (oc *OAuthCreate) SetNillableID(u *uuid.UUID) *OAuthCreate {
	if u != nil {
		oc.SetID(*u)
	}
	return oc
}

// SetUserID sets the "user" edge to the User entity by ID.
func (oc *OAuthCreate) SetUserID(id uuid.UUID) *OAuthCreate {
	oc.mutation.SetUserID(id)
	return oc
}

// SetNillableUserID sets the "user" edge to the User entity by ID if the given value is not nil.
func (oc *OAuthCreate) SetNillableUserID(id *uuid.UUID) *OAuthCreate {
	if id != nil {
		oc = oc.SetUserID(*id)
	}
	return oc
}

// SetUser sets the "user" edge to the User entity.
func (oc *OAuthCreate) SetUser(u *User) *OAuthCreate {
	return oc.SetUserID(u.ID)
}

// Mutation returns the OAuthMutation object of the builder.
func (oc *OAuthCreate) Mutation() *OAuthMutation {
	return oc.mutation
}

// Save creates the OAuth in the database.
func (oc *OAuthCreate) Save(ctx context.Context) (*OAuth, error) {
	oc.defaults()
	return withHooks(ctx, oc.sqlSave, oc.mutation, oc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (oc *OAuthCreate) SaveX(ctx context.Context) *OAuth {
	v, err := oc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (oc *OAuthCreate) Exec(ctx context.Context) error {
	_, err := oc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (oc *OAuthCreate) ExecX(ctx context.Context) {
	if err := oc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (oc *OAuthCreate) defaults() {
	if _, ok := oc.mutation.CreatedAt(); !ok {
		v := oauth.DefaultCreatedAt()
		oc.mutation.SetCreatedAt(v)
	}
	if _, ok := oc.mutation.UpdatedAt(); !ok {
		v := oauth.DefaultUpdatedAt()
		oc.mutation.SetUpdatedAt(v)
	}
	if _, ok := oc.mutation.ID(); !ok {
		v := oauth.DefaultID()
		oc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (oc *OAuthCreate) check() error {
	if _, ok := oc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "OAuth.created_at"`)}
	}
	if _, ok := oc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "OAuth.updated_at"`)}
	}
	if _, ok := oc.mutation.Provider(); !ok {
		return &ValidationError{Name: "provider", err: errors.New(`ent: missing required field "OAuth.provider"`)}
	}
	if v, ok := oc.mutation.Provider(); ok {
		if err := oauth.ProviderValidator(v); err != nil {
			return &ValidationError{Name: "provider", err: fmt.Errorf(`ent: validator failed for field "OAuth.provider": %w`, err)}
		}
	}
	if _, ok := oc.mutation.Sub(); !ok {
		return &ValidationError{Name: "sub", err: errors.New(`ent: missing required field "OAuth.sub"`)}
	}
	if v, ok := oc.mutation.Sub(); ok {
		if err := oauth.SubValidator(v); err != nil {
			return &ValidationError{Name: "sub", err: fmt.Errorf(`ent: validator failed for field "OAuth.sub": %w`, err)}
		}
	}
	return nil
}

func (oc *OAuthCreate) sqlSave(ctx context.Context) (*OAuth, error) {
	if err := oc.check(); err != nil {
		return nil, err
	}
	_node, _spec := oc.createSpec()
	if err := sqlgraph.CreateNode(ctx, oc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	oc.mutation.id = &_node.ID
	oc.mutation.done = true
	return _node, nil
}

func (oc *OAuthCreate) createSpec() (*OAuth, *sqlgraph.CreateSpec) {
	var (
		_node = &OAuth{config: oc.config}
		_spec = sqlgraph.NewCreateSpec(oauth.Table, sqlgraph.NewFieldSpec(oauth.FieldID, field.TypeUUID))
	)
	if id, ok := oc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := oc.mutation.CreatedAt(); ok {
		_spec.SetField(oauth.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := oc.mutation.UpdatedAt(); ok {
		_spec.SetField(oauth.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := oc.mutation.Provider(); ok {
		_spec.SetField(oauth.FieldProvider, field.TypeString, value)
		_node.Provider = value
	}
	if value, ok := oc.mutation.Sub(); ok {
		_spec.SetField(oauth.FieldSub, field.TypeString, value)
		_node.Sub = value
	}
	if nodes := oc.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   oauth.UserTable,
			Columns: []string{oauth.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.user_oauth = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OAuthCreateBulk is the builder for creating many OAuth entities in bulk.
type OAuthCreateBulk struct {
	config
	err      error
	builders []*OAuthCreate
}

// Save creates the OAuth entities in the database.
func (ocb *OAuthCreateBulk) Save(ctx context.Context) ([]*OAuth, error) {
	if ocb.err != nil {
		return nil, ocb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ocb.builders))
	nodes := make([]*OAuth, len(ocb.builders))
	mutators := make([]Mutator, len(ocb.builders))
	for i := range ocb.builders {
		func(i int, root context.Context) {
			builder := ocb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*OAuthMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, ocb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ocb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, ocb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ocb *OAuthCreateBulk) SaveX(ctx context.Context) []*OAuth {
	v, err := ocb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ocb *OAuthCreateBulk) Exec(ctx context.Context) error {
	_, err := ocb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ocb *OAuthCreateBulk) ExecX(ctx context.Context) {
	if err := ocb.Exec(ctx); err != nil {
		panic(err)
	}
}
