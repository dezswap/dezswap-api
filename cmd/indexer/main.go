package main

import (
	"fmt"
	"github.com/dezswap/dezswap-api/pkg"
	"math"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dezswap/dezswap-api/configs"
	"github.com/dezswap/dezswap-api/indexer"
	"github.com/dezswap/dezswap-api/indexer/repo"
	"github.com/dezswap/dezswap-api/pkg/logging"
	"github.com/dezswap/dezswap-api/pkg/xpla"
	"github.com/go-co-op/gocron"
)

type repeatableJob struct {
	each         func() error
	errorHandler func(err error)
	delay        time.Duration
	errCount     uint
	tolerance    uint
	exponential  bool
}

func runJob(j *repeatableJob, logger logging.Logger) {
	fName := runtime.FuncForPC(reflect.ValueOf(j.each).Pointer()).Name()
	logger.Info(fmt.Sprintf("job(%s) datetime(%s)", fName, time.Now().String()))

	start := time.Now()
	err := j.each()
	elapsed := time.Since(start)
	logger.Debugf(fmt.Sprintf("Binomial took %ds, delay: %ds", elapsed/time.Second, j.delay/time.Second))

	if err != nil {
		j.errCount++
		logger.Error(err)
		if j.errorHandler != nil {
			j.errorHandler(err)
		}

		wait := j.delay
		if j.exponential {
			wait = j.delay * time.Duration(math.Pow(2, float64(j.errCount)))
		}
		time.Sleep(wait)
	} else {
		j.errCount = 0
	}

	if j.errCount == j.tolerance {
		panic(err)
	}
}

func main() {
	c := configs.New()
	c.Log.ChainId = c.Indexer.ChainId
	logger := setLogger(c)
	defer catch(logger)

	assetRepo := repo.NewAssetRepo(xpla.NewClient())
	dbRepo, err := repo.NewDbRepo(c.Indexer.ChainId, c.Indexer.SrcDb, c.Indexer.Db)
	if err != nil {
		panic(err)
	}
	target := fmt.Sprintf("%s:%s", c.Indexer.SrcNode.Host, c.Indexer.SrcNode.Port)
	grpcCli, err := pkg.NewGrpcClient(target)
	if err != nil {
		panic(err)
	}
	nodeRepo, err := repo.NewNodeRepo(grpcCli, c.Indexer)
	if err != nil {
		panic(err)
	}

	indexerRepo := repo.NewRepo(nodeRepo, dbRepo, assetRepo)
	app := indexer.NewDexIndexer(indexerRepo, c.Indexer.ChainId)

	jobs := []*repeatableJob{
		{each: app.UpdateTokens, errorHandler: nil, delay: xpla.BLOCK_SECOND * time.Second, errCount: 0, tolerance: 3, exponential: true},
		{each: app.UpdateVerifiedTokens, errorHandler: nil, delay: xpla.BLOCK_SECOND * time.Second, errCount: 0, tolerance: 3, exponential: true},
		{each: app.UpdateLatestPools, errorHandler: nil, delay: xpla.BLOCK_SECOND * time.Second, errCount: 0, tolerance: 3, exponential: true},
	}

	logger.Info("Starting indexer...")

	s := gocron.NewScheduler(time.UTC)
	s.SingletonModeAll()

	for _, j := range jobs {
		_, err := s.Every(j.delay).Do(runJob, j, logger)
		if err != nil {
			panic(err)
		}
	}

	s.StartBlocking()
}
func setLogger(c configs.Config) logging.Logger {
	c.Log.ChainId = c.Indexer.ChainId
	logger := logging.New("dezswap-api", c.Log)
	if c.Sentry.DSN != "" {
		if err := logging.ConfigureReporter(logger, c.Sentry.DSN, c.Indexer.ChainId, map[string]string{
			"app":      "dezswap-api-indexer",
			"env":      c.Log.Environment,
			"chain_id": c.Indexer.ChainId,
		}); err != nil {
			panic(err)
		}
	}
	return logger
}

func catch(logger logging.Logger) {
	recovered := recover()

	if recovered != nil {
		defer os.Exit(1)

		err, ok := recovered.(error)
		if !ok {
			logger.Errorf("could not convert recovered error into error: %s\n", spew.Sdump(recovered))
			return
		}

		stack := string(debug.Stack())
		logger.WithField("err", logging.NewErrorField(err)).WithField("stack", stack).Errorf("panic caught")
	}
}
