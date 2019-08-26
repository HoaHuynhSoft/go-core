package pllog

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"gopkg.in/sohlich/elogrus.v7"
)

type LogrusLogger struct {
	ElasticHostURL string `long:"log-host-url" description:"the url of elastichsearch database url" env:"LOG_HOST_URL"`
	Sniff          bool   `long:"log-enable-sniff" description:"Enable or disable sniff" env:"LOG_SNIFF"`
	LogIndexPrefix string `long:"log-prefix" description:"the prefix of index name" env:"LOG_INDEX_PREFIX"`
	LogHostName    string `long:"log-host-name" description:"the prefix of index name" env:"LOG_HOST_NAME"`
	Enable         bool   `long:"log-enable" description:"the prefix of index name" env:"LOG_ENABLE"`
	LogLevel       logrus.Level
	IndexNameFunc  func() string
	*logrus.Logger
}

func New() PlLogger {
	logrusLogger := &LogrusLogger{
		LogLevel: logrus.DebugLevel,
	}

	logrusLogger.IndexNameFunc = func() string {
		dt := time.Now()
		return fmt.Sprintf("%s-%s", logrusLogger.LogIndexPrefix, dt.Format("2006-01-02"))
	}
	parser := flags.NewParser(logrusLogger, flags.IgnoreUnknown)
	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}
	if !logrusLogger.Enable {
		return &DefaultLogger{}
	}
	return NewWithRef(logrusLogger)
}

func NewWithRef(logrusLogger *LogrusLogger) PlLogger {
	if !logrusLogger.Enable {
		return &DefaultLogger{}
	}
	log := logrus.New()
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(logrusLogger.ElasticHostURL))
	if err != nil {
		log.Panic(err)
	}

	hook, err := elogrus.NewElasticHookWithFunc(client, logrusLogger.LogHostName, logrusLogger.LogLevel, logrusLogger.IndexNameFunc)
	if err != nil {
		log.Panic(err)
	}
	log.Hooks.Add(hook)

	logrusLogger.Logger = log
	log.Printf("%+v\n", logrusLogger)
	return logrusLogger
}

func (logrusLogger *LogrusLogger) WithFields(fields map[string]interface{}) PlLogentry {
	return logrusLogger.Logger.WithFields(fields)
}

const (
	RequestIDHeaderKey     = "Request-Id"
	CorrelationIDHeaderKey = "Correlation-Id"
	RequestID              = "RequestId"
	CorrelationID          = "CorrelationId"
)

func CreateLogEntryFromContext(ctx context.Context, log PlLogger) PlLogentry {
	fmt.Println("rid", ctx.Value("RequestId").(string))
	return log.WithFields(map[string]interface{}{
		//CorrelationID: ctx.Value(CorrelationID).(string),
		RequestID: ctx.Value("RequestId").(string),
	})
}
