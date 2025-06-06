//  Copyright (c) 2018 Couchbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pl

import (
	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/analysis/token/stop"
	"github.com/blevesearch/bleve/v2/registry"
)

func StopTokenFilterConstructor(config map[string]interface{}, cache *registry.Cache) (analysis.TokenFilter, error) {
	tokenMap, err := cache.TokenMapNamed(StopName)
	if err != nil {
		return nil, err
	}
	return stop.NewStopTokensFilter(tokenMap), nil
}

func init() {
	err := registry.RegisterTokenFilter(StopName, StopTokenFilterConstructor)
	if err != nil {
		panic(err)
	}
}
