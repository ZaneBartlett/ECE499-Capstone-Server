package config

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"tech/app/logger"

	"github.com/mattn/go-sqlite3"
)

// DB - Imported database type from sql
type DB = sql.DB

// NewCfgService - will open/create an SQLite database object
func NewCfgService(dbPath string) (*CfgService, error) {

	cfg := CfgService{callbacks: make(map[string]*cbFuncs)}

	sql.Register("sqlite3_with_hooks",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				conn.RegisterUpdateHook(func(op int, db string, table string, rowid int64) {
					switch op {
					case sqlite3.SQLITE_INSERT:

						logger.LogDebug("Notified of insert: table '%s', configID '%d'\n", table, rowid)

					case sqlite3.SQLITE_UPDATE:

						logger.LogDebug("Notified of update on table '%s', configID '%d'\n", table, rowid)

					}
				})
				return nil
			},
		})

	database, err := sql.Open("sqlite3_with_hooks", dbPath)
	cfg.database = database

	return &cfg, err
}

// Set - Update data in the SQL database in table 'target', return a Get of the updated data
func (cfg *CfgService) Set(target string, data []byte) ([]byte, error) {

	unmarData, err := JsonToMap(data)
	if err != nil {
		logger.Log("Failed to unmarshall data on '%s'", target)
		return nil, err
	}
	response, err := set(cfg.database, target, unmarData)
	if err != nil {
		return nil, err
	}
	marResponse, err := MapToJson(response)
	if err != nil {
		logger.Log("Failed to marshall data on '%s'", target)
		return nil, err
	}

	return marResponse, err
}

// Get - Retrieves data from SQL database from table 'target' and returns a JSON object
func (cfg *CfgService) Get(target string) ([]byte, error) {

	response, err := get(cfg.database, target)
	if err != nil {
		return nil, err
	}
	marResponse, err := MapToJson(response)
	if err != nil {
		logger.Log("Failed to marshall data on '%s'", target)
		return nil, err
	}

	return marResponse, err
}

// Initialize -
func (cfg *CfgService) Initialize() {
	for _, call := range cfg.callbacks {
		call.onCreate(cfg)
	}
}

// CreateTable - Generate the 'target' table with 'schema' columns
func (cfg *CfgService) CreateTable(tableName string, schema []string) error {

	err := createTable(cfg.database, tableName, schema)

	return err
}

// InitTable - Create row 0 (if it does not exist) with default values in the Sql database table "target"
func (cfg *CfgService) InitTable(tableName string, data map[string]string) error {

	err := initTable(cfg.database, tableName, data)

	return err
}

// InitUser - Creates a new user row with rowID
func (cfg *CfgService) InitUser(tableName string, data map[string]string, rowId int) error {

	err := initUser(cfg.database, tableName, data, rowId)

	return err
}

// Register -
func (cfg *CfgService) Register(tableName string, onCreate func(cfg *CfgService) (err error)) {

	cfg.callbacks[tableName] = new(cbFuncs)
	cfg.callbacks[tableName].onCreate = onCreate

	return
}

// SetValue - Used to change on value in a table
func (cfg *CfgService) SetValue(tableName string, column string, value interface{}) error {

	data := make(map[string]interface{})
	data[column] = value
	_, err := set(cfg.database, tableName, data)

	return err
}

// SetUserValue - Change a value in the userAuth table
func (cfg *CfgService) SetUserValue(tableName string, username string, column string, value interface{}) error {

	data := make(map[string]interface{})
	data[column] = value
	_, err := setUser(cfg.database, tableName, username, data)

	return err
}

// GetValue - Used to retrieve one value from a table
func (cfg *CfgService) GetValue(tableName string, column string) (interface{}, error) {

	data, err := get(cfg.database, tableName)

	return data[column], err

}

// GetUser - Used to retrieve the row information of 'username'
func (cfg *CfgService) GetUser(tableName string, username string) (map[string]interface{}, error) {

	data, err := getUser(cfg.database, tableName, username)

	return data, err
}

