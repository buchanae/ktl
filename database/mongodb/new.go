package mongodb

import (
  "time"

	"gopkg.in/mgo.v2"
)

func DefaultConfig() *mgo.DialInfo {
  return &mgo.DialInfo{
    Addrs: []string{"localhost:27017"},
    Database: "ktl",
    Timeout: 30 * time.Second,
  }
}

// MongoDB provides access to ktl collections stored in Mongo DB.
type MongoDB struct {
	sess  *mgo.Session
  batches *mgo.Collection
}

// NewMongoDB returns a new MongoDB instance.
func NewMongoDB(conf *mgo.DialInfo) (*MongoDB, error) {
	sess, err := mgo.DialWithInfo(conf)
	if err != nil {
		return nil, err
	}

	db := &MongoDB{
		sess:  sess,
    batches: sess.DB("").C("batches"),
	}
	return db, nil
}

// Init creates tables in MongoDB.
func (db *MongoDB) Init() error {
  /*
	names, err := db.sess.DB("").CollectionNames()
	if err != nil {
		return fmt.Errorf("listing collections: %s", err)
	}

	var tasksFound bool
	var nodesFound bool
	for _, n := range names {
		switch n {
		case "tasks":
			tasksFound = true
		case "nodes":
			nodesFound = true
		}
	}

	if !tasksFound {
		err = db.tasks.Create(&mgo.CollectionInfo{})
		if err != nil {
			return fmt.Errorf("error creating tasks collection in database %s: %v", db.conf.Database, err)
		}

		err = db.tasks.EnsureIndex(mgo.Index{
			Key:        []string{"-id", "-creationtime"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		})
		if err != nil {
			return err
		}
	}

	if !nodesFound {
		err = db.nodes.Create(&mgo.CollectionInfo{})
		if err != nil {
			return fmt.Errorf("error creating nodes collection in database %s: %v", db.conf.Database, err)
		}

		err = db.nodes.EnsureIndex(mgo.Index{
			Key:        []string{"id"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		})
		if err != nil {
			return err
		}
	}
  */

	return nil
}

// Close closes the database session.
func (db *MongoDB) Close() error {
	db.sess.Close()
	return nil
}
