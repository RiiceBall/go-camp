package ioc

import (
	events2 "gitee.com/geekbang/basic-go/webook/interactive/events"
	"gitee.com/geekbang/basic-go/webook/interactive/repository/dao"
	"gitee.com/geekbang/basic-go/webook/internal/events"
	"gitee.com/geekbang/basic-go/webook/pkg/migrator/events/fixer"
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

func InitSaramaClient() sarama.Client {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true
	client, err := sarama.NewClient(cfg.Addr, scfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSaramaSyncProducer(client sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return p
}

func InitConsumers(c1 *events2.InteractiveReadEventConsumer, fixConsumer *fixer.Consumer[dao.Interactive]) []events.Consumer {
	return []events.Consumer{c1, fixConsumer}
}
