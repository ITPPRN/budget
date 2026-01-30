export const STATIC_FILTER_OPTIONS = [
    {
        id: "Admin Expense", name: "Admin Expense", level: 1,
        children: [
            {
                id: "ADM1_Staff expense", name: "ADM1_Staff expense", level: 2,
                children: [
                    {
                        id: "Staff1_Salary&OT", name: "Staff1_Salary&OT", level: 3,
                        children: [
                            { id: "61210010", name: "61210010-เงินเดือน/ค่าล่วงเวลา-บริหาร", level: 4 },
                            { id: "61210011", name: "61210011-เบี้ยเลี้ยง-บริหาร", level: 4 },
                            { id: "61210012", name: "61210012-โบนัส-บริหาร", level: 4 },
                            { id: "61210075", name: "61210075-เบี้ยเลี้ยงนักศึกษาฝึกงาน-บริหาร", level: 4 },
                            { id: "61210078", name: "61210078-ค่าล่วงเวลา-บริหาร", level: 4 },
                            { id: "61210080", name: "61210080-ค่าแรงจูงใจพนักงาน-บริหาร", level: 4 },
                            { id: "61210081", name: "61210081 -เบี้ยขยัน-บริหาร", level: 4 },
                            { id: "61210073-1", name: "61210073-1-ค่านายหน้า-บริหาร", level: 4 },
                            { id: "Dummy-1", name: "Dummy-1-กองทุนทดแทน", level: 4 },
                            { id: "Dummy-2", name: "Dummy-2-กองทุนสงเคราะห์", level: 4 },
                        ]
                    },
                    {
                        id: "Staff10_Commission", name: "Staff10_Commission", level: 3,
                        children: [
                            { id: "61210073-1-comm", name: "61210073-1-ค่านายหน้าในการขาย-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Staff2_Welfare", name: "Staff2_Welfare", level: 3,
                        children: [
                            { id: "61210013", name: "61210013-ค่าสวัสดิการพนักงาน-บริหาร", level: 4 },
                            { id: "61210014", name: "61210014-ค่ารักษาพยาบาล-บริหาร", level: 4 },
                        ]
                    },
                    {
                        id: "Staff3_Director fee", name: "Staff3_Director fee", level: 3,
                        children: [
                            { id: "61210072", name: "61210072-ค่าเบี้ยประชุม", level: 4 },
                            { id: "61210072-salary", name: "61210072-เงินเดือน/กรรมการอิสระ", level: 4 }
                        ]
                    },
                    {
                        id: "Staff4_Other benefits", name: "Staff4_Other benefits", level: 3,
                        children: [
                            { id: "61210009", name: "61210009-กองทุนสำรองเลี้ยงชีพ", level: 4 },
                            { id: "61210015", name: "61210015-เงินสมทบประกันสังคมนายจ้าง-บริหาร", level: 4 },
                            { id: "61210073", name: "61210073 -ผลประโยชน์พนักงานโดยใช้หุ้นเป็นเกณฑ์", level: 4 },
                            { id: "61210079", name: "61210079-ผลประโยช์ของพนักงานอื่น-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Staff5_Training", name: "Staff5_Training", level: 3,
                        children: [
                            { id: "61210032", name: "61210032-การฝึกอบรมและพัฒนาบุคคลากร-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Staff6_Recruitment", name: "Staff6_Recruitment", level: 3,
                        children: [
                            { id: "61210069", name: "61210069-ค่าใช้จ่าย-จัดหาบุคลากร", level: 4 }
                        ]
                    },
                    {
                        id: "Staff7_EBO", name: "Staff7_EBO", level: 3,
                        children: [
                            { id: "61210016", name: "61210016-ผลประโยชน์พนักงานหลังออกจากงาน-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Staff8_Outsorce", name: "Staff8_Outsorce", level: 3,
                        children: [
                            { id: "61210043", name: "61210043-ค่ารักษาความปลอดภัย-บริหาร", level: 4 },
                            { id: "61210044", name: "61210044-ค่าทำความสะอาด-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Staff9_Travelling", name: "Staff9_Travelling", level: 3,
                        children: [
                            { id: "61210026", name: "61210026-ค่าน้ำมันเชื้อเพลิง-บริหาร", level: 4 },
                            { id: "61210026-fleet", name: "61210026-ค่าน้ำมันเชื้อเพลิงFleetcard-บริหาร", level: 4 },
                            { id: "61210030", name: "61210030-ค่าใช้จ่ายในการเดินทาง-บริหาร", level: 4 },
                            { id: "61210063", name: "61210063-ค่าที่พัก", level: 4 }
                        ]
                    }
                ]
            },
            {
                id: "ADM10_Supplies", name: "ADM10_Supplies", level: 2,
                children: [
                    {
                        id: "Supplies1_Supplies", name: "Supplies1_Supplies", level: 3,
                        children: [
                            { id: "61210017", name: "61210017-วัสดุสิ้นเปลืองใช้ไป-บริหาร", level: 4 },
                            { id: "61210052", name: "61210052-ต้นทุนอะไหล่-เบิกใช้ในกิจการ", level: 4 }
                        ]
                    },
                    {
                        id: "Supplies2_Stationary", name: "Supplies2_Stationary", level: 3,
                        children: [
                            { id: "61210018", name: "61210018-เครื่องเขียนแบบพิมพ์-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Supplies3_Stamp duty", name: "Supplies3_Stamp duty", level: 3,
                        children: [
                            { id: "61210022", name: "61210022-ค่าอากรแสตมป์-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Supplies4_Miscellaneous", name: "Supplies4_Miscellaneous", level: 3,
                        children: [
                            { id: "61210047", name: "61210047-ค่าใช้จ่ายเบ็ดเตล็ด-บริหาร", level: 4 }
                        ]
                    }
                ]
            },
            {
                id: "ADM11_Other expense", name: "ADM11_Other expense", level: 2,
                children: [
                    {
                        id: "Charity1_Charity", name: "Charity1_Charity", level: 3,
                        children: [
                            { id: "61210037", name: "61210037-ค่าการกุศล-บริหาร", level: 4 },
                            { id: "61210070", name: "61210070-ค่าใช้จ่าย-ศูนย์บริการทำบุญ", level: 4 }
                        ]
                    },
                    {
                        id: "Conference1_Meeting expense", name: "Conference1_Meeting expense", level: 3,
                        children: [{ id: "61210067", name: "61210067-ค่าใช้จ่าย-การจัดประชุมบริษัท", level: 4 }]
                    },
                    {
                        id: "Debt1_Doubtful expense", name: "Debt1_Doubtful expense", level: 3,
                        children: [{ id: "61210050", name: "61210050-หนี้สงสัยจะสูญ-บริหาร", level: 4 }]
                    },
                    {
                        id: "Debt2_Bad debt", name: "Debt2_Bad debt", level: 3,
                        children: [{ id: "61210051", name: "61210051-หนี้สูญ-บริหาร", level: 4 }]
                    },
                    {
                        id: "Entertain1_Entertain", name: "Entertain1_Entertain", level: 3,
                        children: [
                            { id: "61210036", name: "61210036-ค่ารับรอง-บริหาร", level: 4 },
                            { id: "61210057", name: "61210057-ค่ารับรอง(บวกกลับ)-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Freight1_Freight", name: "Freight1_Freight", level: 3,
                        children: [{ id: "61210046", name: "61210046-ค่าขนส่งรถยนต์-บริหาร", level: 4 }]
                    },
                    {
                        id: "Incident1_Damage expense", name: "Incident1_Damage expense", level: 3,
                        children: [{ id: "61210071", name: "61210071-ค่าใช้จ่าย-ความเสียหายระหว่างงานซ่อม", level: 4 }]
                    },
                    {
                        id: "Insur1_Building insurance", name: "Insur1_Building insurance", level: 3,
                        children: [{ id: "61210031", name: "61210031-ค่าเบี้ยประกันภัย-บริหาร", level: 4 }]
                    },
                    {
                        id: "IR1_Investor", name: "IR1_Investor", level: 3,
                        children: [
                            { id: "61210066", name: "61210066-ค่าใช้จ่าย-นักลงทุนสัมพันธ์", level: 4 },
                            { id: "61210074", name: "61210074-ค่าประชาสัมพันธ์-นักลงทุนสัมพันธ์", level: 4 }
                        ]
                    },
                    {
                        id: "IT1_Software&Systems", name: "IT1_Software&Systems", level: 3,
                        children: [{ id: "61210068", name: "61210068-ค่าใช้จ่าย-ระบบสารสนเทศ", level: 4 }]
                    },
                    {
                        id: "NonD1_Non-deduct expense", name: "NonD1_Non-deduct expense", level: 3,
                        children: [{ id: "61210058", name: "61210058-ค่าใช้จ่ายที่ไม่ใช่ค่ารับอรง (บวกกลับ)-บริหาร", level: 4 }]
                    },
                    {
                        id: "Other1_Other expense", name: "Other1_Other expense", level: 3,
                        children: [{ id: "61210054", name: "61210054-ค่าใช้จ่ายอื่น-บริหาร", level: 4 }]
                    },
                    {
                        id: "Postage1_Postage", name: "Postage1_Postage", level: 3,
                        children: [
                            { id: "61210028", name: "61210028-ค่าขนส่งพัสดุ-บริหาร", level: 4 },
                            { id: "61210045", name: "61210045-ค่าขนส่งอื่น-ไม่ใช่รถยนต์และอะไหล่", level: 4 },
                            { id: "61210045-2", name: "61210045-ค่าขนส่ง", level: 4 },
                            { id: "61210046-2", name: "61210046-ค่าขนส่ง-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Regis1_Registration Fee", name: "Regis1_Registration Fee", level: 3,
                        children: [{ id: "61210053", name: "61210053-ค่าแจ้งย้ายรถใหม่งานทะเบียน-บริหาร", level: 4 }]
                    },
                    {
                        id: "Regis2_Vehical Tax", name: "Regis2_Vehical Tax", level: 3,
                        children: [{ id: "61210029", name: "61210029-ค่าทะเบียนภาษีรถยนต์-บริหาร", level: 4 }]
                    }
                ]
            },
            {
                id: "ADM2_Share service", name: "ADM2_Share service", level: 2,
                children: [
                    {
                        id: "Share1_Share service", name: "Share1_Share service", level: 3,
                        children: [{ id: "61210056", name: "61210056-ค่าบริหารจัดการ-บริหาร", level: 4 }]
                    }
                ]
            },
            {
                id: "ADM3_Rent & Facilities", name: "ADM3_Rent & Facilities", level: 2,
                children: [
                    {
                        id: "Facilities1_Water Supply", name: "Facilities1_Water Supply", level: 3,
                        children: [{ id: "61210025", name: "61210025-ค่าน้ำประปา-บริหาร", level: 4 }]
                    },
                    {
                        id: "Facilities2_Electricity", name: "Facilities2_Electricity", level: 3,
                        children: [{ id: "61210024", name: "61210024-ค่าไฟฟ้า-บริหาร", level: 4 }]
                    },
                    {
                        id: "Facilities3_Telephone", name: "Facilities3_Telephone", level: 3,
                        children: [{ id: "61210027", name: "61210027-ค่าโทรศัพท์-บริหาร", level: 4 }]
                    },
                    {
                        id: "Rent1_Office supplies", name: "Rent1_Office supplies", level: 3,
                        children: [
                            { id: "61210019", name: "61210019-ค่าเช่าเครื่องถ่ายเอกสาร-บริหาร", level: 4 },
                            { id: "61210021", name: "61210021-ค่าเช่าอื่น", level: 4 }
                        ]
                    },
                    {
                        id: "Rent2_Land", name: "Rent2_Land", level: 3,
                        children: [{ id: "61210020", name: "61210020-ค่าเช่าที่ดิน-บริหาร", level: 4 }]
                    },
                    {
                        id: "Rent3_Vehicle", name: "Rent3_Vehicle", level: 3,
                        children: [{ id: "61210021-vehicle", name: "61210021-ค่าเช่ารถยนต์-บริหาร", level: 4 }]
                    }
                ]
            },
            {
                id: "ADM4_D&A", name: "ADM4_D&A", level: 2,
                children: [
                    {
                        id: "D&A1_Depreciation", name: "D&A1_Depreciation", level: 3,
                        children: [
                            { id: "61210034", name: "61210034-ค่าเสื่อมราคา-บริหาร", level: 4 },
                            { id: "61210062", name: "61210062-ค่าเสื่อมราคาสัญญาเช่า-บริหาร", level: 4 },
                            { id: "61210060-1", name: "61210060-1-ค่าเสื่อมราคาสัญญาเช่า-บริหาร", level: 4 },
                            { id: "61210072-1", name: "61210072-1-ค่าเสื่อมราคา-บริหาร-IP", level: 4 }
                        ]
                    },
                    {
                        id: "D&A2_Amortization", name: "D&A2_Amortization", level: 3,
                        children: [
                            { id: "61210035", name: "61210035-ค่าลิขสิทธิ์โปรแกรมตัดจ่าย-บริหาร", level: 4 },
                            { id: "61210065", name: "61210065-ค่าโปรแกรมตัดจ่าย-บริหาร", level: 4 }
                        ]
                    }
                ]
            },
            {
                id: "ADM5_Professional fee", name: "ADM5_Professional fee", level: 2,
                children: [
                    {
                        id: "Prof1_Consult", name: "Prof1_Consult", level: 3,
                        children: [
                            { id: "61210042", name: "61210042-ค่าที่ปรึกษาทางกฏหมาย-บริหาร", level: 4 },
                            { id: "61210055", name: "61210055-ค่าที่ปรึกษาอื่น-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Prof2_Professional fee", name: "Prof2_Professional fee", level: 3,
                        children: [{ id: "61210041", name: "61210041-ค่าธรรมเนียมวิชาชีพอิสระ-บริหาร", level: 4 }]
                    }
                ]
            },
            {
                id: "ADM6_Maintenance", name: "ADM6_Maintenance", level: 2,
                children: [
                    {
                        id: "Maint1_Repair", name: "Maint1_Repair", level: 3,
                        children: [
                            { id: "61210033", name: "61210033-ค่าซ่อมแซมบำรุงรักษา-บริหาร", level: 4 },
                            { id: "61210062-repair", name: "61210062-ค่าซ่อมแซมรถใหม่", level: 4 }
                        ]
                    }
                ]
            },
            {
                id: "ADM7_Loss on", name: "ADM7_Loss on", level: 2,
                children: [
                    {
                        id: "Loss4_Write-off", name: "Loss4_Write-off", level: 3,
                        children: [{ id: "71110013", name: "71110013-ผลขาดทุนจากการตัดจำหน่าย (Write off)", level: 4 }]
                    },
                    {
                        id: "Loss5_Adjustment", name: "Loss5_Adjustment", level: 3,
                        children: [{ id: "61210049", name: "61210049-ส่วนต่างจากการปรับปรุงบัญชีสินค้าคงเหลือ-บริหาร", level: 4 }]
                    },
                    {
                        id: "Loss6_Inventory write-off (Tax)", name: "Loss6_Inventory write-off (Tax)", level: 3,
                        children: [
                            { id: "71110014", name: "71110014-รายจ่ายจากการทำลายสินค้าคงเหลือทางภาษี", level: 4 },
                            { id: "71110015", name: "71110015-รายจ่ายจากการทำลายสินทรัพย์ถาวรทางภาษี", level: 4 }
                        ]
                    }
                ]
            },
            {
                id: "ADM8_Other tax", name: "ADM8_Other tax", level: 2,
                children: [
                    {
                        id: "Othertax1_Local tax", name: "Othertax1_Local tax", level: 3,
                        children: [
                            { id: "61210023", name: "61210023-ค่าภาษีป้ายและภาษีโรงเรือน-บริหาร", level: 4 },
                            { id: "61210057-1", name: "61210057-1-ภาษีที่ดินและภาษีโรงเรือน-บริหาร", level: 4 },
                            { id: "61210074-1", name: "61210074-1 -ค่าภาษีที่ดินและสิ่งปลูกสร้าง", level: 4 },
                            { id: "61210074-2", name: "61210074-2-ภาษีโรงเรือน และ สิ่งปลูกสร้าง-บริหาร", level: 4 }
                        ]
                    },
                    {
                        id: "Othertax2_SPE", name: "Othertax2_SPE", level: 3,
                        children: [{ id: "61210064", name: "61210064-ภาษีธุรกิจเฉพาะ", level: 4 }]
                    },
                    {
                        id: "Othertax3_Surcharge", name: "Othertax3_Surcharge", level: 3,
                        children: [{ id: "61210048", name: "61210048-ค่าเบี้ยปรับเงินเพิ่ม-บริหาร", level: 4 }]
                    }
                ]
            },
            {
                id: "ADM9_Fee", name: "ADM9_Fee", level: 2,
                children: [
                    {
                        id: "Finance4_Fee", name: "Finance4_Fee", level: 3,
                        children: [
                            { id: "61210038", name: "61210038-ค่าธรรมเนียมโอนเงิน-บริหาร", level: 4 },
                            { id: "61210039", name: "61210039-ค่าธรรมเนียมธุรกรรมทางการเงิน-บริหาร", level: 4 },
                            { id: "61210040", name: "61210040-ค่าธรรมเนียมอื่น-บริหาร", level: 4 }
                        ]
                    }
                ]
            }
        ]
    },
    {
        id: "Finance Cost", name: "Finance Cost", level: 1,
        children: [
            {
                id: "FIN1_Finance cost", name: "FIN1_Finance cost", level: 2,
                children: [
                    {
                        id: "Interest expense (Demolish)", name: "Interest expense (Demolish)", level: 3,
                        children: [{ id: "81110013", name: "81110013-ดอกเบี้ยตัดจำหน่าย-ค่ารื้อถอน", level: 4 }]
                    },
                    {
                        id: "Interest expense (Lease)", name: "Interest expense (Lease)", level: 3,
                        children: [
                            { id: "51310044", name: "51310044-ดอกเบี้ยจ่ายสัญาเช่า-บริการ", level: 4 },
                            { id: "61210061", name: "61210061-ดอกเบี้ยจ่ายสัญาเช่า-บริการ", level: 4 },
                            { id: "61110027", name: "61110027-ดอกเบี้ยจ่ายสัญาเช่า-ขาย", level: 4 }
                        ]
                    },
                    {
                        id: "Interest expense (Loan)", name: "Interest expense (Loan)", level: 3,
                        children: [
                            { id: "81110004", name: "81110004-ดอกเบี้ยจ่ายกิจการที่เกี่ยวข้องกัน", level: 4 },
                            { id: "81110005", name: "81110005-ดอกเบี้ยจ่ายกิจการอื่น", level: 4 },
                            { id: "81110012", name: "81110012-ดอกเบี้ยจ่ายเงินกู้ยืมระยะยาวธนาคาร", level: 4 }
                        ]
                    },
                    {
                        id: "Interest expense (OD)", name: "Interest expense (OD)", level: 3,
                        children: [{ id: "81110010_od", name: "81110010-ดอกเบี้ยจ่ายเงินเบิกเกินบัญชีธนาคาร", level: 4 }]
                    },
                    {
                        id: "Interest expense (PN)", name: "Interest expense (PN)", level: 3,
                        children: [{ id: "81110011_pn", name: "81110011-ดอกเบี้ยจ่ายตั๋วเงินจ่าย", level: 4 }]
                    },
                    {
                        id: "Other interest expense", name: "Other interest expense", level: 3,
                        children: [{ id: "81110013_other", name: "81110013-ดอกเบี้ยจ่ายอื่น", level: 4 }]
                    }
                ]
            }
        ]
    },
    {
        id: "Selling Expense", name: "Selling Expense", level: 1,
        children: [
            {
                id: "SELL1_Staff expense", name: "SELL1_Staff expense", level: 2,
                children: [
                    {
                        id: "Staff1_Salary&OT", name: "Staff1_Salary&OT", level: 3,
                        children: [
                            { id: "61110010", name: "61110010-ค่านายหน้าในการขาย", level: 4 },
                            { id: "61110015", name: "61110015-ค่านายหน้าอื่น", level: 4 },
                            { id: "61110022", name: "61110022-เงินเดือน/ค่าล่วงเวลา-ขาย", level: 4 },
                            { id: "61110026", name: "61110026-เบี้ยเลี้ยง-ขาย", level: 4 },
                            { id: "61110053", name: "61110053-ค่าแรงจูงใจพนักงาน-ขาย", level: 4 },
                            { id: "61110036", name: "61110036-ค่านายหน้าในการขาย(ตัวแทนขาย)", level: 4 },
                            { id: "61110040", name: "61110040-เงินเดือน/ค่าล่วงเวลาขาย - ศูนย์บริการ", level: 4 },
                            { id: "61110041", name: "61110041-เงินเดือน/ค่าล่วงเวลาขาย - ประกันภัย", level: 4 },
                            { id: "61110042", name: "61110042-ค่านายหน้าในการขาย - ศูนย์บริการ", level: 4 },
                            { id: "61110043", name: "61110043-ค่านายหน้าในการขาย - ประกันภัย", level: 4 },
                            { id: "61110044", name: "61110044-เบี้ยเลี้ยงขาย - ศูนย์บริการ", level: 4 },
                            { id: "61110050", name: "61110050-เบี้ยขยัน-ขาย", level: 4 },
                            { id: "61110054", name: "61110054-ค่าล่วงเวลา- ศูนย์บริการ", level: 4 },
                            { id: "61110056", name: "61110056-เบี้ยขยัน - ศูนย์บริการ", level: 4 },
                            { id: "61110057", name: "61110057-เบี้ยขยัน - ประกันภัย", level: 4 },
                            { id: "61110060", name: "61110060-ค่าแรงจูงใจพนักงาน-ศูนย์บริการ", level: 4 },
                            { id: "61110061", name: "61110061-ค่าแรงจูงใจพนักงาน-ประกันภัย", level: 4 }
                        ]
                    },
                    {
                        id: "Staff2_Welfare", name: "Staff2_Welfare", level: 3,
                        children: [{ id: "61110020", name: "61110020-ค่าสวัสดิการพนักงาน-ขาย", level: 4 }]
                    },
                    {
                        id: "Staff4_Other benefits", name: "Staff4_Other benefits", level: 3,
                        children: [
                            { id: "61110023", name: "61110023-เงินสมทบประกันสังคมนายจ้าง-ขาย", level: 4 },
                            { id: "61110024", name: "61110024-กองทุนสํารองเลี้ยงชีพ-ขาย", level: 4 },
                            { id: "61110046", name: "61110046-กองทุนสํารองเลี้ยงชีพขาย - ศูนย์บริการ", level: 4 },
                            { id: "61110047", name: "61110047-กองทุนสํารองเลี้ยงชีพขาย - ประกันภัย", level: 4 },
                            { id: "61110048", name: "61110048-เงินสมทบประกันสังคมนายจ้างขาย - ศูนย์บริการ", level: 4 },
                            { id: "61110049", name: "61110049-เงินสมทบประกันสังคมนายจ้างขาย - ประกันภัย", level: 4 },
                            { id: "61110052", name: "61110052-ผลประโยชของพนักงานอื่น-ขาย", level: 4 },
                            { id: "61110058", name: "61110058-ผลประโยชของพนักงานอื่น-ศูนย์บริการ", level: 4 }
                        ]
                    },
                    {
                        id: "Staff7_EBO", name: "Staff7_EBO", level: 3,
                        children: [{ id: "61110019", name: "61110019-ผลประโยชน์พนักงานหลังออกจากงาน-ขาย", level: 4 }]
                    },
                    {
                        id: "Staff9_Travelling", name: "Staff9_Travelling", level: 3,
                        children: [
                            { id: "61110021", name: "61110021-ค่าน้ำมันเชื้อเพลิง-ขาย", level: 4 },
                            { id: "61110033", name: "61110033-ค่าใช้จ่ายในการเดินทาง-ขาย", level: 4 }
                        ]
                    }
                ]
            },
            {
                id: "SELL2_Marketing expense", name: "SELL2_Marketing expense", level: 2,
                children: [
                    {
                        id: "Market1_Advertising", name: "Market1_Advertising", level: 3,
                        children: [{ id: "61110013", name: "61110013-ค่าโฆษณา", level: 4 }]
                    },
                    {
                        id: "Market2_Promotion", name: "Market2_Promotion", level: 3,
                        children: [{ id: "61110012", name: "61110012-ค่าส่งเสริมการขาย-อื่น", level: 4 }]
                    },
                    {
                        id: "Market3_Event", name: "Market3_Event", level: 3,
                        children: [{ id: "61110014", name: "61110014-ค่าจัดกิจกรรมทางการตลาด", level: 4 }]
                    }
                ]
            },
            {
                id: "SELL3_D&A", name: "SELL3_D&A", level: 2,
                children: [
                    {
                        id: "D&A1_Depreciation", name: "D&A1_Depreciation", level: 3,
                        children: [
                            { id: "61110016", name: "61110016-ค่าเสื่อมราคา-ขาย", level: 4 },
                            { id: "61110028", name: "61110028-ค่าเสื่อมราคาสัญญาเช่า-ขาย", level: 4 }
                        ]
                    }
                ]
            },
            {
                id: "SELL4_Maintenance", name: "SELL4_Maintenance", level: 2,
                children: [
                    {
                        id: "Maint1_Repair", name: "Maint1_Repair", level: 3,
                        children: [{ id: "61110032", name: "61110032-ค่าซ่อมแซมบำรุงรักษา-ขาย", level: 4 }]
                    }
                ]
            },
            {
                id: "SELL5_Other selling", name: "SELL5_Other selling", level: 2,
                children: [
                    {
                        id: "Finance4_Fee", name: "Finance4_Fee", level: 3,
                        children: [
                            { id: "61110029", name: "61110029-ค่าธรรมเนียมการขาย-ออนไลน์", level: 4 },
                            { id: "61110030", name: "61110030-ค่าธุรกรรมการชําระเงินออนไลน์", level: 4 },
                            { id: "61110038", name: "61110038-ค่าธรรมเนียมธุรกรรมทางการเงิน-ขาย", level: 4 }
                        ]
                    },
                    {
                        id: "Freight1_Freight", name: "Freight1_Freight", level: 3,
                        children: [
                            { id: "61110018", name: "61110018-ค่าขนส่งรถยนต์และอะไหล่-ขาย", level: 4 },
                            { id: "61110031", name: "61110031-ค่าขนส่ง-ออนไลน์", level: 4 }
                        ]
                    },
                    {
                        id: "Miscellaneous expenses", name: "Miscellaneous expenses", level: 3,
                        children: [{ id: "61110034", name: "61110034-ค่าใช้จ่ายเบ็ดเตล็ด-ขาย", level: 4 }]
                    },
                    {
                        id: "Rent2_Land", name: "Rent2_Land", level: 3,
                        children: [{ id: "61110025", name: "61110025-ค่าเช่าที่ดิน-ขาย", level: 4 }]
                    }
                ]
            }
        ]
    },
    {
        id: "Service Cost", name: "Service Cost", level: 1,
        children: [
            {
                id: "SVC10_Other costs", name: "SVC10_Other costs", level: 2,
                children: [
                    {
                        id: "Postage1_Postage", name: "Postage1_Postage", level: 3,
                        children: [
                            { id: "51310030", name: "51310030-ค่าขนส่งพัสดุ-บริการ", level: 4 },
                            { id: "51310040", name: "51310040-ค่าขนส่งอื่น-ไม่ใช่รถยนต์และอะไหล่-บริการ", level: 4 }
                        ]
                    },
                    {
                        id: "Rent2_Land", name: "Rent2_Land", level: 3,
                        children: [{ id: "51310023", name: "51310023-ค่าเช่าที่ดิน-บริการ", level: 4 }]
                    },
                    {
                        id: "Supplies1_Supplies", name: "Supplies1_Supplies", level: 3,
                        children: [{ id: "51310020", name: "51310020-วัสดุสิ้นเปลืองใช้ไป-บริการ", level: 4 }]
                    },
                    {
                        id: "Supplies2_Stationary", name: "Supplies2_Stationary", level: 3,
                        children: [{ id: "51310021", name: "51310021-เครื่องเขียนแบบพิมพ์-บริการ", level: 4 }]
                    },
                    {
                        id: "Supplies4_Miscellaneous", name: "Supplies4_Miscellaneous", level: 3,
                        children: [{ id: "51310042", name: "51310042-ค่าใช้จ่ายเบ็ดเตล็ด-บริการ", level: 4 }]
                    },
                    {
                        id: "ว่าง", name: "ว่าง", level: 3,
                        children: [{ id: "51310012", name: "51310012-ต้นทุนค่าบริการ-อื่น", level: 4 }]
                    }
                ]
            },
            {
                id: "SVC7_Overhead", name: "SVC7_Overhead", level: 2,
                children: [
                    {
                        id: "Staff1_Salary&OT", name: "Staff1_Salary&OT", level: 3,
                        children: [
                            { id: "51310010", name: "51310010-เงินเดือน/ค่าล่วงเวลา-บริการ", level: 4 },
                            { id: "51310013", name: "51310013-ค่านายหน้า-บริการ", level: 4 },
                            { id: "51310014", name: "51310014-เบี้ยเลี้ยง-บริการ", level: 4 },
                            { id: "51310015", name: "51310015-โบนัส-บริการ", level: 4 },
                            { id: "51310048", name: "51310048-เบี้ยเลี้ยงนักศึกษาฝึกงาน บริการ", level: 4 },
                            { id: "51310050", name: "51310050-เบี้ยขยัน-บริการ", level: 4 },
                            { id: "51310051", name: "51310051-ค่าล่วงเวลา-บริการ", level: 4 },
                            { id: "51310053", name: "51310053-ค่าแรงจูงใจพนักงาน-บริการ", level: 4 }
                        ]
                    },
                    {
                        id: "Staff2_Welfare", name: "Staff2_Welfare", level: 3,
                        children: [
                            { id: "51310016", name: "51310016-ค่าสวัสดิการพนักงาน-บริการ", level: 4 },
                            { id: "51310017", name: "51310017-ค่ารักษาพยาบาล-บริการ", level: 4 }
                        ]
                    },
                    {
                        id: "Staff4_Other benefits", name: "Staff4_Other benefits", level: 3,
                        children: [
                            { id: "51310018", name: "51310018-เงินสมทบประกันสังคมนายจ้าง-บริการ", level: 4 },
                            { id: "51310043", name: "51310043-กองทุนสํารองเลี้ยงชีพ-บริการ", level: 4 },
                            { id: "51310052", name: "51310052-ผลประโยชของพนักงานอื่น-บริการ", level: 4 }
                        ]
                    },
                    {
                        id: "Staff5_Training", name: "Staff5_Training", level: 3,
                        children: [{ id: "51310036", name: "51310036-การอบรมและพัฒนาบุคลากร-บริการ", level: 4 }]
                    },
                    {
                        id: "Staff7_EBO", name: "Staff7_EBO", level: 3,
                        children: [{ id: "51310019", name: "51310019-ผลประโยชน์พนักงานหลังออกจากงาน-บริการ", level: 4 }]
                    },
                    {
                        id: "Staff8_Outsorce", name: "Staff8_Outsorce", level: 3,
                        children: [
                            { id: "51310033", name: "51310033-ค่ารักษาความปลอดภัย-บริการ", level: 4 },
                            { id: "51310034", name: "51310034-ค่าทำความสะอาด-บริการ", level: 4 }
                        ]
                    },
                    {
                        id: "Staff9_Travelling", name: "Staff9_Travelling", level: 3,
                        children: [
                            { id: "51310028", name: "51310028-ค่าน้ำมันเชื้อเพลิง-บริการ", level: 4 },
                            { id: "51310032", name: "51310032-ค่าใช้จ่ายในการเดินทาง-บริการ", level: 4 },
                            { id: "51310035-1", name: "51310035-1-ค่าที่พัก-บริการ", level: 4 }
                        ]
                    }
                ]
            },
            {
                id: "SVC8_D&A", name: "SVC8_D&A", level: 2,
                children: [
                    {
                        id: "D&A1_Depreciation", name: "D&A1_Depreciation", level: 3,
                        children: [
                            { id: "51310038", name: "51310038-ค่าเสื่อมราคา-บริการ", level: 4 },
                            { id: "51310045", name: "51310045-ค่าเสื่อมราคาสัญญาเช่า-บริการ", level: 4 }
                        ]
                    },
                    {
                        id: "D&A2_Amortization", name: "D&A2_Amortization", level: 3,
                        children: [{ id: "51310039", name: "51310039-ค่าลิขสิทธิ์โปรแกรมตัดจ่าย-บริการ", level: 4 }]
                    }
                ]
            },
            {
                id: "SVC9_Maintenance", name: "SVC9_Maintenance", level: 2,
                children: [
                    {
                        id: "Maint1_Repair", name: "Maint1_Repair", level: 3,
                        children: [
                            { id: "51310037", name: "51310037-ค่าซ่อมแซมบำรุงรักษา-บริการ", level: 4 },
                            { id: "51310038-repair", name: "51310038-ค่าเสื่อมราคา-บริการ", level: 4 }
                        ]
                    }
                ]
            }
        ]
    },
    {
        id: "Tax Expense", name: "Tax Expense", level: 1,
        children: [
            {
                id: "TAX1_CIT", name: "TAX1_CIT", level: 2,
                children: [
                    {
                        id: "Coporate income tax", name: "Coporate income tax", level: 3,
                        children: [{ id: "90000000", name: "90000000-ภาษีเงินได้นิติบุคคล", level: 4 }]
                    }
                ]
            },
            {
                id: "TAX2_Deferred tax", name: "TAX2_Deferred tax", level: 2,
                children: [
                    {
                        id: "Deferred tax expense", name: "Deferred tax expense", level: 3,
                        children: [{ id: "90000020", name: "90000020-ค่าใช้จ่ายภาษีเงินได้รอตัดบัญชี", level: 4 }]
                    }
                ]
            }
        ]
    }
];
