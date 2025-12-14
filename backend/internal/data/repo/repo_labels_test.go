package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func labelFactory() LabelCreate {
	return LabelCreate{
		Name:        fk.Str(10),
		Description: fk.Str(100),
	}
}

func useLabels(t *testing.T, len int) []LabelOut {
	t.Helper()

	labels := make([]LabelOut, len)
	for i := 0; i < len; i++ {
		itm := labelFactory()

		item, err := tRepos.Labels.Create(context.Background(), tGroup.ID, itm)
		require.NoError(t, err)
		labels[i] = item
	}

	t.Cleanup(func() {
		for _, item := range labels {
			_ = tRepos.Labels.delete(context.Background(), item.ID)
		}
	})

	return labels
}

func TestLabelRepository_Get(t *testing.T) {
	labels := useLabels(t, 1)
	label := labels[0]

	// Get by ID
	foundLoc, err := tRepos.Labels.GetOne(context.Background(), label.ID)
	require.NoError(t, err)
	assert.Equal(t, label.ID, foundLoc.ID)
}

func TestLabelRepositoryGetAll(t *testing.T) {
	useLabels(t, 10)

	all, err := tRepos.Labels.GetAll(context.Background(), tGroup.ID)
	require.NoError(t, err)
	assert.Len(t, all, 10)
}

func TestLabelRepository_Create(t *testing.T) {
	loc, err := tRepos.Labels.Create(context.Background(), tGroup.ID, labelFactory())
	require.NoError(t, err)

	// Get by ID
	foundLoc, err := tRepos.Labels.GetOne(context.Background(), loc.ID)
	require.NoError(t, err)
	assert.Equal(t, loc.ID, foundLoc.ID)

	err = tRepos.Labels.delete(context.Background(), loc.ID)
	require.NoError(t, err)
}

func TestLabelRepository_Update(t *testing.T) {
	loc, err := tRepos.Labels.Create(context.Background(), tGroup.ID, labelFactory())
	require.NoError(t, err)

	updateData := LabelUpdate{
		ID:          loc.ID,
		Name:        fk.Str(10),
		Description: fk.Str(100),
	}

	update, err := tRepos.Labels.UpdateByGroup(context.Background(), tGroup.ID, updateData)
	require.NoError(t, err)

	foundLoc, err := tRepos.Labels.GetOne(context.Background(), loc.ID)
	require.NoError(t, err)

	assert.Equal(t, update.ID, foundLoc.ID)
	assert.Equal(t, update.Name, foundLoc.Name)
	assert.Equal(t, update.Description, foundLoc.Description)

	err = tRepos.Labels.delete(context.Background(), loc.ID)
	require.NoError(t, err)
}

func TestLabelRepository_Delete(t *testing.T) {
	loc, err := tRepos.Labels.Create(context.Background(), tGroup.ID, labelFactory())
	require.NoError(t, err)

	err = tRepos.Labels.delete(context.Background(), loc.ID)
	require.NoError(t, err)

	_, err = tRepos.Labels.GetOne(context.Background(), loc.ID)
	require.Error(t, err)
}

func TestLabelRepository_ParentChild_Create(t *testing.T) {
	ctx := context.Background()

	// Create parent label
	parent, err := tRepos.Labels.Create(ctx, tGroup.ID, labelFactory())
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, parent.ID)

	// Create child label
	childData := labelFactory()
	childData.ParentID = parent.ID
	child, err := tRepos.Labels.Create(ctx, tGroup.ID, childData)
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, child.ID)

	// Verify child has parent
	foundChild, err := tRepos.Labels.GetOne(ctx, child.ID)
	require.NoError(t, err)
	require.NotNil(t, foundChild.Parent)
	assert.Equal(t, parent.ID, foundChild.Parent.ID)

	// Verify parent has child
	foundParent, err := tRepos.Labels.GetOne(ctx, parent.ID)
	require.NoError(t, err)
	assert.Len(t, foundParent.Children, 1)
	assert.Equal(t, child.ID, foundParent.Children[0].ID)
}

func TestLabelRepository_ParentChild_CascadeDelete(t *testing.T) {
	ctx := context.Background()

	// Create parent label
	parent, err := tRepos.Labels.Create(ctx, tGroup.ID, labelFactory())
	require.NoError(t, err)

	// Create child labels
	child1Data := labelFactory()
	child1Data.ParentID = parent.ID
	child1, err := tRepos.Labels.Create(ctx, tGroup.ID, child1Data)
	require.NoError(t, err)

	child2Data := labelFactory()
	child2Data.ParentID = parent.ID
	child2, err := tRepos.Labels.Create(ctx, tGroup.ID, child2Data)
	require.NoError(t, err)

	// Delete parent
	err = tRepos.Labels.delete(ctx, parent.ID)
	require.NoError(t, err)

	// Verify children are also deleted (cascade)
	_, err = tRepos.Labels.GetOne(ctx, child1.ID)
	require.Error(t, err)

	_, err = tRepos.Labels.GetOne(ctx, child2.ID)
	require.Error(t, err)
}

