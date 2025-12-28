package repo

import (
	"context"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/group"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/groupinvitationtoken"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/item"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/label"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent/location"
)

type GroupRepository struct {
	db               *ent.Client
	groupMapper      MapFunc[*ent.Group, Group]
	invitationMapper MapFunc[*ent.GroupInvitationToken, GroupInvitation]
}

func NewGroupRepository(db *ent.Client) *GroupRepository {
	gmap := func(g *ent.Group) Group {
		return Group{
			ID:        g.ID,
			Name:      g.Name,
			CreatedAt: g.CreatedAt,
			UpdatedAt: g.UpdatedAt,
			Currency:  strings.ToUpper(g.Currency),
		}
	}

	imap := func(i *ent.GroupInvitationToken) GroupInvitation {
		return GroupInvitation{
			ID:        i.ID,
			ExpiresAt: i.ExpiresAt,
			Uses:      i.Uses,
			Group:     gmap(i.Edges.Group),
		}
	}

	return &GroupRepository{
		db:               db,
		groupMapper:      gmap,
		invitationMapper: imap,
	}
}

type (
	Group struct {
		ID        uuid.UUID `json:"id,omitempty"`
		Name      string    `json:"name,omitempty"`
		CreatedAt time.Time `json:"createdAt,omitempty"`
		UpdatedAt time.Time `json:"updatedAt,omitempty"`
		Currency  string    `json:"currency,omitempty"`
	}

	GroupUpdate struct {
		Name     string `json:"name"`
		Currency string `json:"currency"`
	}

	GroupInvitationCreate struct {
		Token     []byte    `json:"-"`
		ExpiresAt time.Time `json:"expiresAt"`
		Uses      int       `json:"uses"`
	}

	GroupInvitation struct {
		ID        uuid.UUID `json:"id"`
		ExpiresAt time.Time `json:"expiresAt"`
		Uses      int       `json:"uses"`
		Group     Group     `json:"group"`
	}

	GroupStatistics struct {
		TotalUsers        int     `json:"totalUsers"`
		TotalItems        int     `json:"totalItems"`
		TotalLocations    int     `json:"totalLocations"`
		TotalLabels       int     `json:"totalLabels"`
		TotalItemPrice    float64 `json:"totalItemPrice"`
		TotalWithWarranty int     `json:"totalWithWarranty"`
	}

	ValueOverTimeEntry struct {
		Date  time.Time `json:"date"`
		Value float64   `json:"value"`
		Name  string    `json:"name"`
	}

	ValueOverTime struct {
		PriceAtStart float64              `json:"valueAtStart"`
		PriceAtEnd   float64              `json:"valueAtEnd"`
		Start        time.Time            `json:"start"`
		End          time.Time            `json:"end"`
		Entries      []ValueOverTimeEntry `json:"entries"`
	}

	TotalsByOrganizer struct {
		ID    uuid.UUID `json:"id"`
		Name  string    `json:"name"`
		Total float64   `json:"total"`
	}
)

func (r *GroupRepository) GetAllGroups(ctx context.Context) ([]Group, error) {
	return r.groupMapper.MapEachErr(r.db.Group.Query().All(ctx))
}

func (r *GroupRepository) StatsLocationsByPurchasePrice(ctx context.Context, gid uuid.UUID) ([]TotalsByOrganizer, error) {
	var v []TotalsByOrganizer

	err := r.db.Location.Query().
		Where(
			location.HasGroupWith(group.ID(gid)),
		).
		GroupBy(location.FieldID, location.FieldName).
		Aggregate(func(sq *sql.Selector) string {
			t := sql.Table(item.Table)
			sq.Join(t).On(sq.C(location.FieldID), t.C(item.LocationColumn))

			return sql.As(sql.Sum(t.C(item.FieldPurchasePrice)), "total")
		}).
		Scan(ctx, &v)
	if err != nil {
		return nil, err
	}

	return v, err
}

func (r *GroupRepository) StatsLabelsByPurchasePrice(ctx context.Context, gid uuid.UUID) ([]TotalsByOrganizer, error) {
	var v []TotalsByOrganizer

	err := r.db.Label.Query().
		Where(
			label.HasGroupWith(group.ID(gid)),
		).
		GroupBy(label.FieldID, label.FieldName).
		Aggregate(func(sq *sql.Selector) string {
			itemTable := sql.Table(item.Table)

			jt := sql.Table(label.ItemsTable)

			sq.Join(jt).On(sq.C(label.FieldID), jt.C(label.ItemsPrimaryKey[0]))
			sq.Join(itemTable).On(jt.C(label.ItemsPrimaryKey[1]), itemTable.C(item.FieldID))

			return sql.As(sql.Sum(itemTable.C(item.FieldPurchasePrice)), "total")
		}).
		Scan(ctx, &v)
	if err != nil {
		return nil, err
	}

	return v, err
}

