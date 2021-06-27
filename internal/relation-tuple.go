package watcher

import (
	"encoding/json"
	"fmt"
	"strings"

	watchpb "github.com/authorizer-tech/access-watcher/gen/go/authorizer-tech/accesswatcher/v1alpha1"
	"github.com/pkg/errors"
)

var ErrInvalidSubjectSetString = fmt.Errorf("the provided SubjectSet string is malformed")
var ErrInvalidRelationTupleString = fmt.Errorf("the RelationTuple string is malformed")

type Subject interface {
	json.Marshaler

	String() string
	FromString(string) (Subject, error)
	Equals(interface{}) bool
	ToProto() *watchpb.Subject
}

type SubjectID struct {
	ID string `json:"id"`
}

func (s SubjectID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

func (s *SubjectID) Equals(v interface{}) bool {
	uv, ok := v.(*SubjectID)
	if !ok {
		return false
	}
	return uv.ID == s.ID
}

func (s *SubjectID) FromString(str string) (Subject, error) {
	s.ID = str
	return s, nil
}

func (s *SubjectID) String() string {
	return s.ID
}

func (s *SubjectID) ToProto() *watchpb.Subject {
	return &watchpb.Subject{
		Ref: &watchpb.Subject_Id{
			Id: s.ID,
		},
	}
}

type SubjectSet struct {
	Namespace string `json:"namespace"`
	Object    string `json:"object"`
	Relation  string `json:"relation"`
}

func (s *SubjectSet) Equals(v interface{}) bool {
	uv, ok := v.(*SubjectSet)
	if !ok {
		return false
	}
	return uv.Relation == s.Relation && uv.Object == s.Object && uv.Namespace == s.Namespace
}

func (s *SubjectSet) String() string {
	return fmt.Sprintf("%s:%s#%s", s.Namespace, s.Object, s.Relation)
}

func (s SubjectSet) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

func (s *SubjectSet) ToProto() *watchpb.Subject {
	return &watchpb.Subject{
		Ref: &watchpb.Subject_Set{
			Set: &watchpb.SubjectSet{
				Namespace: s.Namespace,
				Object:    s.Object,
				Relation:  s.Relation,
			},
		},
	}
}

func (s *SubjectSet) FromString(str string) (Subject, error) {
	parts := strings.Split(str, "#")
	if len(parts) != 2 {
		return nil, errors.WithStack(ErrInvalidSubjectSetString)
	}

	innerParts := strings.Split(parts[0], ":")
	if len(innerParts) != 2 {
		return nil, errors.WithStack(ErrInvalidSubjectSetString)
	}

	s.Namespace = innerParts[0]
	s.Object = innerParts[1]
	s.Relation = parts[1]

	return s, nil
}

// SubjectFromString parses the string s and returns a Subject - either
// a SubjectSet or an explicit SubjectID.
func SubjectFromString(s string) (Subject, error) {
	if strings.Contains(s, "#") {
		return (&SubjectSet{}).FromString(s)
	}
	return (&SubjectID{}).FromString(s)
}

type InternalRelationTuple struct {
	Namespace string  `json:"namespace"`
	Object    string  `json:"object"`
	Relation  string  `json:"relation"`
	Subject   Subject `json:"subject"`
}

// String returns r as a relation tuple in string format.
func (r *InternalRelationTuple) String() string {
	return fmt.Sprintf("%s:%s#%s@%s", r.Namespace, r.Object, r.Relation, r.Subject)
}

// ToProto serializes r in it's equivalent protobuf format.
func (r *InternalRelationTuple) ToProto() *watchpb.RelationTuple {
	return &watchpb.RelationTuple{
		Namespace: r.Namespace,
		Object:    r.Object,
		Relation:  r.Relation,
		Subject:   r.Subject.ToProto(),
	}
}

func RelationTupleFromString(s string) (*InternalRelationTuple, error) {
	part1 := strings.Split(s, "#")

	if len(part1) < 2 {
		return nil, ErrInvalidRelationTupleString
	}

	part2 := strings.Split(part1[0], ":")
	if len(part2) < 2 {
		return nil, ErrInvalidRelationTupleString
	}

	namespace := part2[0]
	object := part2[1]

	part3 := strings.Split(part1[1], "@")
	if len(part3) < 2 {
		return nil, ErrInvalidRelationTupleString
	}

	relation := part3[0]
	subject, err := SubjectFromString(part3[1])
	if err != nil {
		return nil, err
	}

	tuple := &InternalRelationTuple{
		Namespace: namespace,
		Object:    object,
		Relation:  relation,
		Subject:   subject,
	}
	return tuple, nil
}
