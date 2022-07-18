package dbstore

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

const (
	driverName = "postgres"
	dsn        = "host=localhost port=5432 user=postgres password=123456 dbname=postgres sslmode=disable"
)

func TestTokenStore(t *testing.T) {
	Convey("Test mysql token store", t, func() {
		store := NewDefaultStore(NewConfig(driverName, dsn))
		defer store.clean()

		ctx := context.Background()

		Convey("Test authorization code store", func() {
			info := &models.Token{
				ClientID:      "1",
				UserID:        "1_1",
				RedirectURI:   "http://localhost/",
				Scope:         "all",
				Code:          "11_11_11",
				CodeCreateAt:  time.Now(),
				CodeExpiresIn: time.Second * 5,
			}
			err := store.Create(ctx, info)
			So(err, ShouldBeNil)

			cinfo, err := store.GetByCode(ctx, info.Code)
			So(err, ShouldBeNil)
			So(cinfo.GetUserID(), ShouldEqual, info.UserID)

			err = store.RemoveByCode(ctx, info.Code)
			So(err, ShouldBeNil)

			cinfo, err = store.GetByCode(ctx, info.Code)
			So(err, ShouldBeNil)
			So(cinfo, ShouldBeNil)
		})

		Convey("Test access token store", func() {
			info := &models.Token{
				ClientID:        "1",
				UserID:          "1_1",
				RedirectURI:     "http://localhost/",
				Scope:           "all",
				Access:          "1_1_1",
				AccessCreateAt:  time.Now(),
				AccessExpiresIn: time.Second * 5,
			}
			err := store.Create(ctx, info)
			So(err, ShouldBeNil)

			ainfo, err := store.GetByAccess(ctx, info.GetAccess())
			So(err, ShouldBeNil)
			So(ainfo.GetUserID(), ShouldEqual, info.GetUserID())

			err = store.RemoveByAccess(ctx, info.GetAccess())
			So(err, ShouldBeNil)

			ainfo, err = store.GetByAccess(ctx, info.GetAccess())
			So(err, ShouldBeNil)
			So(ainfo, ShouldBeNil)
		})

		Convey("Test refresh token store", func() {
			info := &models.Token{
				ClientID:         "1",
				UserID:           "1_2",
				RedirectURI:      "http://localhost/",
				Scope:            "all",
				Access:           "1_2_1",
				AccessCreateAt:   time.Now(),
				AccessExpiresIn:  time.Second * 5,
				Refresh:          "1_2_2",
				RefreshCreateAt:  time.Now(),
				RefreshExpiresIn: time.Second * 15,
			}
			err := store.Create(ctx, info)
			So(err, ShouldBeNil)

			ainfo, err := store.GetByAccess(ctx, info.GetAccess())
			So(err, ShouldBeNil)
			So(ainfo.GetUserID(), ShouldEqual, info.GetUserID())

			err = store.RemoveByAccess(ctx, info.GetAccess())
			So(err, ShouldBeNil)

			ainfo, err = store.GetByAccess(ctx, info.GetAccess())
			So(err, ShouldBeNil)
			So(ainfo, ShouldBeNil)

			rinfo, err := store.GetByRefresh(ctx, info.GetRefresh())
			So(err, ShouldBeNil)
			So(rinfo.GetUserID(), ShouldEqual, info.GetUserID())

			err = store.RemoveByRefresh(ctx, info.GetRefresh())
			So(err, ShouldBeNil)

			rinfo, err = store.GetByRefresh(ctx, info.GetRefresh())
			So(err, ShouldBeNil)
			So(rinfo, ShouldBeNil)
		})
	})
}

func TestNewStoreWithOpts_ShouldReturnStoreNotNil(t *testing.T) {
	// ARRANGE
	db, mockDB, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	tableName := "custom_table_name"

	// Mock sql exec create table
	mockDB.ExpectExec(regexp.QuoteMeta(`create table if not exists custom_table_name (
		id bigint not null primary key,
		expired_at bigint,
		code varchar(255),
		access varchar(255),
		refresh varchar(255),
		data varchar(2048)
	)`)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Mock query:
	mockDB.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM custom_table_name WHERE expired_at<=? OR (code='' AND access='' AND refresh='')")).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0))

	// ACTION
	store := NewStoreWithOpts(
		sqlxDB,
		WithTableName(tableName),
		WithGCTimeInterval(1000),
	)

	defer store.clean()

	// ASSERT
	assert.NotNil(t, store)
	assert.NotNil(t, store.ticker)
	assert.Equal(t, store.tableName, tableName)
}