func (r *GroupRepository) StatsPurchasePrice(ctx context.Context, gid uuid.UUID, start, end time.Time) (*ValueOverTime, error) {
	// Get the Totals for the Start and End of the Given Time Period
	q := `
	SELECT
		SUM(CASE WHEN created_at < $1 THEN purchase_price ELSE 0 END) AS price_at_start,
		SUM(CASE WHEN created_at < $2 THEN purchase_price ELSE 0 END) AS price_at_end
	FROM items
	WHERE group_items = $3 AND archived = false
`
	stats := ValueOverTime{
		Start: start,
		End:   end,
	}

	var maybeStart *float64
	var maybeEnd *float64

	row := r.db.Sql().QueryRowContext(ctx, q, sqliteDateFormat(start), sqliteDateFormat(end), gid)
	err := row.Scan(&maybeStart, &maybeEnd)
	if err != nil {
		return nil, err
	}

	stats.PriceAtStart = orDefault(maybeStart, 0)
	stats.PriceAtEnd = orDefault(maybeEnd, 0)

	var v []struct {
		Name          string    `json:"name"`
		CreatedAt     time.Time `json:"created_at"`
		PurchasePrice float64   `json:"purchase_price"`
	}

	// Get Created Date and Price of all items between start and end
	err = r.db.Item.Query().
		Where(
			item.HasGroupWith(group.ID(gid)),
			item.CreatedAtGTE(start),
			item.CreatedAtLTE(end),
			item.Archived(false),
		).
		Select(
			item.FieldName,
			item.FieldCreatedAt,
			item.FieldPurchasePrice,
		).
		Scan(ctx, &v)

	if err != nil {
		return nil, err
	}

	stats.Entries = make([]ValueOverTimeEntry, len(v))

	for i, vv := range v {
		stats.Entries[i] = ValueOverTimeEntry{
			Date:  vv.CreatedAt,
			Value: vv.PurchasePrice,
		}
	}

	return &stats, nil
}

func (r *GroupRepository) StatsGroup(ctx context.Context, gid uuid.UUID) (GroupStatistics, error) {
	q := `
		SELECT
            (SELECT COUNT(*) FROM user_groups WHERE group_id = $2) AS total_users,
            (SELECT COUNT(*) FROM items WHERE group_items = $2 AND items.archived = false) AS total_items,
            (SELECT COUNT(*) FROM locations WHERE group_locations = $2) AS total_locations,
            (SELECT COUNT(*) FROM labels WHERE group_labels = $2) AS total_labels,
            (SELECT SUM(purchase_price*quantity) FROM items WHERE group_items = $2 AND items.archived = false) AS total_item_price,
            (SELECT COUNT(*)
                FROM items
                    WHERE group_items = $2
                    AND items.archived = false
                    AND (items.lifetime_warranty = true OR items.warranty_expires > $1)
                ) AS total_with_warranty;
`
	var stats GroupStatistics
	row := r.db.Sql().QueryRowContext(ctx, q, sqliteDateFormat(time.Now()), gid)

	var maybeTotalItemPrice *float64
	var maybeTotalWithWarranty *int

	err := row.Scan(&stats.TotalUsers, &stats.TotalItems, &stats.TotalLocations, &stats.TotalLabels, &maybeTotalItemPrice, &maybeTotalWithWarranty)
	if err != nil {
		return GroupStatistics{}, err
	}

	stats.TotalItemPrice = orDefault(maybeTotalItemPrice, 0)
	stats.TotalWithWarranty = orDefault(maybeTotalWithWarranty, 0)

	return stats, nil
}

func (r *GroupRepository) GroupCreate(ctx context.Context, name string) (Group, error) {
	return r.groupMapper.MapErr(r.db.Group.Create().
		SetName(name).
		Save(ctx))
}

func (r *GroupRepository) GroupUpdate(ctx context.Context, id uuid.UUID, data GroupUpdate) (Group, error) {
	entity, err := r.db.Group.UpdateOneID(id).
		SetName(data.Name).
		SetCurrency(strings.ToLower(data.Currency)).
		Save(ctx)

	return r.groupMapper.MapErr(entity, err)
}

func (r *GroupRepository) GroupByID(ctx context.Context, id uuid.UUID) (Group, error) {
	return r.groupMapper.MapErr(r.db.Group.Get(ctx, id))
}

func (r *GroupRepository) InvitationGet(ctx context.Context, token []byte) (GroupInvitation, error) {
	return r.invitationMapper.MapErr(r.db.GroupInvitationToken.Query().
		Where(groupinvitationtoken.Token(token)).
		WithGroup().
		Only(ctx))
}

func (r *GroupRepository) InvitationCreate(ctx context.Context, groupID uuid.UUID, invite GroupInvitationCreate) (GroupInvitation, error) {
	entity, err := r.db.GroupInvitationToken.Create().
		SetGroupID(groupID).
		SetToken(invite.Token).
		SetExpiresAt(invite.ExpiresAt).
		SetUses(invite.Uses).
		Save(ctx)
	if err != nil {
		return GroupInvitation{}, err
	}

	return r.InvitationGet(ctx, entity.Token)
}

func (r *GroupRepository) InvitationUpdate(ctx context.Context, id uuid.UUID, uses int) error {
	_, err := r.db.GroupInvitationToken.UpdateOneID(id).SetUses(uses).Save(ctx)
	return err
}

// InvitationPurge removes all expired invitations or those that have been used up.
// It returns the number of deleted invitations.
func (r *GroupRepository) InvitationPurge(ctx context.Context) (amount int, err error) {
	q := r.db.GroupInvitationToken.Delete()
	q.Where(groupinvitationtoken.Or(
		groupinvitationtoken.ExpiresAtLT(time.Now()),
		groupinvitationtoken.UsesLTE(0),
	))

	return q.Exec(ctx)
}
