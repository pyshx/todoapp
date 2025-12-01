package task

type Visibility string

const (
	VisibilityOnlyMe      Visibility = "only_me"
	VisibilityCompanyWide Visibility = "company_wide"
)

func (v Visibility) IsValid() bool  { return v == VisibilityOnlyMe || v == VisibilityCompanyWide }
func (v Visibility) String() string { return string(v) }

func ParseVisibility(s string) (Visibility, bool) {
	v := Visibility(s)
	if !v.IsValid() {
		return "", false
	}
	return v, true
}