func set(database *DB, target string, data map[string]interface{}) (map[string]interface{}, error) {

	// Generate query to initialize Table 'target'
	var colValString []string
	for col, value := range data {
		switch value.(type) {
		case (string):
			colValString = append(colValString, col+"='"+value.(string)+"'")

		case (bool):
			if value.(bool) {
				colValString = append(colValString, col+"=1")
			} else {
				colValString = append(colValString, col+"=0")
			}

		case (int), (uint), (int8), (uint8), (int16), (uint16), (int64), (uint64):
			colValString = append(colValString, col+"="+strconv.FormatInt(value.(int64), 10))

		case (float32), (float64):
			colValString = append(colValString, col+"="+strconv.FormatFloat(value.(float64), 'f', -1, 64))

		default:
			err := fmt.Errorf("Failure to update column:%s with value:%v", col, value)
			return nil, err
		}
	}

	query := "UPDATE " + target + " SET " + strings.Join(colValString, ",") + " WHERE configID=0"
	statement, err := database.Prepare(query)
	if err != nil {
		return nil, err
	}

	_, err = statement.Exec()
	if err != nil {
		return nil, err
	}

	err = statement.Close()
	if err != nil {
		return nil, err
	}

	return get(database, target)
}

func setUser(database *DB, target string, username string, data map[string]interface{}) (map[string]interface{}, error) {

	// Generate query to initialize Table 'target'
	var colValString []string
	for col, value := range data {
		switch value.(type) {
		case (string):
			colValString = append(colValString, col+"='"+value.(string)+"'")

		case (bool):
			if value.(bool) {
				colValString = append(colValString, col+"=1")
			} else {
				colValString = append(colValString, col+"=0")
			}

		case (int), (uint), (int8), (uint8), (int16), (uint16), (int64), (uint64):
			colValString = append(colValString, col+"="+strconv.FormatInt(value.(int64), 10))

		case (float32), (float64):
			colValString = append(colValString, col+"="+strconv.FormatFloat(value.(float64), 'f', -1, 64))

		default:
			err := fmt.Errorf("Failure to update column:%s with value:%v", col, value)
			return nil, err
		}
	}

	query := "UPDATE " + target + " SET " + strings.Join(colValString, ",") + " WHERE username='" + username + "'"
	statement, err := database.Prepare(query)
	if err != nil {
		return nil, err
	}

	_, err = statement.Exec()
	if err != nil {
		return nil, err
	}

	err = statement.Close()
	if err != nil {
		return nil, err
	}

	return get(database, target)
}

func get(database *DB, target string) (map[string]interface{}, error) {

	// Query database, store in 'row'
	query := "SELECT * FROM " + target + " WHERE configID=0"

	row, err := database.Query(query)
	defer row.Close()
	if err != nil {
		return nil, err
	}

	// Store row information into a map
	dataColumns, err := row.Columns()
	if err != nil {
		return nil, err
	}

	dataValues := make([]interface{}, len(dataColumns))
	for i := range dataValues {
		dataValues[i] = new(interface{})
	}

	for row.Next() {
		err = row.Scan(dataValues...)
		if err != nil {
			return nil, err
		}
		break
	}

	data := make(map[string]interface{}, len(dataColumns))
	for i, column := range dataColumns {
		data[column] = *(dataValues[i].(*interface{}))
	}

	return data, err
}

func getUser(database *DB, target string, username string) (map[string]interface{}, error) {

	// Query database, store in 'row'
	query := "SELECT * FROM " + target + " WHERE username='" + username + "'"

	row, err := database.Query(query)
	defer row.Close()
	if err != nil {
		return nil, err
	}

	// Store row information into a map
	dataColumns, err := row.Columns()
	if err != nil {
		return nil, err
	}

	dataValues := make([]interface{}, len(dataColumns))
	for i := range dataValues {
		dataValues[i] = new(interface{})
	}

	for row.Next() {
		err = row.Scan(dataValues...)
		if err != nil {
			return nil, err
		}
		break
	}

	data := make(map[string]interface{}, len(dataColumns))
	for i, column := range dataColumns {
		data[column] = *(dataValues[i].(*interface{}))
	}

	return data, err
}

