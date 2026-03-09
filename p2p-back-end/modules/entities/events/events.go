package events

type Event interface {
	String() string
}

type MessageUserEvent struct {
	EventName string      `json:"event_name"`
	Data      interface{} `json:"data"`
}

func (e *MessageUserEvent) String() string {
	return e.EventName
}

type MessageCompaniesEvent struct {
	Companies []CompanyEvent `json:"companies"`
}

func (e *MessageCompaniesEvent) String() string {
	return "autocorp.company.change"
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

func (e *MessageDepartmentEvent) String() string {
	return "autocorp.department.change"
}

type DepartmentEvent struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type MessageSectionEvent struct {
	Sections []SectionEvent `json:"sections"`
}

func (e *MessageSectionEvent) String() string {
	return "autocorp.section.change"
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

func (e *MessagePositionEvent) String() string {
	return "autocorp.position.change"
}

type PositionEvent struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// --- Sync Request Events (Begin) ---

type RequestCompanySyncEvent struct{}

func (e *RequestCompanySyncEvent) String() string { return "autocorp.company.begin" }

type RequestDepartmentSyncEvent struct{}

func (e *RequestDepartmentSyncEvent) String() string { return "autocorp.department.begin" }

type RequestUserSyncEvent struct{}

func (e *RequestUserSyncEvent) String() string { return "autocorp.user.begin" }
