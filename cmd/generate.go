/*
Package cmd

Copyright © 2021 Joe Glista <josephsglista@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/tools/imports"
)

const (
	mssqlVarChar  = "varchar"
	mssqlNVarChar = "nvarchar"
	mssqlChar     = "char"
	mssqlInt      = "int"
	mssqlFloat    = "float"
	mssqlBit      = "bit"
	mssqlTime     = "datetime"
	mssqlBinary   = "binary"
)

type DatabaseExistsRow struct {
	DatabaseExists bool `db:"DatabaseExists"`
}

type TableExistsRow struct {
	TableExists bool `db:"TableExists"`
}

type Columns struct {
	TableCatalog           string  `db:"TABLE_CATALOG"`
	TableSchema            string  `db:"TABLE_SCHEMA"`
	TableName              string  `db:"TABLE_NAME"`
	ColumnName             string  `db:"COLUMN_NAME"`
	OrdinalPosition        int     `db:"ORDINAL_POSITION"`
	ColumnDefault          *string `db:"COLUMN_DEFAULT"`
	IsNullable             string  `db:"IS_NULLABLE"`
	DataType               string  `db:"DATA_TYPE"`
	CharacterMaximumLength *int    `db:"CHARACTER_MAXIMUM_LENGTH"`
	CharacterOctetLength   *int    `db:"CHARACTER_OCTET_LENGTH"`
	NumericPrecision       *int    `db:"NUMERIC_PRECISION"`
	NumericPrecisionRadix  *int    `db:"NUMERIC_PRECISION_RADIX"`
	NumericScale           *int    `db:"NUMERIC_SCALE"`
	DateTimePrecision      *int    `db:"DATETIME_PRECISION"`
	CharacterSetCatalog    *string `db:"CHARACTER_SET_CATALOG"`
	CharacterSetSchema     *string `db:"CHARACTER_SET_SCHEMA"`
	CharacterSetName       *string `db:"CHARACTER_SET_NAME"`
	CollationCatalog       *string `db:"COLLATION_CATALOG"`
	CollationSchema        *string `db:"COLLATION_SCHEMA"`
	CollationName          *string `db:"COLLATION_NAME"`
	DomainCatalog          *string `db:"DOMAIN_CATALOG"`
	DomainSchema           *string `db:"DOMAIN_SCHEMA"`
	DomainName             *string `db:"DOMAIN_NAME"`
}

//go:embed templates/*
var generateTmpl embed.FS

var mssqlTypeMap = map[string]string{
	mssqlVarChar:  "string",
	mssqlNVarChar: "string",
	mssqlChar:     "string",
	mssqlInt:      "int64",
	mssqlFloat:    "float64",
	mssqlBit:      "bool",
	mssqlTime:     "time.Time",
	mssqlBinary:   "[]byte",
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate reads the structure of your database objects and generates go types.",
	Long: `Using generate, you can lookup obejcts in your database and generate a Go type representing
a table or a view. For example:

sqlizer generate -d {YourDatabase} -t {YourTable}
	`,
	Run: func(cmd *cobra.Command, args []string) {
		rows, err := readTable(cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "While parsing database table: %s", err.Error())
			return
		}

		fmtBytes, err := generateTypes(rows)
		if err != nil {
			fmt.Fprintf(os.Stderr, "While generating code: %s", err.Error())
			return
		}

		err = writeTypes(cmd, fmtBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "While writing generated code: %s", err.Error())
			return
		}
	},
}

func checkDatabaseExists(db *sqlx.DB, cmd *cobra.Command) (bool, error) {
	var dbExists []DatabaseExistsRow

	err := db.Select(
		&dbExists,
		`
			IF DB_ID(?) IS NOT NULL
			BEGIN
				SELECT 1 AS DatabaseExists;
			END
			ELSE
			BEGIN
				SELECT 0 AS DatabaseExists;
			END
		`,
		cmd.Flag("database").Value.String(),
	)
	if err != nil {
		return false, err
	}
	if !dbExists[0].DatabaseExists {
		return false, nil
	}

	return true, nil
}

func checkTableExists(db *sqlx.DB, cmd *cobra.Command) (bool, error) {
	var tableExists []TableExistsRow

	err := db.Select(
		&tableExists,
		`
			IF (
				EXISTS (
					SELECT *
					FROM UserManagement.INFORMATION_SCHEMA.TABLES
					WHERE TABLE_NAME = ?
				)
			)
			BEGIN
				SELECT 1 AS TableExists;
			END
			ELSE
			BEGIN
				SELECT 0 AS TableExists;
			END
		`,
		cmd.Flag("table").Value.String(),
	)
	if err != nil {
		return false, err
	}
	if !tableExists[0].TableExists {
		return false, nil
	}

	return true, nil
}

func readTable(cmd *cobra.Command) ([]Columns, error) {
	u := url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(viper.GetString("server.user"), viper.GetString("server.pass")),
		Host:   fmt.Sprintf("%s:%s", viper.GetString("server.host"), viper.GetString("server.port")),
	}

	db, err := sqlx.Connect("mssql", u.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "While connecting to database server: %s", err.Error())
		return nil, err
	}
	defer db.Close()

	dbExists, err := checkDatabaseExists(db, cmd)
	if err != nil {
		return nil, err
	}
	if !dbExists {
		return nil, errors.New("databaseG does not exist")
	}

	tableExists, err := checkTableExists(db, cmd)
	if err != nil {
		return nil, err
	}
	if !tableExists {
		return nil, errors.New("table does not exist")
	}

	t := cmd.Flag("table").Value.String()
	var rows []Columns
	// TODO: Don't use string building to inject the database name, this is a SQL injection risk.
	err = db.Select(
		&rows,
		fmt.Sprintf("SELECT * FROM %s.INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME=?", cmd.Flag("database").Value.String()),
		t,
	)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func generateTypes(rows []Columns) ([]byte, error) {
	tmpl, err := template.New("generate.tmpl").Funcs(template.FuncMap{
		"ToLower": strings.ToLower,
	}).ParseFS(generateTmpl, "templates/generate.tmpl")
	if err != nil {
		return nil, err
	}

	tmplBytes := []byte{}
	tmplBuff := bytes.NewBuffer(tmplBytes)
	err = tmpl.Execute(tmplBuff, parseResults(rows))
	if err != nil {
		return nil, err
	}

	fmtBytes, err := format.Source(tmplBuff.Bytes())
	if err != nil {
		return nil, err
	}

	return fmtBytes, nil
}

func writeTypes(cmd *cobra.Command, fileBytes []byte) error {
	tLower := strings.ToLower(cmd.Flag("table").Value.String())
	err := os.Mkdir(tLower, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "While writing directory: %s", err.Error())
		return err
	}

	filePath := path.Join(tLower, tLower+".go")
	err = os.WriteFile(filePath, fileBytes, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "While writing generated code: %s", err.Error())
		return err
	}

	genFileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "While opening file for imports generation: %s", err.Error())
		return err
	}

	formatted, err := imports.Process(tLower+".go", genFileBytes, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "While formatting imports: %s", err.Error())
		return err
	}
	os.WriteFile(filePath, formatted, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "While writing formatted file: %s", err.Error())
		return err
	}

	return nil
}

type ParsedColumns struct {
	TableName  string
	Attributes []Attribute
}

type Attribute struct {
	Name string
	Type string
}

func parseResults(c []Columns) ParsedColumns {
	pc := ParsedColumns{}

	for _, col := range c {
		pc.TableName = col.TableName

		attr := Attribute{Name: col.ColumnName}
		if val, ok := mssqlTypeMap[col.DataType]; ok {
			attr.Type = val
		}
		pc.Attributes = append(pc.Attributes, attr)
	}

	return pc
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("database", "d", "", "the database name")
	generateCmd.Flags().StringP("table", "t", "", "the target table")
	generateCmd.MarkFlagRequired("database")
	generateCmd.MarkFlagRequired("table")
}
