package testfixtures

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type fixtureFile struct {
	path     string
	fileName string
	content  []byte
}

func (f *fixtureFile) fileNameWithoutExtension() string {
	return strings.Replace(f.fileName, filepath.Ext(f.fileName), "", 1)
}

func (f *fixtureFile) delete(tx *sql.Tx, h DataBaseHelper) error {
	_, err := tx.Exec(fmt.Sprintf("DELETE FROM %s", h.quoteKeyword(f.fileNameWithoutExtension())))
	return err
}

func (f *fixtureFile) insert(tx *sql.Tx, h DataBaseHelper) error {
	var rows []interface{}
	err := yaml.Unmarshal(f.content, &rows)
	if err != nil {
		return err
	}

	for _, row := range rows {
		record := row.(map[interface{}]interface{})
		var values []interface{}

		sqlColumns := ""
		sqlValues := ""
		i := 1
		for key, value := range record {
			if len(sqlColumns) > 0 {
				sqlColumns += ", "
				sqlValues += ", "
			}
			sqlColumns += h.quoteKeyword(key.(string))
			switch h.paramType() {
			case paramTypeDollar:
				sqlValues += fmt.Sprintf("$%d", i)
			case paramTypeQuestion:
				sqlValues += "?"
			case paramTypeColon:
				if isDateTime(value) {
					sqlValues += fmt.Sprintf("to_date(:%d, 'YYYY-MM-DD HH24:MI:SS')", i)
				} else if isDate(value) {
					sqlValues += fmt.Sprintf("to_date(:%d, 'YYYY-MM-DD')", i)
				} else if isTime(value) {
					sqlValues += fmt.Sprintf("to_date(:%d, 'HH24:MI:SS')", i)
				} else {
					sqlValues += fmt.Sprintf(":%d", i)
				}
			}
			i++
			values = append(values, value)
		}

		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", h.quoteKeyword(f.fileNameWithoutExtension()), sqlColumns, sqlValues)
		_, err = tx.Exec(sql, values...)
		if err != nil {
			return err
		}
	}
	return nil
}

func getYmlFiles(foldername string) ([]*fixtureFile, error) {
	var files []*fixtureFile
	fileinfos, err := ioutil.ReadDir(foldername)
	if err != nil {
		return nil, err
	}

	for _, fileinfo := range fileinfos {
		if !fileinfo.IsDir() && filepath.Ext(fileinfo.Name()) == ".yml" {
			fixture := &fixtureFile{
				path:     path.Join(foldername, fileinfo.Name()),
				fileName: fileinfo.Name(),
			}
			fixture.content, err = ioutil.ReadFile(fixture.path)
			if err != nil {
				return nil, err
			}
			files = append(files, fixture)
		}
	}
	return files, nil
}

// LoadFixtureFiles load all specified fixtures files to database
func LoadFixtureFiles(db *sql.DB, h DataBaseHelper, files ...string) error {
	var fixtureFiles []*fixtureFile
	var err error
	for _, f := range files {
		fixture := &fixtureFile{
			path:     f,
			fileName: filepath.Base(f),
		}
		fixture.content, err = ioutil.ReadFile(fixture.path)
		if err != nil {
			return err
		}
		fixtureFiles = append(fixtureFiles, fixture)
	}

	return loadFixtures(db, h, fixtureFiles...)
}

// LoadFixtures loads all fixtures in a given folder in the database
func LoadFixtures(foldername string, db *sql.DB, h DataBaseHelper) error {
	fixturesFiles, err := getYmlFiles(foldername)
	if err != nil {
		return err
	}

	return loadFixtures(db, h, fixturesFiles...)
}

func loadFixtures(db *sql.DB, h DataBaseHelper, fixturesFiles ...*fixtureFile) error {
	if !skipDatabaseNameCheck {
		if !dbnameRegexp.MatchString(h.databaseName(db)) {
			return errNotTestDatabase
		}
	}

	err := h.disableReferentialIntegrity(db, func(tx *sql.Tx) error {
		for _, file := range fixturesFiles {
			err := file.delete(tx, h)
			if err != nil {
				return err
			}

			err = h.whileInsertOnTable(tx, file.fileNameWithoutExtension(), func() error {
				return file.insert(tx, h)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
