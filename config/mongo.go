package config

import (
	"log"

	"gopkg.in/mgo.v2"
)

// MongoDB holds the information required for connecting
// to a MongoDB replicaset.
type MongoDB struct {
	User       string `envconfig:"MONGODB_USER"`
	Pw         string `envconfig:"MONGODB_PW"`
	Hosts      string `envconfig:"MONGODB_HOSTS"`
	MasterHost string `envconfig:"MONGODB_MASTER_HOST_NAME"`
	AuthDB     string `envconfig:"MONGODB_AUTH_DB_NAME"`
	DB         string `envconfig:"MONGODB_DB_NAME"`
}

// Must will attempt to initiate a new mgo.Session
// with the replicaset and will panic if it encounters any issues.
func (m *MongoDB) Must() *mgo.Session {
	return m.must(m.Hosts)
}

// MustMaster will attempt to initiate a new mgo.Session
// with the Master host and will panic if it encounters any issues.
func (m *MongoDB) MustMaster() *mgo.Session {
	return m.must(m.MasterHost)
}

func (m *MongoDB) must(host string) *mgo.Session {
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

	return s
}

// LoadMongoDBFromEnv will attempt to load a MongoCreds object
// from environment variables. If not populated, nil
// is returned.
func LoadMongoDBFromEnv() *MongoDB {
	var mongo MongoDB
	LoadEnvConfig(&mongo)
	if mongo.Hosts == "" {
		return nil
	}
	return &mongo
}
