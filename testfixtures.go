package testfixtures

import (
    "database/sql"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "strings"
    "gopkg.in/yaml.v2"
)

type FixtureFile struct {
    Path string
    FileName string
    Content []byte
}

func (f *FixtureFile) FileNameWithoutExtension() string {
    return strings.Replace(f.FileName, filepath.Ext(f.FileName), "", 1)
}

func (f *FixtureFile) Delete(tx *sql.Tx) error {
    _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s", f.FileNameWithoutExtension()))
    return err
}

func (f *FixtureFile) Insert(tx *sql.Tx) error {
    rows := make([]interface{}, 0)
    err := yaml.Unmarshal(f.Content, &rows)
    if err != nil {
        return err
    }

    for _, row := range rows {
        record := row.(map[interface{}]interface{})
        values := make([]interface{}, 0)

        sqlColumns := ""
        sqlValues := ""
        i := 1
        for key, value := range record {
            if sqlColumns != "" {
                sqlColumns = sqlColumns + ","
                sqlValues = sqlValues + ","
            }
            sqlColumns = sqlColumns + key.(string)
            sqlValues = fmt.Sprintf("%s$%d", sqlValues, i)
            i++
            values = append(values, value)
        }

        sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", f.FileNameWithoutExtension(), sqlColumns, sqlValues)
        _, err = tx.Exec(sql, values...)
        if err != nil {
            return err
        }
    }
    return nil
}

func getYmlFiles(foldername string) ([]*FixtureFile, error) {
    files := make([]*FixtureFile, 0)
    fileinfos, err := ioutil.ReadDir(foldername)
    if err != nil {
        return nil, err
    }

    for _, fileinfo := range fileinfos {
        if !fileinfo.IsDir() && filepath.Ext(fileinfo.Name()) == ".yml" {
            fixture := &FixtureFile{
                Path: foldername+"/"+fileinfo.Name(),
                FileName: fileinfo.Name(),
            }
            fixture.Content, err = ioutil.ReadFile(fixture.Path)
            if err != nil {
                return nil, err
            }
            files = append(files, fixture)
        }
    }
    return files, nil
}

func LoadFixtures(foldername string, db *sql.DB, h DataBaseHelper) error {
    files, err := getYmlFiles(foldername)
    if err != nil {
        return err
    }

    tx, err := db.Begin()
    if err != nil {
        return err
    }
    h.DisableTriggers(db)

    for _, file := range files {
        err := file.Delete(tx)
        if err != nil {
            tx.Rollback()
            h.EnableTriggers(db)
            return err
        }

        err = file.Insert(tx)
        if err != nil {
            tx.Rollback()
            h.EnableTriggers(db)
            return err
        }
    }
    err = tx.Commit()
    return err
}
