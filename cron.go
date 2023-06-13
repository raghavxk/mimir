package mimir

import (
	"context"
	"fmt"
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

func (c *Cron) Register(cronSchedule string, fName string, h RunF) {
	err := c.cronClient.AddFunc(cronSchedule, wrapperHandle(c, newHandler(cronSchedule, fName, h)))
	if err != nil {
		panic(fmt.Sprintf("failed to register cron : %s with error : %v", fName, err))
	}
}

func (c *Cron) Run() {
	c.cronClient.Run()
}

func wrapperHandle(c *Cron, h *Handler) func() {
	return func() {
		ctx := context.Background()

		// lock mutex

		// run
		if err := h.run(ctx); err != nil {
			return
		}
		// unlock mutex
	}
}
