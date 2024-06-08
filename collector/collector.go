package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/devon-mar/tacacs-exporter/config"

	"github.com/nwaples/tacplus"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "tacacs"
)

type Collector struct {
	Target     string
	Module     *config.Module
	remoteAddr string
	duration   prometheus.Gauge
	statusCode prometheus.Gauge
	success    prometheus.Gauge
}

func NewCollector(target string, remoteAddress string, module *config.Module) Collector {
	return Collector{
		Target:     target,
		Module:     module,
		remoteAddr: remoteAddress,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "scrape_duration_seconds",
			Help:      "TACACS response time in seconds.",
		}),
		statusCode: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "status_code",
			Help:      "TACACS Authentication reply status code. Common values are Pass(1) and Fail(2).",
		}),
		success: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "success",
			Help:      "1 if the TACACS probe was successful.",
		}),
	}
}

// Describe implements prometheus.Collector
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	c.duration.Describe(ch)
	c.statusCode.Describe(ch)
	c.success.Describe(ch)
}

// Collect implements prometheus.Collector
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	if err := c.probe(); err != nil {
		slog.Error("Error sending TACACS request", "err", err, "target", c.Target)
		c.success.Set(0)
	} else {
		c.success.Set(1)
	}
	c.duration.Collect(ch)
	c.statusCode.Collect(ch)
	c.success.Collect(ch)
}

func (c Collector) probe() error {
	client := tacplus.Client{
		Addr: c.Target,
		ConnConfig: tacplus.ConnConfig{
			Mux:          c.Module.SingleConnect,
			LegacyMux:    c.Module.LegacySingleConnect,
			Secret:       c.Module.Secret,
			ReadTimeout:  c.Module.Timeout,
			WriteTimeout: c.Module.Timeout,
		},
	}

	logFields := slog.With("target", c.Target, "username", c.Module.Username)

	ctx, cancel := context.WithTimeout(context.Background(), c.Module.Timeout)
	defer cancel()

	authenStart := tacplus.AuthenStart{
		Action:        tacplus.AuthenActionLogin,
		PrivLvl:       0,
		AuthenType:    tacplus.AuthenTypePAP,
		AuthenService: tacplus.AuthenServiceLogin,
		User:          c.Module.Username,
		Port:          c.Module.Port,
		RemAddr:       c.remoteAddr,
		Data:          []byte(c.Module.Password),
	}

	begin := time.Now()
	authReply, session, err := client.SendAuthenStart(ctx, &authenStart)
	c.duration.Set(time.Since(begin).Seconds())
	if err != nil {
		logFields.Error("Error sending tacacs authentication start", "err", err)
		return err
	}
	logFields.Debug("reply received message", "message", authReply.ServerMsg)

	if session != nil {
		session.Close()
	}

	c.statusCode.Set((float64)(authReply.Status))
	return nil
}
