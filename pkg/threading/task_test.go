package threading

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
	"time"
)

type TaskTestSuite struct {
	suite.Suite
}

func TestTasks(t *testing.T) {
	suite.Run(t, &TaskTestSuite{})
}

func (s *TaskTestSuite) TestWorkerPanics() {
	task := NewTask[string](func() (string, error) { panic("an error") })
	strings, err := RunTasks(task)
	s.Len(strings, 1)
	s.NotNil(err)
}

func (s *TaskTestSuite) TestWorkerExecutesSuccessful() {
	task := NewTask[string](func() (string, error) { return "hello", nil })
	strings, err := RunTasks(task)
	s.Len(strings, 1)
	s.Equal("hello", strings[0])
	s.Nil(err)
}

func (s *TaskTestSuite) TestManyWorkers() {
	r := rand.New(rand.NewSource(0))
	waitARandomNumberOfSeconds := func() (string, error) {
		toWait := time.Millisecond * time.Duration(r.Int()%5000)
		time.Sleep(toWait)
		return fmt.Sprintf("Waited %f seconds", toWait.Seconds()), nil
	}
	var tasks []*Task[string]
	for i := 0; i < 100; i++ {
		tasks = append(tasks, NewTask[string](waitARandomNumberOfSeconds))
	}
	strings, err := RunTasks(tasks...)
	for _, str := range strings {
		fmt.Println(str)
	}
	s.Len(strings, 100)
	s.Nil(err)
}
