package repo

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dezswap/dezswap-api/indexer"
	indexer_db "github.com/dezswap/dezswap-api/pkg/db/indexer"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func newMockedDbRepo(t *testing.T, mapper dbMapper) (*dbRepoImpl, sqlmock.Sqlmock) {
	t.Helper()

	conn, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 conn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{NamingStrategy: schema.NamingStrategy{}})
	require.NoError(t, err)

	return &dbRepoImpl{mapper, gormDB, gormDB, "chainId"}, mock
}

func idRows(n int) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"id"})
	for i := 1; i <= n; i++ {
		rows.AddRow(i)
	}
	return rows
}

// The insert must carry no id column and must leave created_at out of the conflict
// update, so an existing row keeps the moment it was first seen.
func Test_SaveTokens_UpsertsWithoutIdAndKeepsCreatedAt(t *testing.T) {
	r, mock := newMockedDbRepo(t, &dbMapperImpl{})

	columns := regexp.QuoteMeta(`INSERT INTO "tokens" ("created_at","updated_at","deleted_at","chain_id","address","protocol","symbol","name","decimals","icon","verified") VALUES`)
	conflict := regexp.QuoteMeta(`ON CONFLICT ("chain_id","address") DO UPDATE SET "updated_at"="excluded"."updated_at","protocol"="excluded"."protocol","symbol"="excluded"."symbol","name"="excluded"."name","decimals"="excluded"."decimals","icon"="excluded"."icon","verified"="excluded"."verified"`)

	mock.ExpectBegin()
	mock.ExpectQuery(columns + `.*` + conflict).WillReturnRows(idRows(1))
	mock.ExpectCommit()

	// an id on the way in must not reach the statement, or tokens_pkey rather than the
	// declared arbiter would decide the conflict
	require.NoError(t, r.SaveTokens([]indexer.Token{
		{ID: 42, ChainId: "chainId", Address: "addr1", Symbol: "A", Decimals: 6},
	}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SaveTokens_SendsEveryRowInOneStatement(t *testing.T) {
	r, mock := newMockedDbRepo(t, &dbMapperImpl{})

	mock.ExpectBegin()
	// two value tuples in a single insert rather than a statement per row
	mock.ExpectQuery(`INSERT INTO "tokens".*VALUES \(\$1,.*\),\(\$12,`).WillReturnRows(idRows(2))
	mock.ExpectCommit()

	require.NoError(t, r.SaveTokens([]indexer.Token{
		{ChainId: "chainId", Address: "addr1", Symbol: "A", Decimals: 6},
		{ChainId: "chainId", Address: "addr2", Symbol: "B", Decimals: 6},
	}))
	require.NoError(t, mock.ExpectationsWereMet())
}

// nilModelMapper mimics a mapper that leaves the embedded gorm.Model unset, the way
// poolToPoolModel does.
type nilModelMapper struct{ dbMapper }

func (nilModelMapper) tokensToModels(tokens []indexer.Token) ([]indexer_db.Token, error) {
	models := make([]indexer_db.Token, 0, len(tokens))
	for _, token := range tokens {
		models = append(models, indexer_db.Token{
			ChainModel: indexer_db.ChainModel{ChainId: token.ChainId, Address: token.Address},
			Symbol:     token.Symbol,
		})
	}
	return models, nil
}

func Test_SaveTokens_ToleratesModelsWithoutAnEmbeddedModel(t *testing.T) {
	r, mock := newMockedDbRepo(t, nilModelMapper{})

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tokens"`).WillReturnRows(idRows(1))
	mock.ExpectCommit()

	// clearing the id must not dereference a nil gorm.Model
	require.NotPanics(t, func() {
		require.NoError(t, r.SaveTokens([]indexer.Token{{ChainId: "chainId", Address: "addr1"}}))
	})
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SaveTokens_NoRowsTouchesNothing(t *testing.T) {
	r, mock := newMockedDbRepo(t, &dbMapperImpl{})

	require.NoError(t, r.SaveTokens(nil))
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SaveLatestPools_SendsEveryRowInOneStatement(t *testing.T) {
	r, mock := newMockedDbRepo(t, &dbMapperImpl{})

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "latest_pools".*VALUES \(\$1,.*\),\(\$13,.*` + regexp.QuoteMeta(`ON CONFLICT ("address","chain_id") DO UPDATE`)).
		WillReturnRows(idRows(2))
	mock.ExpectCommit()

	require.NoError(t, r.SaveLatestPools([]indexer.PoolInfo{
		{ChainId: "chainId", Address: "pool1", Asset0: "a0", Asset0Amount: "1", Asset1: "a1", Asset1Amount: "2", LpAmount: "3"},
		{ChainId: "chainId", Address: "pool2", Asset0: "a0", Asset0Amount: "4", Asset1: "a1", Asset1Amount: "5", LpAmount: "6"},
	}, 100))
	require.NoError(t, mock.ExpectationsWereMet())
}

func Test_SaveLatestPools_NoRowsTouchesNothing(t *testing.T) {
	r, mock := newMockedDbRepo(t, &dbMapperImpl{})

	require.NoError(t, r.SaveLatestPools(nil, 100))
	require.NoError(t, mock.ExpectationsWereMet())
}
