package config

type cbFuncs struct {
	onCreate func(cfg *CfgService) (err error)
}

// CfgService - Contains a pointer to the SQLite database
// TODO - callbacks aren't technically being used, so this should change, however they work for now as is
type CfgService struct {
	database  *DB
	callbacks map[string]*cbFuncs
}
