//go:build unit || integration || acceptance

package helpers

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/internal/cmd/provider"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// DB returns the database connection as specified by the configuration
// file. Note that the migrations are handled by the provider.
func DB() *sqlx.DB {
	return provideDB().DB.GetSqlxDB()
}

func provideDB() *provider.Database {
	return provider.ProvideDB("chorus", provider.WithClient("acceptance-tests"), provider.WithMigrations(migration.GetMigration))
}

func Populate(query string, args ...interface{}) {

	db := provideDB().DB

	if _, err := db.Exec(prepareForInsertion(query), args...); err != nil {
		panic(err)
	}
}

func Dump(query string, args ...interface{}) {
	db := provideDB().DB

	rows, err := db.QueryContext(context.Background(), prepareForInsertion(query), args...)
	if err != nil {
		fmt.Println("Error executing query:", err)
	}

	defer rows.Close()
	if !rows.Next() {
		fmt.Println("No rows returned")
		return
	}
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println("Error getting columns:", err)
		return
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	if err := rows.Scan(valuePtrs...); err != nil {
		fmt.Println("Error scanning row:", err)
		return
	}
	fmt.Println("----------- Dumped query result: ------")
	for _, col := range columns {
		fmt.Printf("%s, ", col)
	}
	fmt.Println()
	for i := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			val = string(b)
		}
		fmt.Printf("%v, ", val)
	}
	fmt.Println("\n-------------------------")
	// fmt.Println("Dumped query result:")
	// spew.Dump(values)

}

var (
	regexStatement   = regexp.MustCompile(`\)[\s\n\t]*;`)
	regexInsertTable = regexp.MustCompile("INSERT INTO (\"?[a-z_]+\"?)")
	regexNow         = regexp.MustCompile("(\\W)(NOW|now)\\(")
	regexFalse       = regexp.MustCompile(`([(, ])(FALSE|false)([),])`)
	regexTrue        = regexp.MustCompile(`([(, ])(TRUE|true)([),])`)
)

func prepareForInsertion(query string) string {
	return sqlx.Rebind(sqlx.DOLLAR, query)
}

func CountRow(t *testing.T, db *sqlx.DB, table, filterName string, filterValue uint64) int {
	bindSymbol := "$"
	var count int
	var query = fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s=%s1`, table, filterName, bindSymbol)
	err := db.Get(&count, query, filterValue)
	require.Nil(t, err)
	return count
}
