package node

import (
	"fmt"
	"net"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

func (c *Config) Check(checkConfig CheckConfig) error {
	if matched, _ := regexp.MatchString(checkConfig.Not, c.Remarks); matched {
		err := fmt.Errorf("remarks %s not allowed, ignore", c.Remarks)
		log.Warn(err)
		return err
	}

	duration, err := time.ParseDuration(checkConfig.TCPTimeout)
	if err != nil {
		return err
	}

	_, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.Server, c.ServerPort), duration)
	if err != nil {
		return err
	}

	log.Infof("check %s:%d is passed", c.Server, c.ServerPort)
	c.CheckTime = time.Now()

	return nil
}
