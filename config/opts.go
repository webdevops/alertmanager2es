package config

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Debug   bool `           long:"debug"        env:"DEBUG"    description:"debug mode"`
			Verbose bool `short:"v"  long:"verbose"      env:"VERBOSE"  description:"verbose mode"`
			LogJson bool `           long:"log.json"     env:"LOG_JSON" description:"Switch log output to json format"`
		}

		// elasticsearch
		Elasticsearch struct {
			// ElasticSearch settings
			Addresses []string `long:"elasticsearch.address"      env:"ELASTICSEARCH_ADDRESS"  delim:" "  description:"ElasticSearch urls" required:"true"`
			Username  string   `long:"elasticsearch.username"     env:"ELASTICSEARCH_USERNAME"            description:"ElasticSearch username for HTTP Basic Authentication"`
			Password  string   `long:"elasticsearch.password"     env:"ELASTICSEARCH_PASSWORD"            description:"ElasticSearch password for HTTP Basic Authentication" json:"-"`
			ApiKey    string   `long:"elasticsearch.apikey"       env:"ELASTICSEARCH_APIKEY"              description:"ElasticSearch base64-encoded token for authorization; if set, overrides username and password" json:"-"`
			Index     string   `long:"elasticsearch.index"        env:"ELASTICSEARCH_INDEX"               description:"ElasticSearch index name (placeholders: %y for year, %m for month and %d for day)" default:"alertmanager-%y.%m"`
		}

		// general options
		ServerBind string `long:"bind"     env:"SERVER_BIND"   description:"Server address"     default:":9097"`
	}
)

func (o *Opts) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		log.Panic(err)
	}
	return jsonBytes
}
