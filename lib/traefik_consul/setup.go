package traefikConsul

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var initializedFlag int32
var finilizer func()

func setInitFlag() {
	atomic.AddInt32(&initializedFlag, 1)
}

func checkInitFlag() bool {
	return atomic.LoadInt32(&initializedFlag) != 0
}

var initOnce sync.Once

// Initialize initializes configuration for traefik on consul
// retry is the number of retrying. if retry<0, this won't stop retrying until success.
func Initialize(retry int32) error {
	var err error
	initOnce.Do(func() {
		err = initialize(retry)
	})

	return err
}
func initialize(retry int32) error {
	addr := os.Getenv("PP_CONSUL_ADDR")
	prefix := os.Getenv("PP_TRAEFIK_PREFIX")
	if len(prefix) == 0 {
		prefix = "traefik"
	}

	client, err := NewClient(prefix, addr)
	if err != nil {
		return errors.New("traefikConsul.NewClient() error:" + err.Error())
	}

	var advertiseAddr string
	if addr := os.Getenv("PP_ADVERTISE_ADDR"); len(addr) != 0 {
		u, err := url.Parse(addr)
		if err != nil {
			logrus.WithError(err).Errorf("URL format is illegal. The default one will be used.", addr)
		} else {
			if u.Scheme == "" {
				u.Scheme = "http"
			}
			advertiseAddr = u.String()
		}
	} else if iface := os.Getenv("PP_IFACE"); len(iface) != 0 {
		addr, err := IPAddressFromIface(iface)

		if err != nil {
			logrus.WithError(err).Errorf("Network interface(%s) was not found. The default one will be used.", iface)
		} else {
			advertiseAddr = "http://" + addr
		}

	}

	if len(advertiseAddr) == 0 {
		addr, err := DefaultIPAddress()

		if err != nil {
			return errors.New("traefikConsul.DefaultIPAddress() error" + err.Error())
		}

		advertiseAddr = "http://" + addr
	}

	backend := os.Getenv("PP_TRAEFIK_BACKEND")
	if len(backend) == 0 {
		backend = "backend1"
	}

	for {
		if has, err := client.HasFrontend(); err != nil {
			if retry != 0 {
				retry--
				continue
			}
			return errors.New("consul-manager.HasFrontend() error: " + err.Error())
		} else {
			if !has {
				if err := client.NewFrontend("frontend1" /*default frontend*/, backend); err != nil {
					logrus.WithError(err).Error("consul-manager.NewFrontend() error")
				}
			}
		}
	}

	host, _ := os.Hostname()
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	rgen := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, 6)
	for i := range result {
		result[i] = chars[rgen.Intn(len(chars))]
	}
	serverName := "ppweb-" + host + "-" + string(result)

	if b, err := ioutil.ReadFile("./traefik_config_backup"); err == nil {
		if err := client.RestoreBackup(backend, serverName, b); err != nil {
			logrus.WithError(err).Error("RestoreBackup() error")
		}
	}
	var wg sync.WaitGroup
	once := sync.Once{}
	shutdown := func() {
		wg.Wait()
		backup, err := client.BackupBackend(backend, serverName)

		if err != nil {
			logrus.WithError(err).Error("BackupBackend() error")

			return
		}

		ioutil.WriteFile("./traefik_config_backup", backup, 0777)

		if err := client.DeleteBackend(backend, serverName); err != nil {
			logrus.WithError(err).Error("BackupBackend() error")

			return
		}
	}
	finilizer = func() {
		once.Do(shutdown)
	}

	wg.Add(1)
	go func() {
		if err := client.RegisterNewBackend(backend, serverName, advertiseAddr); err != nil {
			logrus.WithError(err).Error("RegisterNewBackend() error")
		}

		wg.Done()
	}()

	return nil
}

func Finalize() {
	if checkInitFlag() {
		finilizer()
	}
}
