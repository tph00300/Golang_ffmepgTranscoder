package database

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/new_transcoder/transcoder/utils/exceptions"
)

type DBMS struct {
	dbms   string
	user   string
	passwd string
	host   string
	port   string
	dbname string
}

func NewDBMS(dbname string) *DBMS {
	d := DBMS{}
	d.dbms = "mysql"
	d.user = "atsdevops"
	d.passwd = "Qwpo1209"
	d.host = "180.66.171.100"
	d.port = "3306"
	d.dbname = dbname
	return &d
}
func NewDBMSWithHost(id string, password string, host string, dbname string) *DBMS {
	d := DBMS{}
	d.dbms = "mysql"
	d.user = id
	d.passwd = password
	d.host = host
	d.port = "3306"
	d.dbname = dbname
	return &d
}

//MYSQL VERSION
func (d *DBMS) MySQLMultirowQuery(q string) (*[]map[string]interface{}, error) {
	//init DBMS
	db, err := sql.Open(d.dbms, d.user+":"+d.passwd+"@tcp("+d.host+")/"+d.dbname)
	if err != nil {
		return nil, exceptions.New("Can not open mysql : " + err.Error())
	}
	defer db.Close()
	//query
	rows, err := db.Query(q)
	if err != nil {
		return nil, exceptions.New("query fail [rows] : " + err.Error())
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, exceptions.New("query fail [cols] : " + err.Error())
	}
	defer rows.Close()
	//output
	var tmp []map[string]interface{}
	count := 0
	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}
		// Scan the result into the column pointers...
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, exceptions.New("scan fail : " + err.Error())
		}
		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			if columns[i] != nil {
				val := string(columns[i].([]byte))
				m[colName] = val
			}
		}
		tmp = append(tmp, m)
		count++
	}
	result := make([]map[string]interface{}, count)
	copy(result, tmp)
	return &result, nil
}
func (d *DBMS) MySQLExec(q string) error {
	//init DBMS
	db, err := sql.Open(d.dbms, d.user+":"+d.passwd+"@tcp("+d.host+")/"+d.dbname)
	if err != nil {
		return exceptions.New("Can not open mysql : " + err.Error())
	}
	defer db.Close()
	// INSERT 문 실행
	result, err := db.Exec(q)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}
