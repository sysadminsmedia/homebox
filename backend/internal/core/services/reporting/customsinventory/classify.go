package customsinventory

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/sysadminsmedia/homebox/backend/internal/data/types"
)

// Rules are the customs rules for one export.
type Rules struct {
	Reference, Cutoff  types.Date
	MinMonths          int                    // after applying the default
	DutyFree, Dutiable map[uuid.UUID]struct{} // already expanded
	Tags               TagIndex
}

type Classification struct {
	Status    Status
	Reason    Reason
	ReasonTag string      // name of the item's own matching tag (alphabetically first), for tag_* reasons
	Issues    []IssueCode // in IssueOrder
}

// Classify applies the customs rules (FR-9) in their fixed order.
func Classify(it Item, r Rules) Classification {
	dutiableTag := firstTagName(it.TagIDs, r.Dutiable, r.Tags)
	dutyFreeTag := firstTagName(it.TagIDs, r.DutyFree, r.Tags)
	hasDutyFreeTag := matchesAny(it.TagIDs, r.DutyFree)
	purchase := it.PurchaseDate.Time()

	switch {
	case matchesAny(it.TagIDs, r.Dutiable):
		c := Classification{Status: StatusDutiable, Reason: ReasonTagDutiable, ReasonTag: dutiableTag}
		if hasDutyFreeTag {
			c.Issues = []IssueCode{IssueConflictTags}
		}
		return c
	case !purchase.IsZero() && purchase.After(r.Reference.Time()):
		return Classification{
			Status: StatusDutiable, Reason: ReasonPurchasedAfterReference,
			Issues: []IssueCode{IssuePurchasedAfterReference},
		}
	case !purchase.IsZero() && purchase.After(r.Cutoff.Time()):
		c := Classification{Status: StatusDutiable, Reason: ReasonOwnedTooShort}
		if hasDutyFreeTag {
			c.Issues = []IssueCode{IssueConflictDateTag}
		}
		return c
	case !purchase.IsZero():
		return Classification{Status: StatusDutyFree, Reason: ReasonDate}
	case hasDutyFreeTag:
		return Classification{Status: StatusDutyFree, Reason: ReasonTagDutyFree, ReasonTag: dutyFreeTag}
	default:
		return Classification{
			Status: StatusDutiable, Reason: ReasonMissingPurchaseDate,
			Issues: []IssueCode{IssueMissingPurchaseDate},
		}
	}
}

// firstTagName returns the alphabetically first name among the item's own
// tags that are in set.
func firstTagName(itemTags []uuid.UUID, set map[uuid.UUID]struct{}, ti TagIndex) string {
	var names []string
	for _, id := range itemTags {
		if _, ok := set[id]; ok {
			names = append(names, ti.Name(id))
		}
	}
	if len(names) == 0 {
		return ""
	}
	return slices.Min(names)
}

// ReasonComment is the comment text for a classification (FR-10), without
// the "; Wert fehlt" suffix.
func ReasonComment(c Classification, minMonths int) string {
	switch c.Reason {
	case ReasonOwnedTooShort:
		return fmt.Sprintf("Kauf nach Stichtag (< %d Monate)", minMonths)
	case ReasonMissingPurchaseDate:
		return "Kaufdatum fehlt"
	case ReasonTagDutiable:
		return "zu verzollen lt. Tag „" + c.ReasonTag + "“"
	case ReasonPurchasedAfterReference:
		return "Kauf nach Referenzdatum"
	case ReasonTagDutyFree:
		return "zollfrei lt. Tag „" + c.ReasonTag + "“"
	default:
		return ""
	}
}
