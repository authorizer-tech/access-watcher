package watcher

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	watchpb "github.com/authorizer-tech/access-watcher/gen/go/authorizer-tech/accesswatcher/v1alpha1"
)

type AccessWatcherOption func(*AccessWatcher)

func WithChangelogDatastore(ds ChangelogDatastore) AccessWatcherOption {
	return func(w *AccessWatcher) {
		w.store = ds
	}
}

type AccessWatcher struct {
	watchpb.UnimplementedWatchServiceServer

	store ChangelogDatastore
}

func NewAccessWatcher(opts ...AccessWatcherOption) (*AccessWatcher, error) {

	w := &AccessWatcher{}

	for _, opt := range opts {
		opt(w)
	}

	return w, nil
}

func (w *AccessWatcher) Watch(req *watchpb.WatchRequest, server watchpb.WatchService_WatchServer) error {

	namespaces := req.GetNamespaces()
	snaptoken := req.GetSnaptoken()

	timestamp, err := timestampFromSnaptoken(snaptoken)
	if err != nil {
		return err
	}

	ctx := server.Context()

	iter, err := w.store.GetRelationTupleChanges(ctx, namespaces, timestamp)
	if err != nil {
		return err
	}

	for iter.Next() {
		val, err := iter.Value()
		if err != nil {
			return err
		}

		response, err := toWatchResponse(val)
		if err != nil {
			return err
		}

		if err := server.Send(response); err != nil {
			return err
		}
	}

	return nil
}

func (w *AccessWatcher) Close(ctx context.Context) error {
	return nil
}

// toWatchResponse converts the ChangelogEntry into an appropriate WatchResponse, or
// returns an error if the transformation is unsuccessful.
func toWatchResponse(e *ChangelogEntry) (*watchpb.WatchResponse, error) {

	snaptoken, err := snaptokenFromTimestamp(e.Timestamp)
	if err != nil {
		return nil, err
	}

	var action watchpb.RelationTupleDelta_Action
	switch e.Operation {
	case "INSERT":
		action = watchpb.RelationTupleDelta_INSERT
	case "DELETE":
		action = watchpb.RelationTupleDelta_DELETE
	default:
		action = watchpb.RelationTupleDelta_ACTION_UNSPECIFIED
	}

	r := &watchpb.WatchResponse{
		RelationTupleDelta: &watchpb.RelationTupleDelta{
			Action:        action,
			RelationTuple: e.RelationTuple.ToProto(),
		},
		Snaptoken: snaptoken,
	}

	return r, nil
}

func snaptokenFromTimestamp(t time.Time) (string, error) {

	s := fmt.Sprintf(`{"timestamp": "%s"}`, t)
	bytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	snaptoken := base64.StdEncoding.EncodeToString(bytes)
	return snaptoken, nil
}

func timestampFromSnaptoken(snaptoken string) (time.Time, error) {

	data, err := base64.StdEncoding.DecodeString(snaptoken)
	if err != nil {
		return time.Time{}, err
	}

	var snapshot struct {
		Timestamp time.Time `json:"timestamp"`
	}

	if err := json.Unmarshal(data, &snapshot); err != nil {
		return time.Time{}, err
	}

	return snapshot.Timestamp, nil
}
