package services

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/sysadminsmedia/homebox/backend/internal/sys/config"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sysadminsmedia/homebox/backend/internal/core/currencies"
	"github.com/sysadminsmedia/homebox/backend/internal/core/services/reporting/eventbus"
	"github.com/sysadminsmedia/homebox/backend/internal/data/ent"
	"github.com/sysadminsmedia/homebox/backend/internal/data/repo"
	"github.com/sysadminsmedia/homebox/backend/pkgs/faker"
)

var (
	fk   = faker.NewFaker()
	tbus = eventbus.New()

	tCtx    = Context{}
	tClient *ent.Client
	tRepos  *repo.AllRepos
	tUser   repo.UserOut
	tGroup  repo.Group
	tSvc    *AllServices
)

func bootstrap() {
	var (
		err error
		ctx = context.Background()
	)

	tGroup, err = tRepos.Groups.GroupCreate(ctx, "test-group")
	if err != nil {
		log.Fatal(err)
	}

	password := fk.Str(10)
	tUser, err = tRepos.Users.Create(ctx, repo.UserCreate{
		Name:           fk.Str(10),
		Email:          fk.Email(),
		Password:       &password,
		IsSuperuser:    fk.Bool(),
		DefaultGroupID: tGroup.ID,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func MainNoExit(m *testing.M) int {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1&_time_format=sqlite")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}

	err = client.Schema.Create(context.Background())
	if err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	tClient = client
	tRepos = repo.New(tClient, tbus, config.Storage{
		PrefixPath: "/",
		ConnString: "file://" + os.TempDir(),
	}, "mem://{{ .Topic }}", config.Thumbnail{
		Enabled: false,
		Width:   0,
		Height:  0,
	})

	err = os.MkdirAll(os.TempDir()+"/homebox", 0o755)
	if err != nil {
		return 0
	}

	defaults, _ := currencies.CollectionCurrencies(
		currencies.CollectDefaults(),
	)

	tSvc = New(tRepos, WithCurrencies(defaults))
	defer func() { _ = client.Close() }()

	bootstrap()
	tCtx = Context{
		Context: context.Background(),
		GID:     tGroup.ID,
		UID:     tUser.ID,
	}

	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(MainNoExit(m))
}
