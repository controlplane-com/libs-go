package postgresql

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/common"
)

var pathSeparator = "/"

func NewContainmentQuery() *ContainmentQuery {
	return &ContainmentQuery{
		children: map[string]*ContainmentQuery{},
	}
}

type ContainmentQuery struct {
	name     string
	value    string
	children map[string]*ContainmentQuery
}

var pathRegex = regexp.MustCompilePOSIX("^[a-zA-Z0-9\\-_]*(/[a-zA-Z0-9\\-_]+)*$")

func validatePath(path string) error {
	if !pathRegex.MatchString(path) {
		return errors.New(fmt.Sprintf("Invalid path: \"%s\"\nPaths must match the regular expression: %s", path, pathRegex.String()))
	}
	return nil
}

func (q *ContainmentQuery) AddAtPath(path string, value string) error {
	if err := validatePath(path); err != nil {
		return err
	}
	pieces := strings.Split(path, pathSeparator)
	q.AddAtPathSlice(value, pieces...)
	return nil
}

func (q *ContainmentQuery) AddAtPathSlice(value string, path ...string) {
	query := q.ensureQueryAtPathSlice(path...)
	query.value = value
	query.children = map[string]*ContainmentQuery{}
}

func (q *ContainmentQuery) GetQueryAtPath(path string) (*ContainmentQuery, error) {
	if err := validatePath(path); err != nil {
		return nil, err
	}
	pieces := strings.Split(path, pathSeparator)
	currentQuery := q
	for _, p := range pieces {
		if next, ok := currentQuery.children[p]; ok {
			currentQuery = next
			continue
		}
		return nil, nil
	}
	return currentQuery, nil
}

func (q *ContainmentQuery) ToJSON() string {
	var b strings.Builder
	stack := common.NewStack[toJsonStackFrame]()
	stack.Push(newStackFrame(q))
	for stack.Length() > 0 {
		q.workOnStackFrame(&stack, &b)
	}
	return b.String()
}

func (q *ContainmentQuery) String() string {
	return q.ToJSON()
}

func (q *ContainmentQuery) workOnStackFrame(stack *common.Stack[toJsonStackFrame], b *strings.Builder) {
	currentFrame := stack.Pop()
	lenChildren := len(currentFrame.query.children)

	if currentFrame.nextChildIndex == 0 {
		b.Write([]byte("{\n"))
	}
	if currentFrame.nextChildIndex > 0 && currentFrame.nextChildIndex < lenChildren {
		b.Write([]byte(",\n"))
	}
	for i := currentFrame.nextChildIndex; i < lenChildren; i++ {
		child := currentFrame.query.children[currentFrame.childNames[i]]
		b.Write(getTabs(stack))
		b.WriteRune('"')
		b.Write([]byte(child.name))
		b.Write([]byte("\": "))
		if len(child.children) > 0 {
			currentFrame.nextChildIndex = i + 1
			stack.Push(currentFrame)
			stack.Push(newStackFrame(child))
			return
		}
		b.WriteRune('"')
		b.Write([]byte(child.value))
		b.WriteRune('"')
		if i < lenChildren-1 {
			b.WriteRune(',')
			b.WriteRune('\n')
		}
	}
	b.Write([]byte("\n"))
	b.Write(getTabsWithDelta(stack, -1))
	b.Write([]byte("}"))
}

func (q *ContainmentQuery) ensureQueryAtPathSlice(path ...string) *ContainmentQuery {
	currentQueryLevel := q
	for _, p := range path {
		if child, ok := currentQueryLevel.children[p]; ok {
			currentQueryLevel = child
			continue
		}
		//Create an empty level
		child := NewContainmentQuery()
		child.name = p
		currentQueryLevel.children[p] = child
		currentQueryLevel = child
	}
	return currentQueryLevel
}

func sliceToMap(children []*ContainmentQuery) map[string]*ContainmentQuery {
	outputMap := map[string]*ContainmentQuery{}
	for _, q := range children {
		outputMap[q.name] = q
	}
	return outputMap
}

type toJsonStackFrame struct {
	query          *ContainmentQuery
	childNames     []string
	nextChildIndex int
}

func (f toJsonStackFrame) getChildAt(i int) *ContainmentQuery {
	return f.query.children[f.childNames[i]]
}

func getTabsWithDelta[T any](stack *common.Stack[T], delta int) []byte {
	return []byte(strings.Repeat(" ", 3*(stack.Length()+1+delta)))
}

func getTabs[T any](stack *common.Stack[T]) []byte {
	return getTabsWithDelta(stack, 0)
}

func newStackFrame(query *ContainmentQuery) toJsonStackFrame {
	var names []string
	for k := range query.children {
		names = append(names, k)
	}
	return toJsonStackFrame{
		query:          query,
		childNames:     names,
		nextChildIndex: 0,
	}
}