func createTable(database *DB, tableName string, schema []string) error {

	// Assembles a query string to create a table 'target' with columns 'schema'
	query := "CREATE TABLE IF NOT EXISTS " + tableName +
		" (configID INTEGER PRIMARY KEY," + strings.Join(schema, ",") + ")"

	statement, err := database.Prepare(query)
	if err != nil {
		return err
	}

	_, err = statement.Exec()
	if err != nil {
		return err
	}

	err = statement.Close()
	if err != nil {
		return err
	}

	return err
}

// TODO - Investiagte using ON CONFLICT/OR in query string
func initTable(database *DB, tableName string, data map[string]string) error {

	var fieldString []string
	var valueString []string

	for field, value := range data {
		fieldString = append(fieldString, field)
		valueString = append(valueString, value)
	}

	query := "INSERT INTO " + tableName + " (configID," + strings.Join(fieldString, ",") +
		") VALUES (0," + strings.Join(valueString, ",") + ")"

	statement, err := database.Prepare(query)
	if err != nil {
		return err
	}

	_, err = statement.Exec()
	if err != nil {
		return err
	}

	err = statement.Close()
	if err != nil {
		return err
	}

	return err
}

func initUser(database *DB, tableName string, data map[string]string, rowId int) error {

	var fieldString []string
	var valueString []string

	for field, value := range data {
		fieldString = append(fieldString, field)
		valueString = append(valueString, value)
	}

	query := "INSERT INTO " + tableName + " (configID," + strings.Join(fieldString, ",") +
		") VALUES (?," + strings.Join(valueString, ",") + ")"

	statement, err := database.Prepare(query)
	if err != nil {
		return err
	}

	_, err = statement.Exec(rowId)
	if err != nil {
		return err
	}

	err = statement.Close()
	if err != nil {
		return err
	}

	return err
}

// JsonToMap - Converts a json blob into map format to allow for easier read/write
func JsonToMap(jsonData []byte) (map[string]interface{}, error) {

	mapData := make(map[string]interface{})

	err := json.Unmarshal(jsonData, &mapData)
	if err != nil {
		return nil, err
	}

	return mapData, err
}

// MapToJson - Converts a map into a json blob to allow for packet transfer
func MapToJson(mapData map[string]interface{}) ([]byte, error) {

	jsonData, err := json.MarshalIndent(mapData, "", "\t")
	if err != nil {
		return nil, err
	}

	return jsonData, err
}

// JSONbool - Safely convert JSON value to bool
func JSONbool(value interface{}) (bool, error) {
	var b bool
	var err error

	switch value.(type) {

	case (bool):
		b = value.(bool)
		err = nil

	case (int64):
		b = !(value.(int64) == 0)
		err = nil

	case (float64):
		b = !(value.(float64) == 0.0)
		err = nil

	case (float32):
		b = !(value.(float32) == 0.0)
		err = nil

	case (int32):
		b = !(value.(int32) == 0)
		err = nil

	case (int16):
		b = !(value.(int16) == 0)
		err = nil

	case (int8):
		b = !(value.(int8) == 0)
		err = nil

	case (int):
		b = !(value.(int) == 0)
		err = nil

	case (string):
		if (value.(string) == "") || (value.(string) == "false") || (value.(string) == "FALSE") {
			b = false
		} else {
			b = true
		}

		err = nil

	default:
		b = false
		err = fmt.Errorf("Failure to cast '%v' as int, type is %T", value, value)
	}
	return b, err
}

