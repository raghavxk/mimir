package mimir

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
)

const (
	defaultLagFactor  = 0.5
	defaultCronPrefix = "cron-defaults/"
)

type RunF func(ctx context.Context) error

type Handler struct {
	cron  string
	fName string
	run   RunF
}

func newHandler(cron string, fName string, run RunF) *Handler {
	return &Handler{
		cron:  cron,
		fName: fName,
		run:   run,
	}
}

type (
	RedisConf struct {
		Host     string
		Port     int
		Password string `log:"-"`
	}
	MutexConf struct {
		RedisConf RedisConf
		Prefix    string
		Factor    float64
	}
)

type Cron struct {
	mutexConf   MutexConf
	cronClient  *cron.Cron
	redisClient *redis.Client
}

func NewCron(conf MutexConf, redis *redis.Client) *Cron {
	c := new(Cron)

	// set prefix if missing
	if conf.Prefix == "" {
		conf.Prefix = defaultCronPrefix
	}

	// set lag factor if missing
	if conf.Factor <= 0 {
		conf.Factor = defaultLagFactor
	}

	c.mutexConf = conf
	c.redisClient = redis
	c.cronClient = cron.New()
	return c
}
