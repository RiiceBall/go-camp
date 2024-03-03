package ioc

import (
	"webook/internal/repository/dao"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}

	var config Config = Config{
		DSN: "root:root@tcp(localhost:3306)/webook",
	}
	err := viper.UnmarshalKey("db", &config)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open(config.DSN))
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
