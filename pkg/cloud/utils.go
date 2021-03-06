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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bowei/gce-gen/pkg/cloud/meta"
)

const (
	gaPrefix    = "https://www.googleapis.com/compute/v1/"
	alphaPrefix = "https://www.googleapis.com/compute/alpha/"
	betaPrefix  = "https://www.googleapis.com/compute/beta/"
)

var (
	allPrefixes = []string{gaPrefix, alphaPrefix, betaPrefix}
)

// ResourceID identifies a GCE resource as parsed from compute resource URL.
type ResourceID struct {
	ProjectID string
	Resource  string
	Key       *meta.Key
}

// Equal returns true if two resource IDs are equal.
func (r *ResourceID) Equal(other *ResourceID) bool {
	if r.ProjectID != other.ProjectID || r.Resource != other.Resource {
		return false
	}
	if r.Key != nil && other.Key != nil {
		return *r.Key == *other.Key
	}
	if r.Key == nil && other.Key == nil {
		return true
	}
	return false
}

// ParseResourceURL parses resource URLs of the following formats:
//
//   projects/<proj>/global/<res>/<name>
//   projects/<proj>/regions/<region>/<res>/<name>
//   projects/<proj>/zones/<zone>/<res>/<name>
//   [https://www.googleapis.com/compute/<ver>]/projects/<proj>/global/<res>/<name>
//   [https://www.googleapis.com/compute/<ver>]/projects/<proj>/regions/<region>/<res>/<name>
//   [https://www.googleapis.com/compute/<ver>]/projects/<proj>/zones/<zone>/<res>/<name>
func ParseResourceURL(url string) (*ResourceID, error) {
	errNotValid := fmt.Errorf("%q is not a valid resource URL", url)

	// Remove the "https://..." prefix if present
	for _, prefix := range allPrefixes {
		if strings.HasPrefix(url, prefix) {
			if len(url) < len(prefix) {
				return nil, errNotValid
			}
			url = url[len(prefix):]
			break
		}
	}

	parts := strings.Split(url, "/")
	if len(parts) < 2 || parts[0] != "projects" {
		return nil, errNotValid
	}

	ret := &ResourceID{ProjectID: parts[1]}
	if len(parts) == 2 {
		ret.Resource = "projects"
		return ret, nil
	}

	if len(parts) < 4 {
		return nil, errNotValid
	}

	if len(parts) == 4 {
		switch parts[2] {
		case "regions":
			ret.Resource = "regions"
			ret.Key = meta.GlobalKey(parts[3])
			return ret, nil
		case "zones":
			ret.Resource = "zones"
			ret.Key = meta.GlobalKey(parts[3])
			return ret, nil
		default:
			return nil, errNotValid
		}
	}

	switch parts[2] {
	case "global":
		if len(parts) != 5 {
			return nil, errNotValid
		}
		ret.Resource = parts[3]
		ret.Key = meta.GlobalKey(parts[4])
		return ret, nil
	case "regions":
		if len(parts) != 6 {
			return nil, errNotValid
		}
		ret.Resource = parts[4]
		ret.Key = meta.RegionalKey(parts[5], parts[3])
		return ret, nil
	case "zones":
		if len(parts) != 6 {
			return nil, errNotValid
		}
		ret.Resource = parts[4]
		ret.Key = meta.ZonalKey(parts[5], parts[3])
		return ret, nil
	}
	return nil, errNotValid
}

func copyViaJSON(dest, src interface{}) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dest)
}
