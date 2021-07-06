package main

import (
	"context"
	"flag"
	"math/rand"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/judwhite/go-svc"
	iris "github.com/kataras/iris/v12"
	"github.com/mingcheng/ssr-subscriber/node"
	"github.com/mingcheng/ssr-subscriber/subscriber"
	log "github.com/sirupsen/logrus"
)

var (
	BuildTime = "unknown"
	Version   = "unknown"
)

var (
	configFile string
	IrisApp    *iris.Application
)

type program struct {
	Config      *Configure
	Fetcher     *subscriber.Fetcher
	Subscriber  *subscriber.Subscriber
	RedisClient *redis.Client
}

func (p *program) Init(_ svc.Environment) error {

	if p.Config.RedisAddr != "" {
		p.RedisClient = redis.NewClient(&redis.Options{
			Addr: p.Config.RedisAddr,
		})
	} else {
		log.Warn("redis address is not configured")
	}

	// start fetch and check
	p.Fetcher = &subscriber.Fetcher{
		CheckConfig: p.Config.Check,
		Proxy:       p.Config.Proxy,
		RedisClient: p.RedisClient,
	}

	if err := p.Fetcher.Init(); err != nil {
		return err
	}

	p.Subscriber = &subscriber.Subscriber{
		Fetcher:  p.Fetcher,
		Sources:  append(p.Config.File, p.Config.URL...),
		Interval: time.Duration(p.Config.Interval) * time.Minute,
	}

	return nil
}

func (p *program) Start() error {
	if err := p.Subscriber.Start(); err != nil {
		log.Error(err)
		return err
	}

	IrisApp = iris.New()

	IrisApp.Get("/random/{type:string}", func(ctx iris.Context) {
		var configs []node.Config
		log.Debugf("get all %d configs", len(p.Fetcher.Configs))

		for _, v := range p.Fetcher.Configs {
			configs = append(configs, v)
		}

		if len(configs) == 0 {
			ctx.StatusCode(http.StatusNotFound)
			return
		}

		// initialize global pseudo random generator
		rand.Seed(time.Now().Unix())
		config := configs[rand.Intn(len(configs))]
		log.Trace(config)

		ctx.StatusCode(http.StatusOK)
		switch ctx.Params().Get("type") {
		case "yaml", "yml":
			_, _ = ctx.YAML(config)
		default:
			_, _ = ctx.JSON(config)
		}
	})

	IrisApp.Get("/all", func(ctx iris.Context) {
		ctx.StatusCode(http.StatusOK)
		_, _ = ctx.JSON(p.Fetcher.Configs)
	})

	IrisApp.Handle("GET", "/last-check-time", func(ctx iris.Context) {
		ctx.StatusCode(http.StatusOK)
		_, _ = ctx.WriteString(p.Subscriber.FetchTimestamp.String())
	})

	go IrisApp.Run(iris.Addr(p.Config.Bind))
	return nil
}

func (p *program) Stop() error {
	if err := p.Subscriber.Stop(context.Background()); err != nil {
		log.Error(err)
		return err
	}

	if p.RedisClient != nil {
		if err := p.RedisClient.Close(); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	flag.StringVar(&configFile, "config", "/etc/ssr-subscriber.yml", "subscribe configure file")
	flag.Usage = func() {
		log.Printf("version v%s(%s)", Version, BuildTime)
		os.Exit(0)
	}
	flag.Parse()
}

func main() {
	// parse configure from local file
	config, err := ParseConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := svc.Run(&program{
		Config: config,
	}, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}
