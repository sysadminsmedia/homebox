package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
)

func TestMergeTemplateAndImportedFields(t *testing.T) {
	templateFields := []repo.EntityFieldData{
		{Type: "text", Name: "Template Field", TextValue: "Template Value"},
		{Type: "text", Name: "Country", TextValue: "Template Country"},
	}
	importedFields := []repo.EntityFieldData{
		{Type: "text", Name: "Discogs Release ID", TextValue: "28577464"},
		{Type: "text", Name: "Country", TextValue: "Europe"},
	}

	fields := mergeTemplateAndImportedFields(templateFields, importedFields)

	assert.Equal(t, []repo.EntityFieldData{
		{Type: "text", Name: "Template Field", TextValue: "Template Value"},
		{Type: "text", Name: "Country", TextValue: "Europe"},
		{Type: "text", Name: "Discogs Release ID", TextValue: "28577464"},
	}, fields)
	assert.Equal(t, "Template Country", templateFields[1].TextValue, "merge must not mutate the template slice")
}
