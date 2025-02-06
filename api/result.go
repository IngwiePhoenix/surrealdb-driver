package api

import (
	"encoding/json"

	st "github.com/IngwiePhoenix/surrealdb-driver/surrealtypes"
	"github.com/wI2L/jsondiff"
)

type QueryResult struct {
	Status string      `json:"status"`
	Time   st.Duration `json:"time"`
	Result interface{} `json:"result"`
}

// QueryResponse.Result[n].[ T QueryResultTypes ]
type SimpleResult = interface{}
type ArrayResult = []SimpleResult

type RelationResult struct {
	ID     st.RecordID `json:"id"`
	In     st.RecordID `json:"in"`
	Out    st.RecordID `json:"out"`
	Values st.Object   `json:"-"`
}

func (r *RelationResult) UnmarshalJSON(data []byte) error {
	var head map[string]interface{}
	if err := json.Unmarshal(data, &head); err != nil {
		return err
	}
	var err error
	r.ID, err = st.NewRecordIDFromString(head["id"].(string))
	if err != nil {
		return err
	}
	r.In, err = st.NewRecordIDFromString(head["in"].(string))
	if err != nil {
		return err
	}
	r.Out, err = st.NewRecordIDFromString(head["out"].(string))
	if err != nil {
		return err
	}

	// clear the recycle bin
	delete(head, "id")
	delete(head, "In")
	delete(head, "out")

	// Grab rest and leave.
	r.Values = head
	return nil
}

type JsonPatchRepr = jsondiff.Operation
type JsonPatchResult = []JsonPatchRepr

type ResultTypes interface {
	SimpleResult | ArrayResult | QueryResult | JsonPatchResult | RelationResult
}
