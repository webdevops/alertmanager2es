package main

import (
	"fmt"
	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

const (
	author = "webdevops.io"
)

var (
	argparser    *flags.Parser
	verbose      bool
	daemonLogger *DaemonLogger

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

var opts struct {
	// general settings
	Verbose []bool `long:"verbose" short:"v"  env:"VERBOSE"  description:"verbose mode"`

	// server settings
	ServerBind string `long:"bind"         env:"SERVER_BIND"   description:"Server address" default:":9097"`

	// ElasticSearch settings
	ElasticsearchAddresses []string `long:"elasticsearch.address"      env:"ELASTICSEARCH_ADDRESS"  delim:" "  description:"ElasticSearch urls" required:"true"`
	ElasticsearchUsername  string   `long:"elasticsearch.username"     env:"ELASTICSEARCH_USERNAME"            description:"ElasticSearch username for HTTP Basic Authentication"`
	ElasticsearchPassword  string   `long:"elasticsearch.password"     env:"ELASTICSEARCH_PASSWORD"            description:"ElasticSearch password for HTTP Basic Authentication"`
	ElasticsearchApiKey    string   `long:"elasticsearch.apikey"       env:"ELASTICSEARCH_APIKEY"              description:"ElasticSearch base64-encoded token for authorization; if set, overrides username and password"`
	ElasticsearchIndex     string   `long:"elasticsearch.index"        env:"ELASTICSEARCH_INDEX"               description:"ElasticSearch index name (placeholders: %y for year, %m for month and %d for day)" default:"alertmanager-%y.%m"`
}

func main() {
	initArgparser()

	// set verbosity
	verbose = len(opts.Verbose) >= 1

	// Init logger
	daemonLogger = NewLogger(log.Lshortfile, verbose)
	defer daemonLogger.Close()

	daemonLogger.Infof("starting alertmanager2es v%s (%s; by %v, based on cloudflare/alertmanager2es)", gitTag, gitCommit, author)

	daemonLogger.Infof("Init exporter")
	exporter := &AlertmanagerElasticsearchExporter{}
	exporter.Init()

	cfg := elasticsearch.Config{
		Addresses: opts.ElasticsearchAddresses,
		Username:  opts.ElasticsearchUsername,
		Password:  opts.ElasticsearchPassword,
		APIKey:    opts.ElasticsearchApiKey,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	exporter.ConnectElasticsearch(cfg, opts.ElasticsearchIndex)

	// daemon mode
	daemonLogger.Infof("starting http server on %s", opts.ServerBind)
	startHttpServer(exporter)
}

// init argparser and parse/validate arguments
func initArgparser() {
	argparser = flags.NewParser(&opts, flags.Default)
	_, err := argparser.Parse()

	// check if there is an parse error
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}
}

// start and handle prometheus handler
func startHttpServer(exporter *AlertmanagerElasticsearchExporter) {
	http.HandleFunc("/webhook", http.HandlerFunc(exporter.HttpHandler))
	http.Handle("/metrics", promhttp.Handler())
	daemonLogger.Fatal(http.ListenAndServe(opts.ServerBind, nil))
}
