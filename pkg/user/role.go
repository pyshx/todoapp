package user

type Role string

const (
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

func (r Role) IsValid() bool { return r == RoleEditor || r == RoleViewer }
func (r Role) CanEdit() bool { return r == RoleEditor }
func (r Role) String() string { return string(r) }

func ParseRole(s string) (Role, bool) {
	r := Role(s)
	if !r.IsValid() {
		return "", false
	}
	return r, true
}
