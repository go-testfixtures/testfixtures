package testfixtures

var (
	skipDatabaseNameCheck bool
)

// SkipDatabaseNameCheck If true, loading fixtures will not check if the database
// name constaint "test". Use with caution!
func SkipDatabaseNameCheck(value bool) {
	skipDatabaseNameCheck = value
}
