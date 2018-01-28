/*
Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vindexes

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/youtube/vitess/go/sqltypes"
	"github.com/youtube/vitess/go/vt/proto/topodata"
)

var (
	_ Unique    = (*LookupUnique)(nil)
	_ Lookup    = (*LookupUnique)(nil)
	_ NonUnique = (*LookupNonUnique)(nil)
	_ Lookup    = (*LookupNonUnique)(nil)
)

func init() {
	Register("lookup", NewLookup)
	Register("lookup_unique", NewLookupUnique)
}

// LookupNonUnique defines a vindex that uses a lookup table and create a mapping between from ids and KeyspaceId.
// It's NonUnique and a Lookup.
type LookupNonUnique struct {
	name            string
	scatterIfAbsent bool
	lkp             lookupInternal
}

// String returns the name of the vindex.
func (ln *LookupNonUnique) String() string {
	return ln.name
}

// Cost returns the cost of this vindex as 20.
func (ln *LookupNonUnique) Cost() int {
	return 20
}

// Map returns the corresponding KeyspaceId values for the given ids.
func (ln *LookupNonUnique) Map(vcursor VCursor, ids []sqltypes.Value) ([]Ksids, error) {
	out := make([]Ksids, 0, len(ids))
	results, err := ln.lkp.Lookup(vcursor, ids)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		if len(result.Rows) == 0 {
			if ln.scatterIfAbsent {
				out = append(out, Ksids{Range: &topodata.KeyRange{}})
				continue
			}
			out = append(out, Ksids{})
			continue
		}
		ksids := make([][]byte, 0, len(result.Rows))
		for _, row := range result.Rows {
			ksids = append(ksids, row[0].ToBytes())
		}
		out = append(out, Ksids{IDs: ksids})
	}
	return out, nil
}

// Verify returns true if ids maps to ksids.
func (ln *LookupNonUnique) Verify(vcursor VCursor, ids []sqltypes.Value, ksids [][]byte) ([]bool, error) {
	if ln.scatterIfAbsent {
		out := make([]bool, len(ids))
		for i := range ids {
			out[i] = true
		}
		return out, nil
	}
	return ln.lkp.Verify(vcursor, ids, ksidsToValues(ksids))
}

// Create reserves the id by inserting it into the vindex table.
func (ln *LookupNonUnique) Create(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksids [][]byte, ignoreMode bool) error {
	return ln.lkp.Create(vcursor, rowsColValues, ksidsToValues(ksids), ignoreMode)
}

// Delete deletes the entry from the vindex table.
func (ln *LookupNonUnique) Delete(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksid []byte) error {
	return ln.lkp.Delete(vcursor, rowsColValues, sqltypes.MakeTrusted(sqltypes.VarBinary, ksid))
}

// Update updates the entry in the vindex table.
func (ln *LookupNonUnique) Update(vcursor VCursor, oldValues []sqltypes.Value, ksid []byte, newValues []sqltypes.Value) error {
	return ln.lkp.Update(vcursor, oldValues, sqltypes.MakeTrusted(sqltypes.VarBinary, ksid), newValues)
}

// MarshalJSON returns a JSON representation of LookupHash.
func (ln *LookupNonUnique) MarshalJSON() ([]byte, error) {
	return json.Marshal(ln.lkp)
}

// NewLookup creates a LookupNonUnique vindex.
func NewLookup(name string, m map[string]string) (Vindex, error) {
	lookup := &LookupNonUnique{name: name}
	lookup.lkp.Init(m)
	var err error
	lookup.scatterIfAbsent, err = getScatterIfAbsent(m)
	if err != nil {
		return nil, err
	}
	return lookup, nil
}

func ksidsToValues(ksids [][]byte) []sqltypes.Value {
	values := make([]sqltypes.Value, 0, len(ksids))
	for _, ksid := range ksids {
		values = append(values, sqltypes.MakeTrusted(sqltypes.VarBinary, ksid))
	}
	return values
}

func getScatterIfAbsent(m map[string]string) (bool, error) {
	val, ok := m["scatter_if_absent"]
	if !ok {
		return false, nil
	}
	switch val {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("scatter_if_absent value must be 'true' or 'false': '%s'", val)
	}
}

//====================================================================

// LookupUnique defines a vindex that uses a lookup table.
// The table is expected to define the id column as unique. It's
// Unique and a Lookup.
type LookupUnique struct {
	name string
	lkp  lookupInternal
}

// NewLookupUnique creates a LookupUnique vindex.
func NewLookupUnique(name string, m map[string]string) (Vindex, error) {
	lu := &LookupUnique{name: name}
	lu.lkp.Init(m)
	scatter, err := getScatterIfAbsent(m)
	if err != nil {
		return nil, err
	}
	if scatter {
		return nil, errors.New("scatter_if_absent cannot be true for a unique lookup vindex")
	}
	return lu, nil
}

// String returns the name of the vindex.
func (lu *LookupUnique) String() string {
	return lu.name
}

// Cost returns the cost of this vindex as 10.
func (lu *LookupUnique) Cost() int {
	return 10
}

// Map returns the corresponding KeyspaceId values for the given ids.
func (lu *LookupUnique) Map(vcursor VCursor, ids []sqltypes.Value) ([][]byte, error) {
	out := make([][]byte, 0, len(ids))
	results, err := lu.lkp.Lookup(vcursor, ids)
	if err != nil {
		return nil, err
	}
	for i, result := range results {
		switch len(result.Rows) {
		case 0:
			out = append(out, nil)
		case 1:
			out = append(out, result.Rows[0][0].ToBytes())
		default:
			return nil, fmt.Errorf("Lookup.Map: unexpected multiple results from vindex %s: %v", lu.lkp.Table, ids[i])
		}
	}
	return out, nil
}

// Verify returns true if ids maps to ksids.
func (lu *LookupUnique) Verify(vcursor VCursor, ids []sqltypes.Value, ksids [][]byte) ([]bool, error) {
	return lu.lkp.Verify(vcursor, ids, ksidsToValues(ksids))
}

// Create reserves the id by inserting it into the vindex table.
func (lu *LookupUnique) Create(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksids [][]byte, ignoreMode bool) error {
	return lu.lkp.Create(vcursor, rowsColValues, ksidsToValues(ksids), ignoreMode)
}

// Update updates the entry in the vindex table.
func (lu *LookupUnique) Update(vcursor VCursor, oldValues []sqltypes.Value, ksid []byte, newValues []sqltypes.Value) error {
	return lu.lkp.Update(vcursor, oldValues, sqltypes.MakeTrusted(sqltypes.VarBinary, ksid), newValues)
}

// Delete deletes the entry from the vindex table.
func (lu *LookupUnique) Delete(vcursor VCursor, rowsColValues [][]sqltypes.Value, ksid []byte) error {
	return lu.lkp.Delete(vcursor, rowsColValues, sqltypes.MakeTrusted(sqltypes.VarBinary, ksid))
}

// MarshalJSON returns a JSON representation of LookupUnique.
func (lu *LookupUnique) MarshalJSON() ([]byte, error) {
	return json.Marshal(lu.lkp)
}