// JSONint64 - Safely convert JSON value to int64
func JSONint64(value interface{}) (int64, error) {
	var temp int
	var i int64
	var err error

	switch value.(type) {

	case (bool):
		if value.(bool) == true {
			i = 1
		} else {
			i = 0
		}
		err = nil

	case (int64):
		i = value.(int64)
		err = nil

	case (int32):
		i = int64(value.(int32))
		err = nil

	case (int16):
		i = int64(value.(int16))
		err = nil

	case (int8):
		i = int64(value.(int8))
		err = nil

	case (int):
		i = int64(value.(int))
		err = nil

	case (float64):
		i = int64(value.(float64))
		err = nil

	case (float32):
		i = int64(value.(float32))
		err = nil

	case (string):
		temp, err = strconv.Atoi(value.(string))
		i = int64(temp)

	default:
		i = 0
		err = fmt.Errorf("Failure to cast '%v' as int, type is %T", value, value)
	}

	return i, err
}

// JSONint32 - Safely convert JSON value to int32
func JSONint32(value interface{}) (int32, error) {
	var temp int
	var i int32
	var err error

	switch value.(type) {

	case (bool):
		if value.(bool) == true {
			i = 1
		} else {
			i = 0
		}
		err = nil

	case (int64):
		i = int32(value.(int64))
		err = nil

	case (int32):
		i = value.(int32)
		err = nil

	case (int16):
		i = int32(value.(int16))
		err = nil

	case (int8):
		i = int32(value.(int8))
		err = nil

	case (int):
		i = int32(value.(int))
		err = nil

	case (float64):
		i = int32(value.(float64))
		err = nil

	case (float32):
		i = int32(value.(float32))
		err = nil

	case (string):
		temp, err = strconv.Atoi(value.(string))
		i = int32(temp)

	default:
		i = 0
		err = fmt.Errorf("Failure to cast '%v' as int, type is %T", value, value)
	}

	return i, err
}

// JSONint16 - Safely convert JSON value to int16
func JSONint16(value interface{}) (int16, error) {
	var temp int
	var i int16
	var err error

	switch value.(type) {

	case (bool):
		if value.(bool) == true {
			i = 1
		} else {
			i = 0
		}
		err = nil

	case (int64):
		i = int16(value.(int64))
		err = nil

	case (int32):
		i = int16(value.(int32))
		err = nil

	case (int16):
		i = value.(int16)
		err = nil

	case (int8):
		i = int16(value.(int8))
		err = nil

	case (int):
		i = int16(value.(int))
		err = nil

	case (float64):
		i = int16(value.(float64))
		err = nil

	case (float32):
		i = int16(value.(float32))
		err = nil

	case (string):
		temp, err = strconv.Atoi(value.(string))
		i = int16(temp)

	default:
		i = 0
		err = fmt.Errorf("Failure to cast '%v' as int, type is %T", value, value)
	}

	return i, err
}

// JSONint8 - Safely convert JSON value to int8
func JSONint8(value interface{}) (int8, error) {
	var temp int
	var i int8
	var err error

	switch value.(type) {

	case (bool):
		if value.(bool) == true {
			i = 1
		} else {
			i = 0
		}
		err = nil

	case (int64):
		i = int8(value.(int64))
		err = nil

	case (int32):
		i = int8(value.(int32))
		err = nil

	case (int16):
		i = int8(value.(int16))
		err = nil

	case (int8):
		i = value.(int8)
		err = nil

	case (int):
		i = int8(value.(int))
		err = nil

	case (float64):
		i = int8(value.(float64))
		err = nil

	case (float32):
		i = int8(value.(float32))
		err = nil

	case (string):
		temp, err = strconv.Atoi(value.(string))
		i = int8(temp)

	default:
		i = 0
		err = fmt.Errorf("Failure to cast '%v' as int, type is %T", value, value)
	}

	return i, err
}

