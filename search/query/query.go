//  Copyright (c) 2014 Couchbase, Inc.
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

package query

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/util"
	index "github.com/blevesearch/bleve_index_api"
)

var logger = log.New(io.Discard, "bleve mapping ", log.LstdFlags)

// SetLog sets the logger used for logging
// by default log messages are sent to io.Discard
func SetLog(l *log.Logger) {
	logger = l
}

// A Query represents a description of the type
// and parameters for a query into the index.
type Query interface {
	Searcher(ctx context.Context, i index.IndexReader, m mapping.IndexMapping,
		options search.SearcherOptions) (search.Searcher, error)
}

// A BoostableQuery represents a Query which can be boosted
// relative to other queries.
type BoostableQuery interface {
	Query
	SetBoost(b float64)
	Boost() float64
}

// A FieldableQuery represents a Query which can be restricted
// to a single field.
type FieldableQuery interface {
	Query
	SetField(f string)
	Field() string
}

// A ValidatableQuery represents a Query which can be validated
// prior to execution.
type ValidatableQuery interface {
	Query
	Validate() error
}

// ParseQuery deserializes a JSON representation of
// a PreSearchData object.
func ParsePreSearchData(input []byte) (map[string]interface{}, error) {
	var rv map[string]interface{}

	var tmp map[string]json.RawMessage
	err := util.UnmarshalJSON(input, &tmp)
	if err != nil {
		return nil, err
	}

	for k, v := range tmp {
		switch k {
		case search.KnnPreSearchDataKey:
			var value []*search.DocumentMatch
			if v != nil {
				err := util.UnmarshalJSON(v, &value)
				if err != nil {
					return nil, err
				}
			}
			if rv == nil {
				rv = make(map[string]interface{})
			}
			rv[search.KnnPreSearchDataKey] = value
		}
	}
	return rv, nil
}

// ParseQuery deserializes a JSON representation of
// a Query object.
func ParseQuery(input []byte) (Query, error) {
	var tmp map[string]interface{}
	err := util.UnmarshalJSON(input, &tmp)
	if err != nil {
		return nil, err
	}
	_, hasFuzziness := tmp["fuzziness"]
	_, isMatchQuery := tmp["match"]
	_, isMatchPhraseQuery := tmp["match_phrase"]
	_, hasTerms := tmp["terms"]
	if hasFuzziness && !isMatchQuery && !isMatchPhraseQuery && !hasTerms {
		var rv FuzzyQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	if isMatchQuery {
		var rv MatchQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	if isMatchPhraseQuery {
		var rv MatchPhraseQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	if hasTerms {
		var rv PhraseQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			// now try multi-phrase
			var rv2 MultiPhraseQuery
			err = util.UnmarshalJSON(input, &rv2)
			if err != nil {
				return nil, err
			}
			return &rv2, nil
		}
		return &rv, nil
	}
	_, isTermQuery := tmp["term"]
	if isTermQuery {
		var rv TermQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasMust := tmp["must"]
	_, hasShould := tmp["should"]
	_, hasMustNot := tmp["must_not"]
	if hasMust || hasShould || hasMustNot {
		var rv BooleanQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasConjuncts := tmp["conjuncts"]
	if hasConjuncts {
		var rv ConjunctionQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasDisjuncts := tmp["disjuncts"]
	if hasDisjuncts {
		var rv DisjunctionQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}

	_, hasSyntaxQuery := tmp["query"]
	if hasSyntaxQuery {
		var rv QueryStringQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasMin := tmp["min"].(float64)
	_, hasMax := tmp["max"].(float64)
	if hasMin || hasMax {
		var rv NumericRangeQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasMinStr := tmp["min"].(string)
	_, hasMaxStr := tmp["max"].(string)
	if hasMinStr || hasMaxStr {
		var rv TermRangeQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasStart := tmp["start"]
	_, hasEnd := tmp["end"]
	if hasStart || hasEnd {
		var rv DateRangeStringQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasPrefix := tmp["prefix"]
	if hasPrefix {
		var rv PrefixQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasRegexp := tmp["regexp"]
	if hasRegexp {
		var rv RegexpQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasWildcard := tmp["wildcard"]
	if hasWildcard {
		var rv WildcardQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasMatchAll := tmp["match_all"]
	if hasMatchAll {
		var rv MatchAllQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasMatchNone := tmp["match_none"]
	if hasMatchNone {
		var rv MatchNoneQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasDocIds := tmp["ids"]
	if hasDocIds {
		var rv DocIDQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasBool := tmp["bool"]
	if hasBool {
		var rv BoolFieldQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasTopLeft := tmp["top_left"]
	_, hasBottomRight := tmp["bottom_right"]
	if hasTopLeft && hasBottomRight {
		var rv GeoBoundingBoxQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasDistance := tmp["distance"]
	if hasDistance {
		var rv GeoDistanceQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}
	_, hasPoints := tmp["polygon_points"]
	if hasPoints {
		var rv GeoBoundingPolygonQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}

	_, hasGeo := tmp["geometry"]
	if hasGeo {
		var rv GeoShapeQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}

	_, hasCIDR := tmp["cidr"]
	if hasCIDR {
		var rv IPRangeQuery
		err := util.UnmarshalJSON(input, &rv)
		if err != nil {
			return nil, err
		}
		return &rv, nil
	}

	return nil, fmt.Errorf("unknown query type")
}

// expandQuery traverses the input query tree and returns a new tree where
// query string queries have been expanded into base queries. Returned tree may
// reference queries from the input tree or new queries.
func expandQuery(m mapping.IndexMapping, query Query) (Query, error) {
	var expand func(query Query) (Query, error)
	var expandSlice func(queries []Query) ([]Query, error)

	expandSlice = func(queries []Query) ([]Query, error) {
		expanded := []Query{}
		for _, q := range queries {
			exp, err := expand(q)
			if err != nil {
				return nil, err
			}
			expanded = append(expanded, exp)
		}
		return expanded, nil
	}

	expand = func(query Query) (Query, error) {
		switch q := query.(type) {
		case *QueryStringQuery:
			parsed, err := parseQuerySyntax(q.Query)
			if err != nil {
				return nil, fmt.Errorf("could not parse '%s': %s", q.Query, err)
			}
			return expand(parsed)
		case *ConjunctionQuery:
			children, err := expandSlice(q.Conjuncts)
			if err != nil {
				return nil, err
			}
			q.Conjuncts = children
			return q, nil
		case *DisjunctionQuery:
			children, err := expandSlice(q.Disjuncts)
			if err != nil {
				return nil, err
			}
			q.Disjuncts = children
			return q, nil
		case *BooleanQuery:
			var err error
			q.Must, err = expand(q.Must)
			if err != nil {
				return nil, err
			}
			q.Should, err = expand(q.Should)
			if err != nil {
				return nil, err
			}
			q.MustNot, err = expand(q.MustNot)
			if err != nil {
				return nil, err
			}
			return q, nil
		default:
			return query, nil
		}
	}
	return expand(query)
}

// DumpQuery returns a string representation of the query tree, where query
// string queries have been expanded into base queries. The output format is
// meant for debugging purpose and may change in the future.
func DumpQuery(m mapping.IndexMapping, query Query) (string, error) {
	q, err := expandQuery(m, query)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(q, "", "  ")
	return string(data), err
}
