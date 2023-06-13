package mimir

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
	"log"
	"time"
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

func (c *Cron) lock(ctx context.Context, h *Handler) error {
	schedule, err := cron.Parse(h.cron)
	if err != nil {
		return err
	}

	now := time.Now()
	d := schedule.Next(now).Sub(now)
	d = d + time.Duration(float64(d)*defaultLagFactor)

	success, err := c.redisClient.SetNX(ctx, c.key(h), "locked", d).Result()
	if err != nil {
		return err
	}

	if !success {
		return ErrAlreadyLocked
	}
	return nil
}

func (c *Cron) unlock(ctx context.Context, h *Handler) error {
	return c.redisClient.Del(ctx, c.key(h)).Err()
}

func (c *Cron) key(h *Handler) string {
	return fmt.Sprintf("%s/%s", c.mutexConf.Prefix, h.fName)
}

func wrapperHandle(c *Cron, h *Handler) func() {
	return func() {
		ctx := context.Background()

		// acquire lock
		err := c.lock(ctx, h)
		if err != nil {
			if errors.Is(err, ErrAlreadyLocked) {
				// lock acquired by another instance
				log.Printf("job: %s running on another instance", h.fName)
			} else {
				log.Printf("job: %s error acquiring lock, err : %v", h.fName, err)
			}
		}
		// defer lock release
		defer func(c *Cron, ctx context.Context, h *Handler) {
			err := c.unlock(ctx, h)
			if err != nil {
				log.Printf("job : %s errored releasing lock, err : %v", h.fName, err)
			}
		}(c, ctx, h)

		// run job
		if err := h.run(ctx); err != nil {
			log.Printf("error running job : %s, err : %v", h.fName, err)
		}
		return
	}
}