// JSONint - Safely convert JSON value to int
func JSONint(value interface{}) (int, error) {
	var i int
	var err error

	switch value.(type) {

	case (bool):
		if value.(bool) == true {
			i = 1
		} else {
			i = 0
		}
		err = nil

	case (int64):
		i = int(value.(int64))
		err = nil

	case (int32):
		i = int(value.(int32))
		err = nil

	case (int16):
		i = int(value.(int16))
		err = nil

	case (int8):
		i = int(value.(int8))
		err = nil

	case (int):
		i = value.(int)
		err = nil

	case (float64):
		i = int(value.(float64))
		err = nil

	case (float32):
		i = int(value.(float32))
		err = nil

	case (string):
		i, err = strconv.Atoi(value.(string))

	default:
		i = 0
		err = fmt.Errorf("Failure to cast '%v' as int, type is %T", value, value)
	}

	return i, err
}

// JSONuint32 - Safely convert JSON value to uint32
func JSONuint32(value interface{}) (uint32, error) {
	var temp int
	var retVal uint32
	var err error

	switch value.(type) {

	case (bool):
		if value.(bool) == true {
			retVal = 1
		} else {
			retVal = 0
		}
		err = nil

	case (int64):
		retVal = uint32(value.(int64))
		err = nil

	case (int32):
		retVal = uint32(value.(int32))
		err = nil

	case (int16):
		retVal = uint32(value.(uint16))
		err = nil

	case (int8):
		retVal = uint32(value.(int8))
		err = nil

	case (int):
		retVal = uint32(value.(int))
		err = nil

	case (float64):
		retVal = uint32(value.(float64))
		err = nil

	case (float32):
		retVal = uint32(value.(float32))
		err = nil

	case (string):
		temp, err = strconv.Atoi(value.(string))
		retVal = uint32(temp)

	default:
		retVal = 0
		err = fmt.Errorf("Failure to cast '%v' as int, type is %T", value, value)
	}

	return retVal, err
}

// JSONfloat64 - Safely convert JSON value to float64
func JSONfloat64(value interface{}) (float64, error) {
	var retVal float64
	var err error

	switch value.(type) {

	case (bool):
		if value.(bool) == true {
			retVal = 1.0
		} else {
			retVal = 0.0
		}
		err = nil

	case (int64):
		retVal = float64(value.(int64))
		err = nil

	case (int32):
		retVal = float64(value.(int32))
		err = nil

	case (int16):
		retVal = float64(value.(int16))
		err = nil

	case (int8):
		retVal = float64(value.(int8))
		err = nil

	case (int):
		retVal = float64(value.(int))
		err = nil

	case (float64):
		retVal = value.(float64)
		err = nil

	case (float32):
		retVal = float64(value.(float32))
		err = nil

	case (string):
		retVal, err = strconv.ParseFloat(value.(string), 64)

	default:
		retVal = 0.0
		err = fmt.Errorf("Failure to cast '%v' as int, type is %T", value, value)
	}

	return retVal, err
}

// JSONstring - Safely convert JSON value to string
func JSONstring(value interface{}) (string, error) {
	var s string
	var err error

	switch value.(type) {

	case (bool):
		if value.(bool) == true {
			s = "true"
		} else {
			s = "false"
		}
		err = nil

	case (int64):
		s = fmt.Sprintf("%d", value.(int64))
		err = nil

	case (int32):
		s = fmt.Sprintf("%d", value.(int32))
		err = nil

	case (int16):
		s = fmt.Sprintf("%d", value.(int16))
		err = nil

	case (int8):
		s = fmt.Sprintf("%d", value.(int8))
		err = nil

	case (int):
		s = fmt.Sprintf("%d", value.(int))
		err = nil

	case (float64):
		s = fmt.Sprintf("%f", value.(float64))
		err = nil

	case (string):
		s = value.(string)

	default:
		s = ""
		err = fmt.Errorf("Failure to cast '%v' as int, type is %T", value, value)
	}

	return s, err
}
