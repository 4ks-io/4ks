package controllers

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub/v2"
	pubsubpb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"cloud.google.com/go/storage"
	"github.com/typesense/typesense-go/typesense"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const typesenseProbeTimeout = 2 * time.Second

// firestoreProber checks Firestore connectivity by fetching a sentinel document.
// codes.NotFound means the connection is healthy but the document doesn't exist.
type firestoreProber struct {
	client *firestore.Client
}

// NewFirestoreProber returns a Prober that checks Firestore connectivity.
func NewFirestoreProber(client *firestore.Client) Prober {
	return &firestoreProber{client: client}
}

func (p *firestoreProber) Name() string { return "firestore" }

func (p *firestoreProber) Probe(ctx context.Context) error {
	_, err := p.client.Collection("_readiness").Doc("_ping").Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil
	}
	return err
}

// typesenseProber checks Typesense connectivity via its health endpoint.
// The typesense client manages its own timeout internally.
type typesenseProber struct {
	client *typesense.Client
}

// NewTypesenseProber returns a Prober that checks Typesense connectivity.
func NewTypesenseProber(client *typesense.Client) Prober {
	return &typesenseProber{client: client}
}

func (p *typesenseProber) Name() string { return "typesense" }

func (p *typesenseProber) Probe(ctx context.Context) error {
	ok, err := p.client.Health(ctx, typesenseProbeTimeout)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("unhealthy")
	}
	return nil
}

// pubsubProber checks PubSub connectivity by verifying a topic exists.
type pubsubProber struct {
	client  *pubsub.Client
	topicID string
}

// NewPubSubProber returns a Prober that checks PubSub connectivity for the given topic.
func NewPubSubProber(client *pubsub.Client, topicID string) Prober {
	return &pubsubProber{client: client, topicID: topicID}
}

func (p *pubsubProber) Name() string { return "pubsub" }

func (p *pubsubProber) Probe(ctx context.Context) error {
	topicName := fmt.Sprintf("projects/%s/topics/%s", p.client.Project(), p.topicID)
	_, err := p.client.TopicAdminClient.GetTopic(ctx, &pubsubpb.GetTopicRequest{Topic: topicName})
	if status.Code(err) == codes.NotFound {
		return fmt.Errorf("topic %q not found", p.topicID)
	}
	return err
}

// storageProber checks Cloud Storage connectivity by fetching bucket attributes.
type storageProber struct {
	client *storage.Client
	bucket string
}

// NewStorageProber returns a Prober that checks Cloud Storage connectivity for the given bucket.
func NewStorageProber(client *storage.Client, bucket string) Prober {
	return &storageProber{client: client, bucket: bucket}
}

func (p *storageProber) Name() string { return "storage" }

// Probe lists at most one object — requires storage.objects.list, not storage.buckets.get.
func (p *storageProber) Probe(ctx context.Context) error {
	it := p.client.Bucket(p.bucket).Objects(ctx, &storage.Query{})
	_, err := it.Next()
	if err == iterator.Done {
		return nil // empty bucket is fine
	}
	return err
}
