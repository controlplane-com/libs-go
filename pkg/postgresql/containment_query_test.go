package postgresql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ContainmentQuerySuite struct {
	suite.Suite
}

func TestContainmentQuerySuite(t *testing.T) {
	suite.Run(t, new(ContainmentQuerySuite))
}

func (s *ContainmentQuerySuite) TestNewContainmentQuery() {
	q := NewContainmentQuery()
	require.NotNil(s.T(), q)
}

func (s *ContainmentQuerySuite) TestContainmentQuery_AddValueAtPath() {
	q := NewContainmentQuery()
	err := q.AddAtPath("one", "some_other_value")
	require.NoError(s.T(), err)

	got, err := q.GetQueryAtPath("one")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	assert.Equal(s.T(), "some_other_value", got.value)

	err = q.AddAtPath("one/two/three", "value at level 3")
	require.NoError(s.T(), err)

	got, err = q.GetQueryAtPath("one/two/three")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	assert.Equal(s.T(), "value at level 3", got.value)

	err = q.AddAtPath("one/two/three_sibling", "sibling")
	require.NoError(s.T(), err)

	got, err = q.GetQueryAtPath("one/two/three")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	assert.Equal(s.T(), "value at level 3", got.value)

	got, err = q.GetQueryAtPath("one/two/three_sibling")
	require.NoError(s.T(), err)
	require.NotNil(s.T(), got)
	assert.Equal(s.T(), "sibling", got.value)
}

func (s *ContainmentQuerySuite) TestContainmentQuery_ToJSON() {
	q := NewContainmentQuery()
	require.NoError(s.T(), q.AddAtPath("event/context/name", "my_name"))
	require.NoError(s.T(), q.AddAtPath("event/resource/id", "1234"))

	json := q.ToJSON()
	expected := []string{
		`{
   "event": {
      "context": {
         "name": "my_name"
      },
      "resource": {
         "id": "1234"
      }
   }
}`,
		`{
   "event": {
      "resource": {
         "id": "1234"
      },
      "context": {
         "name": "my_name"
      }
   }
}`,
	}

	assert.Contains(s.T(), expected, json)
}

func (s *ContainmentQuerySuite) TestContainmentQuery_ToQueryString() {
	q := NewContainmentQuery()
	require.NoError(s.T(), q.AddAtPath("context/name", "my_name"))
	require.NoError(s.T(), q.AddAtPath("resource/id", "1234"))

	json := q.String()
	expectedA := `{
   "context": {
      "name": "my_name"
   },
   "resource": {
      "id": "1234"
   }
}`
	expectedB := `{
   "resource": {
      "id": "1234"
   },
   "context": {
      "name": "my_name"
   }
}`

	if !assert.True(s.T(), json == expectedA || json == expectedB, "unexpected json:\n%s", json) {
		fmt.Println(json)
	}
}
