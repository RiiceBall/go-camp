package main

import (
	"database/sql"
	"fmt"
	"math/rand"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:root@tcp(localhost:3306)/webook"

	// 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 删除表
	_, err = db.Exec("DELETE FROM interactives WHERE id > 0")
	if err != nil {
		panic(err)
	}

	query := `INSERT INTO interactives (like_cnt, biz_id, biz) VALUES (?, ?, ?)`
	for i := 0; i < 200; i++ {
		// 随机点赞数
		likeCnt := rand.Int() % 1000
		// 插入数据
		_, err := db.Exec(query, likeCnt, i+1, "article")
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("插入完成！")
}
