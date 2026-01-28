package maps

import (
	"errors"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestMergeSuite struct {
	suite.Suite
}

func TestMerge(t *testing.T) {
	suite.Run(t, &TestMergeSuite{})
}

func (s *TestMergeSuite) TestNilMaps() {
	var a map[string]string
	var b map[string]string
	m, err := Merge(a, b, nil)
	s.Nil(m)
	s.Nil(err)
}

func (s *TestMergeSuite) TestNilA() {
	var a map[string]string
	b := map[string]string{}
	m, err := Merge(a, b, nil)
	s.NotNil(m)
	s.Empty(m)
	s.Nil(err)
}

func (s *TestMergeSuite) TestNilB() {
	a := map[string]string{}
	var b map[string]string
	m, err := Merge(a, b, nil)
	s.NotNil(m)
	s.Empty(m)
	s.Nil(err)
}

func (s *TestMergeSuite) TestNoMergeFunc() {
	a := map[string]string{
		"1": "a1",
		"2": "a2",
	}
	b := map[string]string{
		"2": "b2",
		"3": "b3",
	}
	m, err := Merge(a, b, nil)
	s.NotNil(m)
	s.Equal("a1", m["1"])
	s.Equal("b2", m["2"])
	s.Equal("b3", m["3"])
	s.Nil(err)
}

func (s *TestMergeSuite) TestMergeFunc() {
	a := map[string]string{
		"1": "a1",
		"2": "a2",
	}
	b := map[string]string{
		"2": "b2",
		"3": "b3",
	}
	m, err := Merge(a, b, func(sa string, sb string) (string, error) {
		return sa + sb, nil
	})
	s.NotNil(m)
	s.Equal("a1", m["1"])
	s.Equal("a2b2", m["2"])
	s.Equal("b3", m["3"])
	s.Nil(err)
}

func (s *TestMergeSuite) TestMergeFuncErr() {
	a := map[string]string{
		"1": "a1",
		"2": "a2",
	}
	b := map[string]string{
		"2": "b2",
		"3": "b3",
	}
	m, err := Merge(a, b, func(sa string, sb string) (string, error) {
		return "", errors.New("test_merge_func_error")
	})
	s.Nil(m)
	s.NotNil(err)
	s.Equal("test_merge_func_error", err.Error())
}
