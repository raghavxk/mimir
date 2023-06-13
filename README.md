# Mimir : Distributed Cron Scheduler

## Description
- A simple distributed cron scheduler written in Golang
- allows running cron jobs across instances without duplication

## Inspiration:
- At time of writing this, I couldn't find a library suited to a cron run spec
- An existing solution had similar capabilities, but lacked a specific use case
- Use case required running only a single instance of cron job, even in cases when preceding cron jobs lags and overlaps with succeeding cron job schedule

## Authors:
- Raghavendra Khare | [@raghavxk](https://twitter.com/raghavxk)


## Thanks:
- [Cron - cron lib](https://github.com/robfig/cron)
- [Lite-cron](https://github.com/imiskolee/litecron)