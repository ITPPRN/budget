package utils

type MappingRow struct {
	Master string
	ACG    []string
	HMW    []string
	CLIK   []string
}

// DepartmentMappingData is the source of truth for Master Department aggregation.
// It maps raw granular codes from different entities to a consolidated Master Code.
var DepartmentMappingData = []MappingRow{
	{Master: "ACC",
		ACG:  []string{"ACC", "ACC-AP", "ACC-AR", "ACC-CENTER", "ACC-FA", "ACC-GL", "ACC-MNG", "ACC_AP", "ACC_AR", "ACC_CENTER", "ACC_FA", "ACC_GL", "ACC_MNG"},
		HMW:  []string{"ACC", "ACC-AP", "ACC-AR", "ACC-CENTER", "ACC-FA", "ACC-GL", "ACC-MNG", "ACC_AP", "ACC_AR", "ACC_CENTER", "ACC_FA", "ACC_GL", "ACC_MNG"},
		CLIK: []string{"ACC", "ACC-AP", "ACC-AR", "ACC-CENTER", "ACC-FA", "ACC-GL", "ACC-MNG", "ACC_AP", "ACC_AR", "ACC_CENTER", "ACC_FA", "ACC_GL", "ACC_MNG"}},
	{Master: "BUDGET",
		ACG:  []string{"BUDGET"},
		HMW:  []string{"BUDGET"},
		CLIK: []string{"BUDGET"}},
	{Master: "DPO",
		ACG:  []string{"DM", "DPO"},
		HMW:  []string{"DM", "DPO"},
		CLIK: []string{"DM", "DPO"}},
	{Master: "FIN",
		ACG:  []string{"FIN", "FIN-CENTER", "FIN_CENTER"},
		HMW:  []string{"FIN", "FIN-CENTER", "FIN_CENTER"},
		CLIK: []string{"FIN", "FIN-CENTER", "FIN_CENTER"}},
	{Master: "G-CENTER",
		ACG:  []string{"G-ADMIN", "G-CENTER", "G_CENTER", "G_ADMIN"},
		HMW:  []string{"G-ADMIN", "G-CENTER", "G_CENTER", "G_ADMIN"},
		CLIK: []string{"G-ADMIN", "G-CENTER", "G_CENTER", "G_ADMIN"}},
	{Master: "G-CR",
		ACG:  []string{"G-CR", "G_CR"},
		HMW:  []string{"G-CR", "G_CR"},
		CLIK: []string{"G-CR", "G_CR"}},
	{Master: "G-HR",
		ACG:  []string{"G-HR", "G_HR"},
		HMW:  []string{"G-HR", "G_HR"},
		CLIK: []string{"G-HR", "G_HR"}},
	{Master: "G-MAINT",
		ACG:  []string{"G-MAINT", "G_MAINT"},
		HMW:  []string{"G-MAINT", "G_MAINT"},
		CLIK: []string{"G-MAINT", "G_MAINT"}},
	{Master: "G-PC",
		ACG:  []string{"G-PC", "G_PC"},
		HMW:  []string{"G-PC", "G_PC"},
		CLIK: []string{"G-PC", "G_PC"}},
	{Master: "G-SALARY",
		ACG:  []string{"G-SALARY", "G_SALARY"},
		HMW:  []string{"G-SALARY", "G_SALARY"},
		CLIK: []string{"G-SALARY", "G_SALARY"}},
	{Master: "IA",
		ACG:  []string{"IA", "G-IA"},
		HMW:  []string{"IA", "G-IA"},
		CLIK: []string{"IA", "G-IA"}},
	{Master: "IT-DEVELOP",
		ACG:  []string{"IT-DEVELOP", "IT_DEVELOP", "G-IT-DEVELOP"},
		HMW:  []string{"IT-DEVELOP", "IT_DEVELOP", "G-IT-DEVELOP"},
		CLIK: []string{"IT-DEVELOP", "IT_DEVELOP", "G-IT-DEVELOP"}},
	{Master: "IT-SUPPORT",
		ACG:  []string{"IT-SUPPORT", "IT_SUPPORT", "G-IT-SUPPORT"},
		HMW:  []string{"IT-SUPPORT", "IT_SUPPORT", "G-IT-SUPPORT"},
		CLIK: []string{"IT-SUPPORT", "IT_SUPPORT", "G-IT-SUPPORT"}},
	{Master: "MARKETING",
		ACG:  []string{"MARKETING", "MAREKTING", "Marketing_Graphic", "Marketing-Graphic"},
		HMW:  []string{"MARKETING", "MAREKTING", "Marketing_Graphic", "Marketing-Graphic"},
		CLIK: []string{"MARKETING", "MAREKTING", "Marketing_Graphic", "Marketing-Graphic"}},
	{Master: "MGMT",
		ACG:  []string{"MGMT"},
		HMW:  []string{"MGMT"},
		CLIK: []string{"MGMT"}},
	{Master: "SALE-CENTER",
		ACG:  []string{"G-REG", "SALE-CENTER", "SALE_CENTER", "G_REG"},
		HMW:  []string{"G-REG", "SALE-CENTER", "SALE_CENTER", "G_REG"},
		CLIK: []string{"G-REG", "SALE-CENTER", "SALE_CENTER", "G_REG"}},
	{Master: "SALE-INSURE",
		ACG:  []string{"SALE-INSURE", "SALE_INSURE"},
		HMW:  []string{"SALE-INSURE", "SALE_INSURE"},
		CLIK: []string{"SALE-INSURE", "SALE_INSURE"}},
	{Master: "SECRETARY",
		ACG:  []string{"IR", "SECRETARY"},
		HMW:  []string{"IR", "SECRETARY"},
		CLIK: []string{"IR", "SECRETARY"}},
	{Master: "SERVICE",
		ACG:  []string{"BP", "G-PDI", "G_PDI", "GRPM", "G-STORE", "G_STORE", "SERVICE", "SERVICE-BP", "SERVICE_BP", "SERVICE-GR", "SERVICE_GR", "SERVICE-SA", "SERVICE_SA", "SERVICE-CENTER", "SERVICE_CENTER"},
		HMW:  []string{"BP1", "G-PDI", "G_PDI", "GRPM", "G-STORE", "G_STORE", "SERVICE", "SERVICE-BP", "SERVICE_BP", "SERVICE-CENTER", "SERVICE_CENTER", "SERVICE-GR", "SERVICE_GR", "SERVICE-SA", "SERVICE_SA", "SERIVCE-ADMIN", "SERVICE_ADMIN", "SERVICE-CASHIER", "SERVICE_CASHIER", "BP2"},
		CLIK: []string{"G-PDI", "G_PDI", "G-STORE", "G_STORE", "SERVICE-BP", "SERVICE_BP", "SERVICE-CENTER", "SERVICE_CENTER", "SERVICE-GR", "SERVICE_GR", "SERVICE-SA", "SERVICE_SA", "SERVICE-ADMIN", "SERVICE_ADMIN", "SERVICE-CASHIER", "SERVICE_CASHIER"}},
	{Master: "SERVICE_CLIK",
		ACG:  []string{},
		HMW:  []string{},
		CLIK: []string{"SERVICE", "SERVICE_CLIK", "SERVICE-CLIK", "SERVICE CLIK"}},
	{Master: "STRATEGY",
		ACG:  []string{"STRATEGY"},
		HMW:  []string{"STRATEGY"},
		CLIK: []string{"STRATEGY"}},
	{Master: "TRAINING",
		ACG:  []string{"TRAINING"},
		HMW:  []string{"TRAINING"},
		CLIK: []string{"TRAINING"}},
	{Master: "PLANNING",
		ACG:  []string{"PLANNING", "B-PBA", "B_PBA", "SERVICE-SALE", "SERVICE_SALE", "SALE-CONTROL", "SALE_CONTROL", "SERVICE-CONTROL", "SERVICE_CONTROL", "SC"},
		HMW:  []string{"PLANNING", "B-PBA", "B_PBA", "SERVICE-SALE", "SERVICE_SALE", "SALE-CONTROL", "SALE_CONTROL", "SERVICE-CONTROL", "SERVICE_CONTROL", "SC"},
		CLIK: []string{"PLANNING", "B-PBA", "B_PBA", "SERVICE-SALE", "SERVICE_SALE", "SALE-CONTROL", "SALE_CONTROL", "SERVICE-CONTROL", "SERVICE_CONTROL", "SC"}},
	{Master: "None",
		ACG:  []string{"ACC-RECEIVE", "ACC_RECEIVE", "CENTER", "CONSTRUCTION", "G-HK", "G_HK", "G-SECURITY", "G_SECURITY", "INS", "IT", "IT-CENTER", "IT_CENTER", "MNG", "PC", "SALE", "STORE", "TECH", "FIN-PAY", "FIN_PAY", "FIN-RECEIVE", "FIN_RECEIVE", "REG"},
		HMW:  []string{"ACC-RECEIVE", "ACC_RECEIVE", "CENTER", "CONSTRUCTION", "G-HK", "G_HK", "G-SECURITY", "G_SECURITY", "INS", "IT", "IT-CENTER", "IT_CENTER", "MNG", "PC", "SALE", "STORE", "TECH", "FIN-PAY", "FIN_PAY", "FIN-RECEIVE", "FIN_RECEIVE", "REG"},
		CLIK: []string{"ACC-RECEIVE", "ACC_RECEIVE", "CENTER", "CONSTRUCTION", "G-HK", "G_HK", "IT", "IT-CENTER", "IT_CENTER", "PC", "PLANNING", "FIN-PAY", "FIN_PAY", "FIN-RECEIVE", "FIN_RECEIVE"}},
}

// GetMasterDeptCode maps a granular department code to its consolidated Master Code.
func GetMasterDeptCode(granularCode string) string {
	for _, row := range DepartmentMappingData {
		for _, c := range row.ACG {
			if c == granularCode {
				return row.Master
			}
		}
		for _, c := range row.HMW {
			if c == granularCode {
				return row.Master
			}
		}
		for _, c := range row.CLIK {
			if c == granularCode {
				return row.Master
			}
		}
	}
	return "" // Return empty if no mapping found
}
