/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"errors"
	"testing"

	"github.com/bowei/gce-gen/pkg/cloud/meta"
)

func TestParseResourceURL(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in string
		r  *ResourceID
	}{
		{
			"https://www.googleapis.com/compute/v1/projects/some-gce-project",
			&ResourceID{"some-gce-project", "projects", nil},
		},
		{
			"https://www.googleapis.com/compute/v1/projects/some-gce-project/regions/us-central1",
			&ResourceID{"some-gce-project", "regions", meta.GlobalKey("us-central1")},
		},
		{
			"https://www.googleapis.com/compute/v1/projects/some-gce-project/zones/us-central1-b",
			&ResourceID{"some-gce-project", "zones", meta.GlobalKey("us-central1-b")},
		},
		{
			"https://www.googleapis.com/compute/v1/projects/some-gce-project/global/operations/operation-1513289952196-56054460af5a0-b1dae0c3-9bbf9dbf",
			&ResourceID{"some-gce-project", "operations", meta.GlobalKey("operation-1513289952196-56054460af5a0-b1dae0c3-9bbf9dbf")},
		},
		{
			"https://www.googleapis.com/compute/alpha/projects/some-gce-project/regions/us-central1/addresses/my-address",
			&ResourceID{"some-gce-project", "addresses", meta.RegionalKey("my-address", "us-central1")},
		},
		{
			"https://www.googleapis.com/compute/v1/projects/some-gce-project/zones/us-central1-c/instances/instance-1",
			&ResourceID{"some-gce-project", "instances", meta.ZonalKey("instance-1", "us-central1-c")},
		},
		{
			"projects/some-gce-project",
			&ResourceID{"some-gce-project", "projects", nil},
		},
		{
			"projects/some-gce-project/regions/us-central1",
			&ResourceID{"some-gce-project", "regions", meta.GlobalKey("us-central1")},
		},
		{
			"projects/some-gce-project/zones/us-central1-b",
			&ResourceID{"some-gce-project", "zones", meta.GlobalKey("us-central1-b")},
		},
		{
			"projects/some-gce-project/global/operations/operation-1513289952196-56054460af5a0-b1dae0c3-9bbf9dbf",
			&ResourceID{"some-gce-project", "operations", meta.GlobalKey("operation-1513289952196-56054460af5a0-b1dae0c3-9bbf9dbf")},
		},
		{
			"projects/some-gce-project/regions/us-central1/addresses/my-address",
			&ResourceID{"some-gce-project", "addresses", meta.RegionalKey("my-address", "us-central1")},
		},
		{
			"projects/some-gce-project/zones/us-central1-c/instances/instance-1",
			&ResourceID{"some-gce-project", "instances", meta.ZonalKey("instance-1", "us-central1-c")},
		},
	} {
		r, err := ParseResourceURL(tc.in)
		if err != nil {
			t.Errorf("ParseResourceURL(%q) = %+v, %v; want _, nil", tc.in, r, err)
			continue
		}
		if !r.Equal(tc.r) {
			t.Errorf("ParseResourceURL(%q) = %+v, nil; want %+v, nil", tc.in, r, tc.r)
		}
	}
	// Malformed URLs.
	for _, tc := range []string{
		"",
		"/",
		"/a",
		"/a/b",
		"/a/b/c",
		"/a/b/c/d",
		"/a/b/c/d/e",
		"/a/b/c/d/e/f",
		"https://www.googleapis.com/compute/v1/projects/some-gce-project/global",
		"projects/some-gce-project/global",
		"projects/some-gce-project/global/foo/bar/baz",
		"projects/some-gce-project/zones/us-central1-c/res",
		"projects/some-gce-project/zones/us-central1-c/res/name/extra",
		"https://www.googleapis.com/compute/gamma/projects/some-gce-project/global/addresses/name",
	} {
		r, err := ParseResourceURL(tc)
		if err == nil {
			t.Errorf("ParseResourceURL(%q) = %+v, %v, want _, error", tc, r, err)
		}
	}
}

type A struct {
	A, B, C string
}

type B struct {
	A, B, D string
}

type E struct{}

func (*E) MarshalJSON() ([]byte, error) {
	return nil, errors.New("injected error")
}

func TestCopyVisJSON(t *testing.T) {
	t.Parallel()

	var b B
	srcA := &A{"aa", "bb", "cc"}
	err := copyViaJSON(&b, srcA)
	if err != nil {
		t.Errorf(`copyViaJSON(&b, %+v) = %v, want nil`, srcA, err)
	} else {
		expectedB := B{"aa", "bb", ""}
		if b != expectedB {
			t.Errorf("b == %+v, want %+v", b, expectedB)
		}
	}

	var a A
	srcB := &B{"aaa", "bbb", "ccc"}
	err = copyViaJSON(&a, srcB)
	if err != nil {
		t.Errorf(`copyViaJSON(&a, %+v) = %v, want nil`, srcB, err)
	} else {
		expectedA := A{"aaa", "bbb", ""}
		if a != expectedA {
			t.Errorf("a == %+v, want %+v", a, expectedA)
		}
	}

	if err := copyViaJSON(&a, &E{}); err == nil {
		t.Errorf("copyViaJSON(&a, &E{}) = nil, want error")
	}
}
