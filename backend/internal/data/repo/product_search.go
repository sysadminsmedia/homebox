package repo

import (
	"context"

	"github.com/google/uuid"
)

func (e *ItemsRepository) FetchProductInfoFromEAN(ctx context.Context, gid uuid.UUID) (int, error) {
	// All items where there is no primary photo
	/*itemIDs, err := e.db.Item.Query().
		Where(
			item.HasGroupWith(group.ID(gid)),
			item.HasAttachmentsWith(
				attachment.TypeEQ(attachment.TypePhoto),
				attachment.Not(
					attachment.And(
						attachment.Primary(true),
						attachment.TypeEQ(attachment.TypePhoto),
					),
				),
			),
		).
		IDs(ctx)
	if err != nil {
		return -1, err
	}

	updated := 0
	for _, id := range itemIDs {
		// Find the first photo attachment
		a, err := e.db.Attachment.Query().
			Where(
				attachment.HasItemWith(item.ID(id)),
				attachment.TypeEQ(attachment.TypePhoto),
				attachment.Primary(false),
			).
			First(ctx)
		if err != nil {
			return updated, err
		}

		// Set it as primary
		_, err = e.db.Attachment.UpdateOne(a).
			SetPrimary(true).
			Save(ctx)
		if err != nil {
			return updated, err
		}

		updated++
	}

	return updated, nil*/
	return 200, nil
}
