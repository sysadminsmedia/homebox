// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/groupinvitationtoken"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/predicate"
)

// GroupInvitationTokenUpdate is the builder for updating GroupInvitationToken entities.
type GroupInvitationTokenUpdate struct {
	config
	hooks    []Hook
	mutation *GroupInvitationTokenMutation
}

// Where appends a list predicates to the GroupInvitationTokenUpdate builder.
func (gitu *GroupInvitationTokenUpdate) Where(ps ...predicate.GroupInvitationToken) *GroupInvitationTokenUpdate {
	gitu.mutation.Where(ps...)
	return gitu
}

// SetUpdatedAt sets the "updated_at" field.
func (gitu *GroupInvitationTokenUpdate) SetUpdatedAt(t time.Time) *GroupInvitationTokenUpdate {
	gitu.mutation.SetUpdatedAt(t)
	return gitu
}

// SetToken sets the "token" field.
func (gitu *GroupInvitationTokenUpdate) SetToken(b []byte) *GroupInvitationTokenUpdate {
	gitu.mutation.SetToken(b)
	return gitu
}

// SetExpiresAt sets the "expires_at" field.
func (gitu *GroupInvitationTokenUpdate) SetExpiresAt(t time.Time) *GroupInvitationTokenUpdate {
	gitu.mutation.SetExpiresAt(t)
	return gitu
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (gitu *GroupInvitationTokenUpdate) SetNillableExpiresAt(t *time.Time) *GroupInvitationTokenUpdate {
	if t != nil {
		gitu.SetExpiresAt(*t)
	}
	return gitu
}

// SetUses sets the "uses" field.
func (gitu *GroupInvitationTokenUpdate) SetUses(i int) *GroupInvitationTokenUpdate {
	gitu.mutation.ResetUses()
	gitu.mutation.SetUses(i)
	return gitu
}

// SetNillableUses sets the "uses" field if the given value is not nil.
func (gitu *GroupInvitationTokenUpdate) SetNillableUses(i *int) *GroupInvitationTokenUpdate {
	if i != nil {
		gitu.SetUses(*i)
	}
	return gitu
}

// AddUses adds i to the "uses" field.
func (gitu *GroupInvitationTokenUpdate) AddUses(i int) *GroupInvitationTokenUpdate {
	gitu.mutation.AddUses(i)
	return gitu
}

// SetGroupID sets the "group" edge to the Group entity by ID.
func (gitu *GroupInvitationTokenUpdate) SetGroupID(id uuid.UUID) *GroupInvitationTokenUpdate {
	gitu.mutation.SetGroupID(id)
	return gitu
}

// SetNillableGroupID sets the "group" edge to the Group entity by ID if the given value is not nil.
func (gitu *GroupInvitationTokenUpdate) SetNillableGroupID(id *uuid.UUID) *GroupInvitationTokenUpdate {
	if id != nil {
		gitu = gitu.SetGroupID(*id)
	}
	return gitu
}

// SetGroup sets the "group" edge to the Group entity.
func (gitu *GroupInvitationTokenUpdate) SetGroup(g *Group) *GroupInvitationTokenUpdate {
	return gitu.SetGroupID(g.ID)
}

// Mutation returns the GroupInvitationTokenMutation object of the builder.
func (gitu *GroupInvitationTokenUpdate) Mutation() *GroupInvitationTokenMutation {
	return gitu.mutation
}

// ClearGroup clears the "group" edge to the Group entity.
func (gitu *GroupInvitationTokenUpdate) ClearGroup() *GroupInvitationTokenUpdate {
	gitu.mutation.ClearGroup()
	return gitu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (gitu *GroupInvitationTokenUpdate) Save(ctx context.Context) (int, error) {
	gitu.defaults()
	return withHooks(ctx, gitu.sqlSave, gitu.mutation, gitu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (gitu *GroupInvitationTokenUpdate) SaveX(ctx context.Context) int {
	affected, err := gitu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (gitu *GroupInvitationTokenUpdate) Exec(ctx context.Context) error {
	_, err := gitu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gitu *GroupInvitationTokenUpdate) ExecX(ctx context.Context) {
	if err := gitu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (gitu *GroupInvitationTokenUpdate) defaults() {
	if _, ok := gitu.mutation.UpdatedAt(); !ok {
		v := groupinvitationtoken.UpdateDefaultUpdatedAt()
		gitu.mutation.SetUpdatedAt(v)
	}
}

func (gitu *GroupInvitationTokenUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(groupinvitationtoken.Table, groupinvitationtoken.Columns, sqlgraph.NewFieldSpec(groupinvitationtoken.FieldID, field.TypeUUID))
	if ps := gitu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := gitu.mutation.UpdatedAt(); ok {
		_spec.SetField(groupinvitationtoken.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := gitu.mutation.Token(); ok {
		_spec.SetField(groupinvitationtoken.FieldToken, field.TypeBytes, value)
	}
	if value, ok := gitu.mutation.ExpiresAt(); ok {
		_spec.SetField(groupinvitationtoken.FieldExpiresAt, field.TypeTime, value)
	}
	if value, ok := gitu.mutation.Uses(); ok {
		_spec.SetField(groupinvitationtoken.FieldUses, field.TypeInt, value)
	}
	if value, ok := gitu.mutation.AddedUses(); ok {
		_spec.AddField(groupinvitationtoken.FieldUses, field.TypeInt, value)
	}
	if gitu.mutation.GroupCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   groupinvitationtoken.GroupTable,
			Columns: []string{groupinvitationtoken.GroupColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(group.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gitu.mutation.GroupIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   groupinvitationtoken.GroupTable,
			Columns: []string{groupinvitationtoken.GroupColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(group.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, gitu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{groupinvitationtoken.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	gitu.mutation.done = true
	return n, nil
}

// GroupInvitationTokenUpdateOne is the builder for updating a single GroupInvitationToken entity.
type GroupInvitationTokenUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *GroupInvitationTokenMutation
}

// SetUpdatedAt sets the "updated_at" field.
func (gituo *GroupInvitationTokenUpdateOne) SetUpdatedAt(t time.Time) *GroupInvitationTokenUpdateOne {
	gituo.mutation.SetUpdatedAt(t)
	return gituo
}

// SetToken sets the "token" field.
func (gituo *GroupInvitationTokenUpdateOne) SetToken(b []byte) *GroupInvitationTokenUpdateOne {
	gituo.mutation.SetToken(b)
	return gituo
}

// SetExpiresAt sets the "expires_at" field.
func (gituo *GroupInvitationTokenUpdateOne) SetExpiresAt(t time.Time) *GroupInvitationTokenUpdateOne {
	gituo.mutation.SetExpiresAt(t)
	return gituo
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (gituo *GroupInvitationTokenUpdateOne) SetNillableExpiresAt(t *time.Time) *GroupInvitationTokenUpdateOne {
	if t != nil {
		gituo.SetExpiresAt(*t)
	}
	return gituo
}

// SetUses sets the "uses" field.
func (gituo *GroupInvitationTokenUpdateOne) SetUses(i int) *GroupInvitationTokenUpdateOne {
	gituo.mutation.ResetUses()
	gituo.mutation.SetUses(i)
	return gituo
}

// SetNillableUses sets the "uses" field if the given value is not nil.
func (gituo *GroupInvitationTokenUpdateOne) SetNillableUses(i *int) *GroupInvitationTokenUpdateOne {
	if i != nil {
		gituo.SetUses(*i)
	}
	return gituo
}

// AddUses adds i to the "uses" field.
func (gituo *GroupInvitationTokenUpdateOne) AddUses(i int) *GroupInvitationTokenUpdateOne {
	gituo.mutation.AddUses(i)
	return gituo
}

// SetGroupID sets the "group" edge to the Group entity by ID.
func (gituo *GroupInvitationTokenUpdateOne) SetGroupID(id uuid.UUID) *GroupInvitationTokenUpdateOne {
	gituo.mutation.SetGroupID(id)
	return gituo
}

// SetNillableGroupID sets the "group" edge to the Group entity by ID if the given value is not nil.
func (gituo *GroupInvitationTokenUpdateOne) SetNillableGroupID(id *uuid.UUID) *GroupInvitationTokenUpdateOne {
	if id != nil {
		gituo = gituo.SetGroupID(*id)
	}
	return gituo
}

// SetGroup sets the "group" edge to the Group entity.
func (gituo *GroupInvitationTokenUpdateOne) SetGroup(g *Group) *GroupInvitationTokenUpdateOne {
	return gituo.SetGroupID(g.ID)
}

// Mutation returns the GroupInvitationTokenMutation object of the builder.
func (gituo *GroupInvitationTokenUpdateOne) Mutation() *GroupInvitationTokenMutation {
	return gituo.mutation
}

// ClearGroup clears the "group" edge to the Group entity.
func (gituo *GroupInvitationTokenUpdateOne) ClearGroup() *GroupInvitationTokenUpdateOne {
	gituo.mutation.ClearGroup()
	return gituo
}

// Where appends a list predicates to the GroupInvitationTokenUpdate builder.
func (gituo *GroupInvitationTokenUpdateOne) Where(ps ...predicate.GroupInvitationToken) *GroupInvitationTokenUpdateOne {
	gituo.mutation.Where(ps...)
	return gituo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (gituo *GroupInvitationTokenUpdateOne) Select(field string, fields ...string) *GroupInvitationTokenUpdateOne {
	gituo.fields = append([]string{field}, fields...)
	return gituo
}

// Save executes the query and returns the updated GroupInvitationToken entity.
func (gituo *GroupInvitationTokenUpdateOne) Save(ctx context.Context) (*GroupInvitationToken, error) {
	gituo.defaults()
	return withHooks(ctx, gituo.sqlSave, gituo.mutation, gituo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (gituo *GroupInvitationTokenUpdateOne) SaveX(ctx context.Context) *GroupInvitationToken {
	node, err := gituo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (gituo *GroupInvitationTokenUpdateOne) Exec(ctx context.Context) error {
	_, err := gituo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gituo *GroupInvitationTokenUpdateOne) ExecX(ctx context.Context) {
	if err := gituo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (gituo *GroupInvitationTokenUpdateOne) defaults() {
	if _, ok := gituo.mutation.UpdatedAt(); !ok {
		v := groupinvitationtoken.UpdateDefaultUpdatedAt()
		gituo.mutation.SetUpdatedAt(v)
	}
}

func (gituo *GroupInvitationTokenUpdateOne) sqlSave(ctx context.Context) (_node *GroupInvitationToken, err error) {
	_spec := sqlgraph.NewUpdateSpec(groupinvitationtoken.Table, groupinvitationtoken.Columns, sqlgraph.NewFieldSpec(groupinvitationtoken.FieldID, field.TypeUUID))
	id, ok := gituo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "GroupInvitationToken.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := gituo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, groupinvitationtoken.FieldID)
		for _, f := range fields {
			if !groupinvitationtoken.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != groupinvitationtoken.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := gituo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := gituo.mutation.UpdatedAt(); ok {
		_spec.SetField(groupinvitationtoken.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := gituo.mutation.Token(); ok {
		_spec.SetField(groupinvitationtoken.FieldToken, field.TypeBytes, value)
	}
	if value, ok := gituo.mutation.ExpiresAt(); ok {
		_spec.SetField(groupinvitationtoken.FieldExpiresAt, field.TypeTime, value)
	}
	if value, ok := gituo.mutation.Uses(); ok {
		_spec.SetField(groupinvitationtoken.FieldUses, field.TypeInt, value)
	}
	if value, ok := gituo.mutation.AddedUses(); ok {
		_spec.AddField(groupinvitationtoken.FieldUses, field.TypeInt, value)
	}
	if gituo.mutation.GroupCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   groupinvitationtoken.GroupTable,
			Columns: []string{groupinvitationtoken.GroupColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(group.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gituo.mutation.GroupIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   groupinvitationtoken.GroupTable,
			Columns: []string{groupinvitationtoken.GroupColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(group.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &GroupInvitationToken{config: gituo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, gituo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{groupinvitationtoken.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	gituo.mutation.done = true
	return _node, nil
}
