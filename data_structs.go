package jsonapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var objectSuffix = []byte("{")
var arraySuffix = []byte("[")

// A Document represents a JSON API document as specified here: http://jsonapi.org.
type Document struct {
	Links    *Links                 `json:"links,omitempty"`
	Data     *DataContainer         `json:"data"`
	Included []Data                 `json:"included,omitempty"`
	Meta     map[string]interface{} `json:"meta,omitempty"`
}

// A DataContainer is used to marshal and unmarshal single objects and arrays
// of objects.
type DataContainer struct {
	DataObject *Data
	DataArray  []Data
}

// UnmarshalJSON unmarshals the JSON-encoded data to the DataObject field if the
// root element is an object or to the DataArray field for arrays.
func (c *DataContainer) UnmarshalJSON(payload []byte) error {
	if bytes.HasPrefix(payload, objectSuffix) {
		return json.Unmarshal(payload, &c.DataObject)
	}

	if bytes.HasPrefix(payload, arraySuffix) {
		return json.Unmarshal(payload, &c.DataArray)
	}

	return errors.New("expected a JSON encoded object or array")
}

// MarshalJSON returns the JSON encoding of the DataArray field or the DataObject
// field. It will return "null" if neither of them is set.
func (c *DataContainer) MarshalJSON() ([]byte, error) {
	if c.DataArray != nil {
		return json.Marshal(c.DataArray)
	}

	return json.Marshal(c.DataObject)
}

// Links is a general struct for document links and relationship links.
type Links struct {
	Self     string `json:"self,omitempty"`
	Related  string `json:"related,omitempty"`
	First    string `json:"first,omitempty"`
	Previous string `json:"prev,omitempty"`
	Next     string `json:"next,omitempty"`
	Last     string `json:"last,omitempty"`
}

// Data is a general struct for document data and included data.
type Data struct {
	Type          string                  `json:"type"`
	ID            string                  `json:"id"`
	Attributes    json.RawMessage         `json:"attributes"`
	Relationships map[string]Relationship `json:"relationships,omitempty"`
	Links         *Links                  `json:"links,omitempty"`
}

// Relationship contains reference IDs to the related structs
type Relationship struct {
	Links *Links                     `json:"links,omitempty"`
	Data  *RelationshipDataContainer `json:"data,omitempty"`
	Meta  map[string]interface{}     `json:"meta,omitempty"`
}

// A RelationshipDataContainer is used to marshal and unmarshal single relationship
// objects and arrays of relationship objects.
type RelationshipDataContainer struct {
	DataObject *RelationshipData
	DataArray  []RelationshipData
}

// UnmarshalJSON unmarshals the JSON-encoded data to the DataObject field if the
// root element is an object or to the DataArray field for arrays.
func (c *RelationshipDataContainer) UnmarshalJSON(payload []byte) error {
	if bytes.HasPrefix(payload, objectSuffix) {
		// payload is an object
		return json.Unmarshal(payload, &c.DataObject)
	}

	if bytes.HasPrefix(payload, arraySuffix) {
		// payload is an array
		return json.Unmarshal(payload, &c.DataArray)
	}

	return errors.New("Invalid json for relationship data array/object")
}

// MarshalJSON returns the JSON encoding of the DataArray field or the DataObject
// field. It will return "null" if neither of them is set.
func (c *RelationshipDataContainer) MarshalJSON() ([]byte, error) {
	if c.DataArray != nil {
		return json.Marshal(c.DataArray)
	}
	return json.Marshal(c.DataObject)
}

// RelationshipData represents one specific reference ID.
type RelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CustomObject struct {
	Fields []string
	Object interface{}
}

type FilterFields map[string][]string

func (f FilterFields) ParseQuery(q url.Values) {
	rpm := regexp.MustCompile(`(?i)^fields\[([^\]]+)]$`)

	for k, v := range q {
		matches := rpm.FindStringSubmatch(k)
		if len(matches) > 0 {
			f[matches[1]] = strings.Split(strings.Join(v, ","), ",")
		}
	}
}

type ObjectAttributes map[string]interface{}

func (co CustomObject) JSONToStruct() map[string]string {
	rpm := regexp.MustCompile(`(?i)^([^,]+)(,|$)`)
	res := map[string]string{}
	ref := getType(co.Object)

	for i := 0; i < ref.NumField(); i++ {
		f := ref.Field(i)
		tag, ok := f.Tag.Lookup("json")
		if ok {
			matches := rpm.FindStringSubmatch(tag)
			if len(matches) > 0 && matches[1] != "-" {
				res[matches[1]] = f.Name
			}
		}
	}
	return res
}

func (co CustomObject) MarshalJSON() ([]byte, error) {
	obj := ObjectAttributes{}
	dict := co.JSONToStruct()
	ref := getValue(co.Object)

	for _, f := range co.Fields {
		if dict[f] != "" {
			obj[f] = ref.FieldByName(dict[f]).Interface()
		}
	}

	b, err := json.Marshal(&obj)

	return b, err
}
