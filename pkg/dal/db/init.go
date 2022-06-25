package db

import (
	"github.com/evpeople/softEngineer/pkg/constants"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormopentracing "gorm.io/plugin/opentracing"
)

var DB *gorm.DB

// Init init DB
func Init() {
	var err error
	DB, err = gorm.Open(mysql.Open(constants.MySQLDefaultDSN),
		&gorm.Config{
			PrepareStmt:            true,
			SkipDefaultTransaction: true,
		},
	)
	if err != nil {
		panic(err)
	}

	if err = DB.Use(gormopentracing.New()); err != nil {
		panic(err)
	}

	m := DB.Migrator()
	if !m.HasTable(&User{}) {
		if err = m.CreateTable(&User{}); err != nil {
			panic(err)
		}
	}
	if !m.HasTable(&Car{}) {
		if err = m.CreateTable(&Car{}); err != nil {
			panic(err)
		}
	}
	if !m.HasTable(&Bill{}) {
		if err = m.CreateTable(&Bill{}); err != nil {
			panic(err)
		}
	}
	if !m.HasTable(&PileInfo{}) {
		//m.DropTable(&PileInfo{}) 
		if err = m.CreateTable(&PileInfo{}); err != nil {
			panic(err)
		}
	}
}
