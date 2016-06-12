package mongodb

import (
	"log"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/NYTimes/gizmo/config"
)

// Config holds the information required for connecting
// to a MongoDB replicaset.
type Config struct {
	User       string         `envconfig:"MONGODB_USER"`
	Pw         string         `envconfig:"MONGODB_PW"`
	Hosts      string         `envconfig:"MONGODB_HOSTS"`
	MasterHost string         `envconfig:"MONGODB_MASTER_HOST_NAME"`
	AuthDB     string         `envconfig:"MONGODB_AUTH_DB_NAME"`
	DB         string         `envconfig:"MONGODB_DB_NAME"`
	Mode       string         `envconfig:"MONGODB_MODE"`
	Tags       []bson.DocElem `envconfig:"MONGODB_TAGS"`
}

// Must will attempt to initiate a new mgo.Session
// with the replicaset and will panic if it encounters any issues.
func (m *Config) Must() *mgo.Session {
	return m.must(m.Hosts)
}

// MustMaster will attempt to initiate a new mgo.Session
// with the Master host and will panic if it encounters any issues.
func (m *Config) MustMaster() *mgo.Session {
	return m.must(m.MasterHost)
}

func (m *Config) must(host string) *mgo.Session {
	s, err := mgo.Dial(host)
	if err != nil {
		log.Fatal(err)
	}

	if m.User != "" {
		db := s.DB(m.AuthDB)
		err = db.Login(m.User, m.Pw)
		if err != nil {
			log.Fatal(err)
		}
	}

	m.setMode(s)
	m.setSelectServers(s)

	return s
}

// LoadConfigFromEnv will attempt to load a MongoCreds object
// from environment variables.
func LoadConfigFromEnv() *Config {
	var mongo Config
	config.LoadEnvConfig(&mongo)
	return &mongo
}

func (m *Config) setMode(s *mgo.Session) {
	if m.Mode == "" {
		return
	}

	var mode mgo.Mode

	switch strings.ToLower(m.Mode) {
	case "primary":
		mode = mgo.Primary
	case "primarypreferred":
		mode = mgo.PrimaryPreferred
	case "secondary":
		mode = mgo.Secondary
	case "secondarypreferred":
		mode = mgo.SecondaryPreferred
	case "nearest":
		mode = mgo.Nearest
	case "eventual":
		mode = mgo.Eventual
	case "monotonic":
		mode = mgo.Monotonic
	case "strong":
		mode = mgo.Strong
	default:
		return
	}

	s.SetMode(mode, false)
}

func (m *Config) setSelectServers(s *mgo.Session) {
	if len(m.Tags) == 0 {
		return
	}

	s.SelectServers(bson.D(m.Tags))
	s.Refresh()
}