func TestLabelRepository_ParentChild_CircularReference_Self(t *testing.T) {
	ctx := context.Background()

	// Create a label
	label1, err := tRepos.Labels.Create(ctx, tGroup.ID, labelFactory())
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, label1.ID)

	// Try to set itself as parent - should fail
	updateData := LabelUpdate{
		ID:          label1.ID,
		Name:        label1.Name,
		Description: label1.Description,
		Color:       label1.Color,
		ParentID:    label1.ID,
	}

	_, err = tRepos.Labels.UpdateByGroup(ctx, tGroup.ID, updateData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be its own parent")
}

func TestLabelRepository_ParentChild_CircularReference_Chain(t *testing.T) {
	ctx := context.Background()

	// Create label chain: label1 -> label2 -> label3
	label1, err := tRepos.Labels.Create(ctx, tGroup.ID, labelFactory())
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, label1.ID)

	label2Data := labelFactory()
	label2Data.ParentID = label1.ID
	label2, err := tRepos.Labels.Create(ctx, tGroup.ID, label2Data)
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, label2.ID)

	label3Data := labelFactory()
	label3Data.ParentID = label2.ID
	label3, err := tRepos.Labels.Create(ctx, tGroup.ID, label3Data)
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, label3.ID)

	// Try to set label3 as parent of label1 (would create circular reference)
	updateData := LabelUpdate{
		ID:          label1.ID,
		Name:        label1.Name,
		Description: label1.Description,
		Color:       label1.Color,
		ParentID:    label3.ID,
	}

	_, err = tRepos.Labels.UpdateByGroup(ctx, tGroup.ID, updateData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular reference")
}

func TestLabelRepository_ParentChild_MaxDepth(t *testing.T) {
	ctx := context.Background()

	// Create a chain of 20 labels
	labels := make([]LabelOut, 20)
	for i := 0; i < 20; i++ {
		labelData := labelFactory()
		if i > 0 {
			labelData.ParentID = labels[i-1].ID
		}
		label, err := tRepos.Labels.Create(ctx, tGroup.ID, labelData)
		require.NoError(t, err)
		labels[i] = label
	}

	// Clean up
	defer func() {
		for i := len(labels) - 1; i >= 0; i-- {
			_ = tRepos.Labels.delete(ctx, labels[i].ID)
		}
	}()

	// Try to create 21st level - should fail
	labelData := labelFactory()
	labelData.ParentID = labels[19].ID
	_, err := tRepos.Labels.Create(ctx, tGroup.ID, labelData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum depth")
}

func TestLabelRepository_ParentChild_Update(t *testing.T) {
	ctx := context.Background()

	// Create labels
	parent1, err := tRepos.Labels.Create(ctx, tGroup.ID, labelFactory())
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, parent1.ID)

	parent2, err := tRepos.Labels.Create(ctx, tGroup.ID, labelFactory())
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, parent2.ID)

	childData := labelFactory()
	childData.ParentID = parent1.ID
	child, err := tRepos.Labels.Create(ctx, tGroup.ID, childData)
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, child.ID)

	// Update child to have different parent
	updateData := LabelUpdate{
		ID:          child.ID,
		Name:        child.Name,
		Description: child.Description,
		Color:       child.Color,
		ParentID:    parent2.ID,
	}

	updated, err := tRepos.Labels.UpdateByGroup(ctx, tGroup.ID, updateData)
	require.NoError(t, err)
	require.NotNil(t, updated.Parent)
	assert.Equal(t, parent2.ID, updated.Parent.ID)
}

func TestLabelRepository_ParentChild_ClearParent(t *testing.T) {
	ctx := context.Background()

	// Create parent and child
	parent, err := tRepos.Labels.Create(ctx, tGroup.ID, labelFactory())
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, parent.ID)

	childData := labelFactory()
	childData.ParentID = parent.ID
	child, err := tRepos.Labels.Create(ctx, tGroup.ID, childData)
	require.NoError(t, err)
	defer tRepos.Labels.delete(ctx, child.ID)

	// Clear parent by setting ParentID to uuid.Nil
	updateData := LabelUpdate{
		ID:          child.ID,
		Name:        child.Name,
		Description: child.Description,
		Color:       child.Color,
		ParentID:    uuid.Nil,
	}

	updated, err := tRepos.Labels.UpdateByGroup(ctx, tGroup.ID, updateData)
	require.NoError(t, err)
	assert.Nil(t, updated.Parent)
}
