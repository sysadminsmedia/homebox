package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate --feature sql/versioned-migration,privacy,intercept,sql/lock ./schema --template=./schema/templates/has_id.tmpl
