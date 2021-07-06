package subscriber

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mingcheng/ssr-subscriber/node"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

const cacheKey = "ssr:subscriber"

type Fetcher struct {
	Configs     map[string]node.Config
	Proxy       string
	CheckConfig node.CheckConfig

	RedisClient *redis.Client
}

func (f *Fetcher) Init() error {
	f.Configs = make(map[string]node.Config)

	if f.RedisClient != nil {
		status := f.RedisClient.Ping(context.Background())
		_, err := status.Result()
		if err != nil {
			return err
		}
	}

	return nil
}

// FromURL fetch and parse configs from url
func (f *Fetcher) FromURL(url string) error {
	httpTransport := &http.Transport{}
	client := &http.Client{Transport: httpTransport, Timeout: 5 * time.Second}

	if len(f.Proxy) > 0 {
		log.Printf("Using socks5 proxy address %s", f.Proxy)
		dialer, err := proxy.SOCKS5("tcp", f.Proxy, nil, proxy.Direct)
		if err != nil {
			return err
		}

		httpTransport.DialContext = func(_ context.Context, network, addr string) (conn net.Conn, e error) {
			return dialer.Dial(network, addr)
		}
	}

	response, err := client.Get(url)
	if err != nil || response.StatusCode != http.StatusOK {
		return errors.New("request subscribe url error")
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return f.FromBytes(data)
}

// FromFile parse configs from local base64-hashed file
func (f *Fetcher) FromFile(path string) error {
	stat, err := os.Stat(path)
	if err != nil || !stat.Mode().IsRegular() {
		return errors.New("not a regular file")
	}

	fd, err := os.OpenFile(path, os.O_RDONLY, os.ModeTemporary)
	if err != nil {
		return err
	}

	defer fd.Close()
	return f.FromReader(bufio.NewReader(fd))
}

// FromString parse from string
func (f *Fetcher) FromString(data string) error {
	var (
		err     error
		decoded []byte
	)

	if decoded, err = base64.StdEncoding.DecodeString(data); err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(data)
	}

	if err != nil || len(decoded) <= 0 {
		return err
	}

	configs, err := decode(string(decoded))
	if err != nil {
		return err
	}

	for _, config := range configs {
		f.Configs[config.ID] = config
	}

	return nil
}

// FromReader from steam reader
func (f *Fetcher) FromReader(r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return f.FromBytes(data)
}

// FromBytes get configs from bytes array
func (f *Fetcher) FromBytes(data []byte) error {
	return f.FromString(string(data))
}

func (f *Fetcher) Check() error {
	for k, config := range f.Configs {
		if err := config.Check(f.CheckConfig); err != nil {
			log.Warnf("clean config which is not healthy, %s:%d(%s)", config.Server, config.ServerPort, config.Remarks)
			delete(f.Configs, k)

			if f.RedisClient != nil {
				log.Warnf("delete config with the key %s from redis", k)
				f.RedisClient.HDel(context.Background(), cacheKey, k)
			}
		} else {
			f.Configs[k] = config // update check timestamp
		}
	}

	return nil
}

// Save configs to redis
func (f *Fetcher) Save(ctx context.Context) error {
	if f.RedisClient == nil {
		return fmt.Errorf("redis client is not initized, so can not save")
	}

	for k, config := range f.Configs {
		marshal, err := json.Marshal(config)
		if err != nil {
			return err
		}

		err = f.RedisClient.HSet(ctx, cacheKey, k, marshal).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// Restore cached configs from redis
func (f *Fetcher) Restore(ctx context.Context) error {
	if f.RedisClient == nil {
		return fmt.Errorf("redis client is not initized")
	}

	cmd := f.RedisClient.HKeys(ctx, cacheKey)

	if cmd.Err() != nil {
		return cmd.Err()
	}

	keys, err := cmd.Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		status := f.RedisClient.HGet(ctx, cacheKey, key)
		if data, err := status.Result(); err != nil {
			log.Error(err)
		} else {
			var config node.Config
			err = json.Unmarshal([]byte(data), &config)
			if err != nil {
				log.Error(err)
			}

			f.Configs[key] = config
		}
	}

	log.Infof("restored %d configs", len(f.Configs))
	return nil
}
