package events

type Event interface {
	String() string
}

type MessageUserEvent struct {
	Users []UserEvent `json:"users"`
}

type UserEvent struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	NameTh       string `json:"name_th"`
	NameEn       string `json:"name_en"`
	CompanyID    uint   `json:"company_id"`
	DepartmentID uint   `json:"department_id"`
	SectionID    uint   `json:"section_id"`
	PositionID   uint   `json:"position_id"`
	Deleted      bool   `json:"deleted"`
}

type MessageCompaniesEvent struct {
	Companies []CompanyEvent `json:"companies"`
}

type CompanyEvent struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	BranchName   string `json:"branch_name"`
	BranchNameEn string `json:"branch_name_en"`
	BranchNo     string `json:"branch_no"`
	Address      string `json:"address"`
	TaxID        string `json:"taxid"`
	Province     string `json:"province"`
}

type MessageDepartmentEvent struct {
	Departments []DepartmentEvent `json:"departments"`
}

type DepartmentEvent struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type MessageSectionEvent struct {
	Sections []SectionEvent `json:"sections"`
}

type SectionEvent struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	DepartmentID uint   `json:"department_id"`
}

type MessagePositionEvent struct {
	Positions []PositionEvent `json:"positions"`
}

type PositionEvent struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func (MessageUserEvent) String() string {
	return "autocorp.user.change"
}

func (MessageCompaniesEvent) String() string {
	return "autocorp.company.change"
}

func (MessageDepartmentEvent) String() string {
	return "autocorp.department.change"
}

func (MessageSectionEvent) String() string {
	return "autocorp.section.change"
}

func (MessagePositionEvent) String() string {
	return "autocorp.position.change"
}

type MessageCompaniesBeginEvent struct{}

func (MessageCompaniesBeginEvent) String() string {
	return "autocorp.company.begin"
}

type MessageDepartmentBeginEvent struct{}

func (MessageDepartmentBeginEvent) String() string {
	return "autocorp.department.begin"
}

type MessageUserBeginEvent struct{}

func (MessageUserBeginEvent) String() string {
	return "autocorp.user.begin"
}

type MessageSectionBeginEvent struct{}

func (MessageSectionBeginEvent) String() string {
	return "autocorp.section.begin"
}

type MessagePositionBeginEvent struct{}

func (MessagePositionBeginEvent) String() string {
	return "autocorp.position.begin"
}
