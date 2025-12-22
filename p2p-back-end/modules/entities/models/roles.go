package models


const (
	RoleAdmin     = "admin"      // ผู้ดูแลระบบสูงสุด
	RoleManager   = "manager"    // หัวหน้าแผนก (อนุมัติได้)
	RolePurchaser = "purchaser"  // ฝ่ายจัดซื้อ
	RoleWarehouse = "warehouse"  // ฝ่ายคลังสินค้า
	RoleFinance   = "finance"    // ฝ่ายบัญชี/การเงิน
	RoleEmployee  = "employee"   // พนักงานทั่วไป
)