package task

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

func (s Status) IsValid() bool  { return s == StatusTodo || s == StatusInProgress || s == StatusDone }
func (s Status) String() string { return string(s) }

func ParseStatus(str string) (Status, bool) {
	s := Status(str)
	if !s.IsValid() {
		return "", false
	}
	return s, true
}
