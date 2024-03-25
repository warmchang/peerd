// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package events

import (
	"context"
	"testing"

	p2pcontext "github.com/azure/peerd/internal/context"
	"github.com/azure/peerd/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/record"
)

func TestWithContext(t *testing.T) {
	fcs := fake.NewSimpleClientset([]runtime.Object{
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p2pcontext.NodeName,
				Namespace: p2pcontext.Namespace,
				UID:       "test-uid",
			},
		},
	}...)

	cs := &k8s.ClientSet{Interface: fcs, InPod: true}

	ctx, err := WithContext(context.Background(), cs)
	if err != nil {
		t.Fatal(err)
	}

	if ctx == nil {
		t.Fatal("expected context")
	}

	er := FromContext(ctx).(*eventRecorder)
	if er.objRef.Kind != "Pod" {
		t.Errorf("expected kind to be Pod, got %s", er.objRef.Kind)
	}
}

func TestNewRecorderInNode(t *testing.T) {
	fcs := fake.NewSimpleClientset([]runtime.Object{
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: p2pcontext.NodeName,
				UID:  "test-uid",
			},
		},
	}...)

	cs := &k8s.ClientSet{Interface: fcs, InPod: false}

	r, err := NewRecorder(context.Background(), cs)
	if err != nil {
		t.Fatal(err)
	}

	if r == nil {
		t.Fatal("expected event recorder")
	}

	er := r.(*eventRecorder)
	if er.objRef.Kind != "Node" {
		t.Errorf("expected kind to be Node, got %s", er.objRef.Kind)
	}
	if er.objRef.Name != p2pcontext.NodeName {
		t.Errorf("expected name to be %s, got %s", p2pcontext.NodeName, er.objRef.Name)
	}
	if er.objRef.UID != "test-uid" {
		t.Errorf("expected uid to be test-uid, got %s", er.objRef.UID)
	}
}

func TestNewRecorderInPod(t *testing.T) {
	fcs := fake.NewSimpleClientset([]runtime.Object{
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p2pcontext.NodeName,
				Namespace: p2pcontext.Namespace,
				UID:       "test-uid",
			},
		},
	}...)

	cs := &k8s.ClientSet{Interface: fcs, InPod: true}

	r, err := NewRecorder(context.Background(), cs)
	if err != nil {
		t.Fatal(err)
	}

	if r == nil {
		t.Fatal("expected event recorder")
	}

	er := r.(*eventRecorder)
	if er.objRef.Kind != "Pod" {
		t.Errorf("expected kind to be Pod, got %s", er.objRef.Kind)
	}
	if er.objRef.Name != p2pcontext.NodeName {
		t.Errorf("expected name to be %s, got %s", p2pcontext.NodeName, er.objRef.Name)
	}
	if er.objRef.Namespace != p2pcontext.Namespace {
		t.Errorf("expected namespace to be %s, got %s", p2pcontext.Namespace, er.objRef.Namespace)
	}
	if er.objRef.UID != "test-uid" {
		t.Errorf("expected uid to be test-uid, got %s", er.objRef.UID)
	}
}

func TestExpectedEvents(t *testing.T) {
	er := &eventRecorder{
		recorder: &testRecorder{t},
		objRef: &v1.ObjectReference{
			Kind:       "Node",
			Name:       "node-name",
			UID:        "node.UID",
			APIVersion: "node.APIVersion",
		},
	}

	er.Active()
	er.Connected()
	er.Disconnected()
	er.Initializing()
	er.Failed()
}

func TestFromContext(t *testing.T) {
	er := &eventRecorder{
		recorder: &testRecorder{t},
		objRef: &v1.ObjectReference{
			Kind:       "Node",
			Name:       "node-name",
			UID:        "node.UID",
			APIVersion: "node.APIVersion",
		},
	}

	ctx := context.WithValue(context.Background(), eventsRecorderCtxKey, er)

	er2 := FromContext(ctx)
	if er != er2 {
		t.Errorf("expected event recorders to match")
	}
}

type testRecorder struct {
	t *testing.T
}

// AnnotatedEventf implements record.EventRecorder.
func (*testRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype string, reason string, messageFmt string, args ...interface{}) {
	panic("unimplemented")
}

// Event implements record.EventRecorder.
func (*testRecorder) Event(object runtime.Object, eventtype string, reason string, message string) {
	panic("unimplemented")
}

// Eventf implements record.EventRecorder.
func (t *testRecorder) Eventf(object runtime.Object, eventtype string, reason string, messageFmt string, args ...interface{}) {
	if reason != "P2PActive" && reason != "P2PConnected" && reason != "P2PDisconnected" && reason != "P2PInitializing" && reason != "P2PFailed" {
		t.t.Errorf("unexpected reason: %s", reason)
	}
}

var _ record.EventRecorder = &testRecorder{}
