package dao

import "gorm.io/gorm"

// 初始化表
func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Article{},
		&PublishedArticle{},
		&Job{},
	)
}
