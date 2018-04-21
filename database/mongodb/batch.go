package mongodb

import (
  "context"
	"fmt"

	"github.com/ohsu-comp-bio/ktl"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)


func (db *MongoDB) CreateBatch(ctx context.Context, b *ktl.Batch) error {
  return db.batches.Insert(b)
}

func (db *MongoDB) UpdateBatch(ctx context.Context, b *ktl.Batch) error {
  return db.batches.Update(bson.M{"id": b.ID}, b)
}

func (db *MongoDB) GetBatch(ctx context.Context, id string) (*ktl.Batch, error) {
	q := db.batches.Find(bson.M{"id": id})
  batch := &ktl.Batch{}
	err := q.One(batch)
  if err == mgo.ErrNotFound {
    return nil, ktl.ErrNotFound
  }
	return batch, err
}

// ListBatches returns a list of taskIDs
func (db *MongoDB) ListBatches(ctx context.Context, opts *ktl.BatchListOptions) ([]*ktl.Batch, error) {

  query := bson.M{}
	if opts.Page != "" {
		query["id"] = bson.M{"$lt": opts.Page}
	}

	for k, v := range opts.Tags {
		query[fmt.Sprintf("tags.%s", k)] = bson.M{"$eq": v}
	}

	q := db.batches.Find(query).Sort("-createdAt").Limit(opts.GetPageSize())

	var batches []*ktl.Batch
	err := q.All(&batches)
  return batches, err
}
