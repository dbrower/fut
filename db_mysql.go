package fut

import (
	"database/sql"
	"log"

	"github.com/BurntSushi/migration"
	_ "github.com/go-sql-driver/mysql"
)

// store pointer to sql database
type MysqlDB struct {
	db *sql.DB
}

var migrations = []migration.Migrator{
	migration1,
}

func migration1(tx migration.LimitedTx) error {
	var s = []string{
		`CREATE TABLE IF NOT EXISTS triples (
		id int PRIMARY KEY AUTO_INCREMENT,
		Subject varchar(64),
		Predicate varchar(255),
		Object text,
		Sequence int)`,
	}
	return execlist(tx, s)
}

// execlist exec's each item in the list, return if there is an error.
// Used to work around mysql driver not handling compound exec statements.
func execlist(tx migration.LimitedTx, stms []string) error {
	var err error
	for _, s := range stms {
		_, err = tx.Exec(s)
		if err != nil {
			break
		}
	}
	return err
}

type dbVersion struct {
	// SQL to get the version of this db, returns one row and one column
	GetSQL string
	// SQL to insert a new version of this db. takes one parameter, the new
	// version
	SetSQL string
	// the SQL to create the version table for this db
	CreateSQL string
}

func (d dbVersion) Get(tx migration.LimitedTx) (int, error) {
	v, err := d.get(tx)
	if err != nil {
		// we assume error means there is no migration table
		log.Println(err.Error())
		log.Println("Assuming this is because there is no migration table, yet")
		return 0, nil
	}
	return v, nil
}

func (d dbVersion) Set(tx migration.LimitedTx, version int) error {
	if err := d.set(tx, version); err != nil {
		if err := d.createTable(tx); err != nil {
			return err
		}
		return d.set(tx, version)
	}
	return nil
}

func (d dbVersion) get(tx migration.LimitedTx) (int, error) {
	var version int
	r := tx.QueryRow(d.GetSQL)
	if err := r.Scan(&version); err != nil {
		return 0, err
	}
	return version, nil
}

func (d dbVersion) set(tx migration.LimitedTx, version int) error {
	_, err := tx.Exec(d.SetSQL, version)
	return err
}

func (d dbVersion) createTable(tx migration.LimitedTx) error {
	_, err := tx.Exec(d.CreateSQL)
	if err == nil {
		err = d.set(tx, 0)
	}
	return err
}

var mysqlVersioning = dbVersion{
	GetSQL:    `SELECT max(version) FROM migration_version`,
	SetSQL:    `INSERT INTO migration_version (version, applied) VALUES (?, now())`,
	CreateSQL: `CREATE TABLE migration_version (version INTEGER, applied datetime)`,
}

// NewMySQL returns a Repository backed by a MySQL database, as determined
// by the connection string. An error is returned if any problems are run into.
func NewMySQL(conn string) (*MysqlDB, error) {
	conn += "?parseTime=true"
	db, err := migration.OpenWith(
		"mysql",
		conn,
		migrations,
		mysqlVersioning.Get,
		mysqlVersioning.Set,
	)
	if err != nil {
		return nil, err
	}
	return &MysqlDB{db: db}, nil
}

// AllPurls returns a list of every purl in the database.
func (sq *MysqlDB) IndexItem(item CurateItem) error {
	tx, err := sq.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // does nothing if tx has been committed

	_, err = tx.Exec(
		`DELETE FROM triples WHERE subject = ?`,
		item.PID,
	)
	if err != nil {
		return err
	}

	for seq, t := range item.Properties {
		_, err = tx.Exec(
			`INSERT INTO triples (subject, predicate, object, sequence)
			VALUES (?, ?, ?, ?)`,
			item.PID,
			t.Predicate,
			t.Object,
			seq,
		)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

// FindItem returns a single CurateItem record identified by PID.
func (sq *MysqlDB) FindItem(pid string) (CurateItem, error) {
	result := CurateItem{PID: pid}
	rows, err := sq.db.Query(`
		SELECT predicate, object
		FROM triples
		WHERE subject = ?
		ORDER BY sequence`,
		pid)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var pair Pair
		err2 := rows.Scan(&pair.Predicate, &pair.Object)
		if err2 != nil {
			// propogate error out so it is returned
			// but try to read as many properties as possible
			err = err2
		} else {
			result.Properties = append(result.Properties, pair)
		}
	}
	return result, err
}

func (sq *MysqlDB) FindItemAndFiles(pid string) ([]CurateItem, error) {
	var result []CurateItem
	rows, err := sq.db.Query(`
		SELECT subject, predicate, object
		FROM triples
		WHERE subject = ? OR subject IN (
			SELECT subject
			FROM triples
			WHERE predicate = "isPartOf" and object = ?)
		ORDER BY subject, sequence`,
		pid,
		pid,
	)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	current := &CurateItem{}
	for rows.Next() {
		var subject string
		var pair Pair
		err2 := rows.Scan(&subject, &pair.Predicate, &pair.Object)
		if err2 != nil {
			err = err2
			continue
		}
		if current.PID == "" {
			current.PID = subject
		} else if current.PID != subject {
			result = append(result, *current)
			current = &CurateItem{PID: subject}
		}
		current.Properties = append(current.Properties, pair)
	}
	if current.PID != "" {
		result = append(result, *current)
	}
	return result, nil
}

func (sq *MysqlDB) FindAllRange(offset, count int) ([]CurateItem, error) {
	log.Println("findallrange", offset, count)
	var result []CurateItem
	rows, err := sq.db.Query(`
		SELECT subject, predicate, object
		FROM triples
		WHERE subject IN (
			SELECT subject
			FROM triples
			WHERE predicate = "af-model"
			LIMIT ? OFFSET ?)
		ORDER BY subject, sequence`,
		count,
		offset,
	)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	current := &CurateItem{}
	for rows.Next() {
		var subject string
		var pair Pair
		err2 := rows.Scan(&subject, &pair.Predicate, &pair.Object)
		if err2 != nil {
			err = err2
			continue
		}
		if current.PID == "" {
			current.PID = subject
		} else if current.PID != subject {
			result = append(result, *current)
			current = &CurateItem{PID: subject}
		}
		current.Properties = append(current.Properties, pair)
	}
	if current.PID != "" {
		result = append(result, *current)
	}
	return result, nil
}
