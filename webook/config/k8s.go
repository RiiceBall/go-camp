//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(webook-mysql-service:3308)/webook",
	},
	Redis: RedisConfig{
		Addr: "webook-redis-service:6379",
	},
}
