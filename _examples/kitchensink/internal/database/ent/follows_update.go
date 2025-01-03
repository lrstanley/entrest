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
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/follows"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/pet"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/predicate"
	"github.com/lrstanley/entrest/_examples/kitchensink/internal/database/ent/user"
)

// FollowsUpdate is the builder for updating Follows entities.
type FollowsUpdate struct {
	config
	hooks    []Hook
	mutation *FollowsMutation
}

// Where appends a list predicates to the FollowsUpdate builder.
func (fu *FollowsUpdate) Where(ps ...predicate.Follows) *FollowsUpdate {
	fu.mutation.Where(ps...)
	return fu
}

// SetFollowedAt sets the "followed_at" field.
func (fu *FollowsUpdate) SetFollowedAt(t time.Time) *FollowsUpdate {
	fu.mutation.SetFollowedAt(t)
	return fu
}

// SetNillableFollowedAt sets the "followed_at" field if the given value is not nil.
func (fu *FollowsUpdate) SetNillableFollowedAt(t *time.Time) *FollowsUpdate {
	if t != nil {
		fu.SetFollowedAt(*t)
	}
	return fu
}

// SetUserID sets the "user_id" field.
func (fu *FollowsUpdate) SetUserID(u uuid.UUID) *FollowsUpdate {
	fu.mutation.SetUserID(u)
	return fu
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (fu *FollowsUpdate) SetNillableUserID(u *uuid.UUID) *FollowsUpdate {
	if u != nil {
		fu.SetUserID(*u)
	}
	return fu
}

// SetPetID sets the "pet_id" field.
func (fu *FollowsUpdate) SetPetID(i int) *FollowsUpdate {
	fu.mutation.SetPetID(i)
	return fu
}

// SetNillablePetID sets the "pet_id" field if the given value is not nil.
func (fu *FollowsUpdate) SetNillablePetID(i *int) *FollowsUpdate {
	if i != nil {
		fu.SetPetID(*i)
	}
	return fu
}

// SetUser sets the "user" edge to the User entity.
func (fu *FollowsUpdate) SetUser(u *User) *FollowsUpdate {
	return fu.SetUserID(u.ID)
}

// SetPet sets the "pet" edge to the Pet entity.
func (fu *FollowsUpdate) SetPet(p *Pet) *FollowsUpdate {
	return fu.SetPetID(p.ID)
}

// Mutation returns the FollowsMutation object of the builder.
func (fu *FollowsUpdate) Mutation() *FollowsMutation {
	return fu.mutation
}

// ClearUser clears the "user" edge to the User entity.
func (fu *FollowsUpdate) ClearUser() *FollowsUpdate {
	fu.mutation.ClearUser()
	return fu
}

// ClearPet clears the "pet" edge to the Pet entity.
func (fu *FollowsUpdate) ClearPet() *FollowsUpdate {
	fu.mutation.ClearPet()
	return fu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (fu *FollowsUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, fu.sqlSave, fu.mutation, fu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (fu *FollowsUpdate) SaveX(ctx context.Context) int {
	affected, err := fu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (fu *FollowsUpdate) Exec(ctx context.Context) error {
	_, err := fu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fu *FollowsUpdate) ExecX(ctx context.Context) {
	if err := fu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (fu *FollowsUpdate) check() error {
	if fu.mutation.UserCleared() && len(fu.mutation.UserIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Follows.user"`)
	}
	if fu.mutation.PetCleared() && len(fu.mutation.PetIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Follows.pet"`)
	}
	return nil
}

func (fu *FollowsUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := fu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(follows.Table, follows.Columns, sqlgraph.NewFieldSpec(follows.FieldUserID, field.TypeUUID), sqlgraph.NewFieldSpec(follows.FieldPetID, field.TypeInt))
	if ps := fu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := fu.mutation.FollowedAt(); ok {
		_spec.SetField(follows.FieldFollowedAt, field.TypeTime, value)
	}
	if fu.mutation.UserCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   follows.UserTable,
			Columns: []string{follows.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fu.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   follows.UserTable,
			Columns: []string{follows.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if fu.mutation.PetCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   follows.PetTable,
			Columns: []string{follows.PetColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(pet.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fu.mutation.PetIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   follows.PetTable,
			Columns: []string{follows.PetColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(pet.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, fu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{follows.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	fu.mutation.done = true
	return n, nil
}

// FollowsUpdateOne is the builder for updating a single Follows entity.
type FollowsUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *FollowsMutation
}

// SetFollowedAt sets the "followed_at" field.
func (fuo *FollowsUpdateOne) SetFollowedAt(t time.Time) *FollowsUpdateOne {
	fuo.mutation.SetFollowedAt(t)
	return fuo
}

// SetNillableFollowedAt sets the "followed_at" field if the given value is not nil.
func (fuo *FollowsUpdateOne) SetNillableFollowedAt(t *time.Time) *FollowsUpdateOne {
	if t != nil {
		fuo.SetFollowedAt(*t)
	}
	return fuo
}

// SetUserID sets the "user_id" field.
func (fuo *FollowsUpdateOne) SetUserID(u uuid.UUID) *FollowsUpdateOne {
	fuo.mutation.SetUserID(u)
	return fuo
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (fuo *FollowsUpdateOne) SetNillableUserID(u *uuid.UUID) *FollowsUpdateOne {
	if u != nil {
		fuo.SetUserID(*u)
	}
	return fuo
}

// SetPetID sets the "pet_id" field.
func (fuo *FollowsUpdateOne) SetPetID(i int) *FollowsUpdateOne {
	fuo.mutation.SetPetID(i)
	return fuo
}

// SetNillablePetID sets the "pet_id" field if the given value is not nil.
func (fuo *FollowsUpdateOne) SetNillablePetID(i *int) *FollowsUpdateOne {
	if i != nil {
		fuo.SetPetID(*i)
	}
	return fuo
}

// SetUser sets the "user" edge to the User entity.
func (fuo *FollowsUpdateOne) SetUser(u *User) *FollowsUpdateOne {
	return fuo.SetUserID(u.ID)
}

// SetPet sets the "pet" edge to the Pet entity.
func (fuo *FollowsUpdateOne) SetPet(p *Pet) *FollowsUpdateOne {
	return fuo.SetPetID(p.ID)
}

// Mutation returns the FollowsMutation object of the builder.
func (fuo *FollowsUpdateOne) Mutation() *FollowsMutation {
	return fuo.mutation
}

// ClearUser clears the "user" edge to the User entity.
func (fuo *FollowsUpdateOne) ClearUser() *FollowsUpdateOne {
	fuo.mutation.ClearUser()
	return fuo
}

// ClearPet clears the "pet" edge to the Pet entity.
func (fuo *FollowsUpdateOne) ClearPet() *FollowsUpdateOne {
	fuo.mutation.ClearPet()
	return fuo
}

// Where appends a list predicates to the FollowsUpdate builder.
func (fuo *FollowsUpdateOne) Where(ps ...predicate.Follows) *FollowsUpdateOne {
	fuo.mutation.Where(ps...)
	return fuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (fuo *FollowsUpdateOne) Select(field string, fields ...string) *FollowsUpdateOne {
	fuo.fields = append([]string{field}, fields...)
	return fuo
}

// Save executes the query and returns the updated Follows entity.
func (fuo *FollowsUpdateOne) Save(ctx context.Context) (*Follows, error) {
	return withHooks(ctx, fuo.sqlSave, fuo.mutation, fuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (fuo *FollowsUpdateOne) SaveX(ctx context.Context) *Follows {
	node, err := fuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (fuo *FollowsUpdateOne) Exec(ctx context.Context) error {
	_, err := fuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fuo *FollowsUpdateOne) ExecX(ctx context.Context) {
	if err := fuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (fuo *FollowsUpdateOne) check() error {
	if fuo.mutation.UserCleared() && len(fuo.mutation.UserIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Follows.user"`)
	}
	if fuo.mutation.PetCleared() && len(fuo.mutation.PetIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Follows.pet"`)
	}
	return nil
}

func (fuo *FollowsUpdateOne) sqlSave(ctx context.Context) (_node *Follows, err error) {
	if err := fuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(follows.Table, follows.Columns, sqlgraph.NewFieldSpec(follows.FieldUserID, field.TypeUUID), sqlgraph.NewFieldSpec(follows.FieldPetID, field.TypeInt))
	if id, ok := fuo.mutation.UserID(); !ok {
		return nil, &ValidationError{Name: "user_id", err: errors.New(`ent: missing "Follows.user_id" for update`)}
	} else {
		_spec.Node.CompositeID[0].Value = id
	}
	if id, ok := fuo.mutation.PetID(); !ok {
		return nil, &ValidationError{Name: "pet_id", err: errors.New(`ent: missing "Follows.pet_id" for update`)}
	} else {
		_spec.Node.CompositeID[1].Value = id
	}
	if fields := fuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, len(fields))
		for i, f := range fields {
			if !follows.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			_spec.Node.Columns[i] = f
		}
	}
	if ps := fuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := fuo.mutation.FollowedAt(); ok {
		_spec.SetField(follows.FieldFollowedAt, field.TypeTime, value)
	}
	if fuo.mutation.UserCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   follows.UserTable,
			Columns: []string{follows.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fuo.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   follows.UserTable,
			Columns: []string{follows.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if fuo.mutation.PetCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   follows.PetTable,
			Columns: []string{follows.PetColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(pet.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fuo.mutation.PetIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   follows.PetTable,
			Columns: []string{follows.PetColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(pet.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Follows{config: fuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, fuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{follows.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	fuo.mutation.done = true
	return _node, nil
}
