package replication

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Mapping struct {
	Maps []Map `json:"maps"`
}

type Map struct {
	Name     string            `json:"name"`
	Table    string            `json:"table"`
	FieldMap map[string]string `json:"fieldMap"`
	Fields   []string          `json:"fields"`
	Actions  []Action          `json:"actions"`
}

func (m *Map) Match(c *Change) bool {
	if m == nil || c == nil {
		return false
	}
	if m.Table != fmt.Sprintf("%s.%s", c.Schema, c.Table) {
		return false
	}
	if len(m.Actions) == 0 {
		return true
	}
	for _, a := range m.Actions {
		if c.Action == a {
			return true
		}
	}
	return false
}

func (m *Map) Message(c *Change) *Message {
	switch c.Action {
	case ActionDelete, ActionUpdate:
		return m.outboxMessage(c.OldKeys.KeyNames, c.OldKeys.KeyValues)
	default:
		return m.outboxMessage(c.ColumnNames, c.ColumnValues)
	}
}

func (m *Map) outboxMessage(keyNames []string, keyValues []any) *Message {
	row := m.pivot(keyNames, keyValues)
	payloadMap := map[string]any{}
	_ = json.Unmarshal([]byte(assertParam[string](row, "payload")), &payloadMap)
	message := &Message{
		Id:        assertParam[string](row, "id"),
		Created:   assertParam[string](row, "created"),
		Delivered: time.Now().UTC().Format(time.RFC3339),
		Payload:   payloadMap,
	}
	return message
}

func (m *Map) pivot(keyNames []string, keyValues []any) map[string]any {
	row := map[string]any{}
	for i, name := range keyNames {
		mappedName := m.fieldName(name)
		if mappedName == "" {
			//This means the name was excluded
			continue
		}
		row[mappedName] = keyValues[i]
	}
	return row
}

func (m *Map) fieldName(name string) string {
	if m.FieldMap == nil {
		return name
	}
	for k, v := range m.FieldMap {
		if strings.EqualFold(k, name) {
			if v == "" {
				return name
			}
			return v
		}
	}
	return ""
}
