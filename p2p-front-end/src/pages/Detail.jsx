// import React, { useState, useEffect, useMemo } from 'react';
// import { Box, FormControl, InputLabel, Select, MenuItem, Stack } from '@mui/material';
// import api from '../utils/api/axiosInstance';
// import { BudgetProvider, useBudget } from '../contexts/BudgetContext';
// import FilterPane from '../components/Budget/FilterPane';
// import BudgetTable from '../components/Budget/BudgetTable';
// import ActualTable from '../components/Budget/ActualTable';
// import { downloadExcelFile } from '../utils/exportUtils';

// // Inner component that consumes the context
// const DetailContent = () => {
//   const { selectedLeaves, getAllLeafIds } = useBudget();

//   // Data Fetching State
//   const [budgetDetails, setBudgetDetails] = useState([]);
//   const [loadingDetails, setLoadingDetails] = useState(false);
//   const [actualDetails, setActualDetails] = useState([]);
//   const [loadingActuals, setLoadingActuals] = useState(false);

//   // Date Filter State for Actuals
//   const [actualDateFilter, setActualDateFilter] = useState({ startDate: '', endDate: '' });

//   // Pagination State for Actuals
//   const [actualPage, setActualPage] = useState(0);
//   const [actualRowsPerPage, setActualRowsPerPage] = useState(10);
//   const [actualTotalCount, setActualTotalCount] = useState(0);

//   // Filters State
//   const [selectedEntity, setSelectedEntity] = useState('');
//   const [selectedBranch, setSelectedBranch] = useState('');
//   const [selectedDepartment, setSelectedDepartment] = useState(''); // New State
//   const [orgStructure, setOrgStructure] = useState([]);

//   // Fetch Filter Options
//   useEffect(() => {
//     const fetchFilters = async () => {
//       try {
//         const res = await api.get('/budgets/organization-structure');
//         setOrgStructure(res.data || []);
//       } catch (err) {
//         console.error("Filter Fetch Error", err);
//       }
//     };
//     fetchFilters();
//   }, []);

//   // Derived state for branches
//   const availableBranches = useMemo(() => {
//     if (!selectedEntity) return [];
//     const entityObj = orgStructure.find(o => o.entity === selectedEntity);
//     return entityObj ? entityObj.branches : [];
//   }, [selectedEntity, orgStructure]);

//   // Derived state for departments
//   const availableDepartments = useMemo(() => {
//     if (selectedBranch) {
//       const branchObj = availableBranches.find(b => b.name === selectedBranch);
//       return branchObj ? branchObj.departments : [];
//     } else {
//       // Flatten all unique departments across entire org if no branch selected
//       const depts = new Set();
//       orgStructure.forEach(entity => {
//         entity.branches?.forEach(branch => {
//           branch.departments?.forEach(dept => {
//             depts.add(dept);
//           });
//         });
//       });
//       return Array.from(depts).sort();
//     }
//   }, [selectedBranch, availableBranches, orgStructure]);

//   // Selection States
//   const [syncConfig, setSyncConfig] = useState({
//       actualYear: new Date().getFullYear(),
//       selectedMonths: [],
//       selectedBudget: "",
//       selectedCapexBg: "",
//       selectedCapexActual: ""
//   });

//   // Auto Fetch Details when Selection Changes or Date Filter Changes
//   useEffect(() => {
//     let isMounted = true;

//     const fetchDetails = async () => {
//       // Logic: If selection is empty -> Fetch ALL (send empty list to backend to optimize)
//       const idsToFetch = selectedLeaves.size > 0
//         ? Array.from(selectedLeaves).map(id => id.split('|')[0]).filter(code => code !== "")
//         : [];

//       if (isMounted) {
//         setLoadingDetails(true);
//         setLoadingActuals(true);
//       }

//       try {
//         let currentSync = {};
//         try {
//           const configRes = await api.get('/budgets/configs');
//           const configs = configRes.data || {};

//           const parseMonths = (val) => {
//             if (Array.isArray(val)) return val;
//             if (typeof val === 'string') {
//               try { return JSON.parse(val); } catch (e) { return []; }
//             }
//             return [];
//           };

//           currentSync = {
//             actualYear: configs.actualYear || configs.actual_year || syncConfig.actualYear || new Date().getFullYear(),
//             selectedMonths: parseMonths(configs.selectedMonths || configs.selected_months) || syncConfig.selectedMonths || [],
//             selectedBudget: configs.selectedBudget || configs.selected_budget || syncConfig.selectedBudget || "",
//             selectedCapexBg: configs.selectedCapexBg || configs.selected_capex_bg || syncConfig.selectedCapexBg || "",
//             selectedCapexActual: configs.selectedCapexActual || configs.selected_capex_actual || syncConfig.selectedCapexActual || ""
//           };
//           if (isMounted) setSyncConfig(currentSync);
//         } catch (e) {
//           console.error("Failed to fetch fallback configs in Detail", e);
//           const cached = JSON.parse(localStorage.getItem('dm_lastSyncedConfig') || '{}');
//           if (cached.actualYear) {
//             currentSync = { ...currentSync, ...cached };
//             if (isMounted) setSyncConfig(currentSync);
//           }
//         }

//         const payload = {
//           conso_gls: idsToFetch,
//           start_date: actualDateFilter.startDate,
//           end_date: actualDateFilter.endDate,
//           entities: selectedEntity ? [selectedEntity] : [],
//           branches: selectedBranch ? [selectedBranch] : [],
//           departments: selectedDepartment ? [selectedDepartment] : [],
//           year: String(currentSync.actualYear),
//           months: currentSync.selectedMonths,
//           budget_file_id: currentSync.selectedBudget,
//           capex_file_id: currentSync.selectedCapexBg,
//           capex_actual_file_id: currentSync.selectedCapexActual,
//           page: actualPage,
//           limit: actualRowsPerPage
//         };

//         // Fetch Budget (Fast)
//         api.post('/budgets/details', payload)
//           .then(res => {
//             if (!isMounted) return;
//             const rawBudget = res.data || [];
//             const budgetMap = new Map();
//             rawBudget.forEach(item => {
//               const key = `${item.conso_gl}|${item.gl_name}`;
//               if (!budgetMap.has(key)) {
//                 budgetMap.set(key, { ...item, budget_amounts: [...(item.budget_amounts || [])] });
//               } else {
//                 const existing = budgetMap.get(key);
//                 existing.year_total = (parseFloat(existing.year_total) || 0) + (parseFloat(item.year_total) || 0);
//                 const existingAmounts = existing.budget_amounts;
//                 item.budget_amounts?.forEach(newAmt => {
//                   const match = existingAmounts.find(ea => ea.month === newAmt.month);
//                   if (match) match.amount = (parseFloat(match.amount) || 0) + (parseFloat(newAmt.amount) || 0);
//                   else existingAmounts.push({ ...newAmt });
//                 });
//               }
//             });
//             setBudgetDetails(Array.from(budgetMap.values()));
//           })
//           .catch(err => {
//             console.error("Budget Details Fetch Failed", err);
//             if (isMounted) setBudgetDetails([]);
//           })
//           .finally(() => {
//             if (isMounted) setLoadingDetails(false);
//           });

//         // Fetch Actuals (May be slower due to large DB)
//         api.post('/budgets/actuals-transactions', payload)
//           .then(res => {
//             if (!isMounted) return;
//             const resMap = res.data || {};
//             setActualDetails(resMap.data || []);
//             setActualTotalCount(resMap.total_count || 0);
//           })
//           .catch(err => {
//             console.error("Actual Transactions Fetch Failed", err);
//             if (isMounted) {
//               setActualDetails([]);
//               setActualTotalCount(0);
//             }
//           })
//           .finally(() => {
//             if (isMounted) setLoadingActuals(false);
//           });

//       } catch (err) {
//         console.error("Fetch Setup Error", err);
//         if (isMounted) {
//           setLoadingDetails(false);
//           setLoadingActuals(false);
//         }
//       }
//     };

//     // Debounce slightly to avoid rapid re-fetches if selection changes fast
//     const timeoutId = setTimeout(() => {
//       fetchDetails();
//     }, 300);

//     return () => {
//       isMounted = false;
//       clearTimeout(timeoutId);
//     };
//   }, [selectedLeaves, getAllLeafIds, actualDateFilter, selectedEntity, selectedBranch, selectedDepartment, actualPage, actualRowsPerPage]); // Add dependencies

//   // --- Export Handlers ---
//   const handleBudgetExport = async () => {
//     const idsToFetch = selectedLeaves.size > 0
//         ? Array.from(selectedLeaves).map(id => id.split('|')[0]).filter(code => code !== "")
//         : [];

//     let syncConfig = JSON.parse(localStorage.getItem('dm_lastSyncedConfig') || '{}');
//     const actualYear = syncConfig.actualYear || new Date().getFullYear();
//     const selectedMonths = Array.isArray(syncConfig.selectedMonths) ? syncConfig.selectedMonths : [];

//     const payload = {
//         conso_gls: idsToFetch,
//         entities: selectedEntity ? [selectedEntity] : [],
//         branches: selectedBranch ? [selectedBranch] : [],
//         departments: selectedDepartment ? [selectedDepartment] : [],
//         year: String(actualYear),
//         months: selectedMonths,
//         budget_file_id: syncConfig.selectedBudget,
//         capex_file_id: syncConfig.selectedCapexBg,
//         capex_actual_file_id: syncConfig.selectedCapexActual,
//         // No pagination for export
//     };

//     await downloadExcelFile('/export-budget-detail', payload, `Budget_Detail_Report_${actualYear}.xlsx`);
//   };

//   const handleActualExport = async () => {
//     const idsToFetch = selectedLeaves.size > 0
//         ? Array.from(selectedLeaves).map(id => id.split('|')[0]).filter(code => code !== "")
//         : [];

//     let syncConfig = JSON.parse(localStorage.getItem('dm_lastSyncedConfig') || '{}');
//     const actualYear = syncConfig.actualYear || new Date().getFullYear();
//     const selectedMonths = Array.isArray(syncConfig.selectedMonths) ? syncConfig.selectedMonths : [];

//     const payload = {
//         conso_gls: idsToFetch,
//         start_date: actualDateFilter.startDate,
//         end_date: actualDateFilter.endDate,
//         entities: selectedEntity ? [selectedEntity] : [],
//         branches: selectedBranch ? [selectedBranch] : [],
//         departments: selectedDepartment ? [selectedDepartment] : [],
//         year: String(actualYear),
//         months: selectedMonths,
//         budget_file_id: syncConfig.selectedBudget,
//         capex_file_id: syncConfig.selectedCapexBg,
//         capex_actual_file_id: syncConfig.selectedCapexActual,
//         // No pagination for export
//     };

//     await downloadExcelFile('/export-actual-detail', payload, `Actual_Detail_Report_${actualYear}.xlsx`);
//   };

//   return (
//     <Box sx={{ p: 2, height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>

//       {/* Header & Filters */}
//       <Box sx={{ mb: 2, flexShrink: 0, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
//         <Box sx={{ color: 'primary.main', fontWeight: 'bold', fontSize: '2rem' }}>
//           รายงานรายละเอียด
//         </Box>

//         {/* Filter UI */}
//         <Stack direction="row" spacing={2} sx={{ minWidth: 300 }}>
//           <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
//             <InputLabel>Entity (บริษัท)</InputLabel>
//             <Select
//               value={selectedEntity}
//               label="Entity (บริษัท)"
//               onChange={(e) => {
//                 setSelectedEntity(e.target.value);
//                 setSelectedBranch(''); // Reset branch when entity changes
//                 setSelectedDepartment(''); // Reset department
//               }}
//             >
//               <MenuItem value=""><em>All Entities</em></MenuItem>
//               {orgStructure.map((org) => (
//                 <MenuItem key={org.entity} value={org.entity}>{org.entity}</MenuItem>
//               ))}
//             </Select>
//           </FormControl>

//           <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
//             <InputLabel>Branch (สาขา)</InputLabel>
//             <Select
//               value={selectedBranch}
//               label="Branch (สาขา)"
//               onChange={(e) => {
//                 setSelectedBranch(e.target.value);
//                 setSelectedDepartment(''); // Reset Department
//               }}
//               disabled={!selectedEntity}
//             >
//               <MenuItem value=""><em>All Branches</em></MenuItem>
//               {availableBranches.map((branch) => (
//                 <MenuItem key={branch.name} value={branch.name}>{branch.name}</MenuItem>
//               ))}
//             </Select>
//           </FormControl>

//           {/* Department Filter (New) */}
//           <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
//             <InputLabel>Department (แผนก)</InputLabel>
//             <Select
//               value={selectedDepartment}
//               label="Department (แผนก)"
//               onChange={(e) => setSelectedDepartment(e.target.value)}
//             >
//               <MenuItem value=""><em>All Departments</em></MenuItem>
//               {availableDepartments.map((dept) => (
//                 <MenuItem key={dept} value={dept}>{dept}</MenuItem>
//               ))}
//             </Select>
//           </FormControl>
//         </Stack>
//       </Box>

//       {/* Main Grid */}
//       <Box sx={{
//         display: 'grid',
//         gridTemplateColumns: { xs: '1fr', md: '280px minmax(0, 1fr)' },
//         gridTemplateRows: { xs: '320px 1fr', md: '1fr' },
//         gap: 2,
//         flexGrow: 1,
//         overflow: 'hidden',
//         height: '100%',
//         minHeight: 0
//       }}>

//         {/* Left Pane */}
//         <FilterPane />

//         {/* Right Pane */}
//         <Box sx={{
//           display: 'flex',
//           flexDirection: 'column',
//           overflowX: 'hidden',
//           overflowY: { xs: 'auto', md: 'hidden' },
//           height: '100%',
//           minWidth: 0,
//           gap: 2
//         }}>
//           {/* Top: Budget Table */}
//           <BudgetTable
//             loading={loadingDetails}
//             data={budgetDetails}
//             selectedCount={selectedLeaves.size}
//             onDownload={handleBudgetExport}
//           />

//           {/* Bottom: Actual Table */}
//           <ActualTable
//             loading={loadingActuals}
//             data={actualDetails}
//             dateFilter={actualDateFilter}
//             onDateFilterChange={setActualDateFilter}
//             page={actualPage}
//             rowsPerPage={actualRowsPerPage}
//             totalCount={actualTotalCount}
//             onPageChange={setActualPage}
//             onRowsPerPageChange={setActualRowsPerPage}
//             onDownload={handleActualExport}
//           />
//         </Box>

//       </Box>
//     </Box>
//   );
// };

// // Main Page Component wraps content in Provider
// const DetailPage = () => {
//   return (
//     <BudgetProvider>
//       <DetailContent />
//     </BudgetProvider>
//   );
// };

// export default DetailPage;

import React, { useState, useEffect, useMemo } from "react";
import {
  Box,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Stack,
} from "@mui/material";
import api from "../utils/api/axiosInstance";
import { BudgetProvider, useBudget } from "../contexts/BudgetContext";
import FilterPane from "../components/Budget/FilterPane";
import BudgetTable from "../components/Budget/BudgetTable";
import ActualTable from "../components/Budget/ActualTable";
import { downloadExcelFile } from "../utils/exportUtils";

// Inner component that consumes the context
const DetailContent = () => {
  const { selectedLeaves, getAllLeafIds } = useBudget();

  // Data Fetching State
  const [budgetDetails, setBudgetDetails] = useState([]);
  const [loadingDetails, setLoadingDetails] = useState(false);
  const [actualDetails, setActualDetails] = useState([]);
  const [loadingActuals, setLoadingActuals] = useState(false);

  // Date Filter State for Actuals
  const [actualDateFilter, setActualDateFilter] = useState({
    startDate: "",
    endDate: "",
  });

  // Pagination State for Actuals
  const [actualPage, setActualPage] = useState(0);
  const [actualRowsPerPage, setActualRowsPerPage] = useState(10);
  const [actualTotalCount, setActualTotalCount] = useState(0);

  // Filters State
  const [selectedEntity, setSelectedEntity] = useState("");
  const [selectedBranch, setSelectedBranch] = useState("");
  const [selectedDepartment, setSelectedDepartment] = useState("");
  const [orgStructure, setOrgStructure] = useState([]);

  // 🌟 เพิ่ม State ตัวกระตุ้นให้ Table รีเฟรชเมื่อมีการเปลี่ยนเดือนจากตะกร้า
  const [forceRefresh, setForceRefresh] = useState(0);

  // Fetch Filter Options
  useEffect(() => {
    const fetchFilters = async () => {
      try {
        const res = await api.get("/budgets/organization-structure");
        setOrgStructure(res.data || []);
      } catch (err) {
        console.error("Filter Fetch Error", err);
      }
    };
    fetchFilters();
  }, []);

  // Derived state for branches
  const availableBranches = useMemo(() => {
    if (!selectedEntity) return [];
    const entityObj = orgStructure.find((o) => o.entity === selectedEntity);
    return entityObj ? entityObj.branches : [];
  }, [selectedEntity, orgStructure]);

  // Derived state for departments
  const availableDepartments = useMemo(() => {
    if (selectedBranch) {
      const branchObj = availableBranches.find(
        (b) => b.name === selectedBranch
      );
      return branchObj ? branchObj.departments : [];
    } else {
      const depts = new Set();
      orgStructure.forEach((entity) => {
        entity.branches?.forEach((branch) => {
          branch.departments?.forEach((dept) => {
            depts.add(dept);
          });
        });
      });
      return Array.from(depts).sort();
    }
  }, [selectedBranch, availableBranches, orgStructure]);

  // Selection States
  const [syncConfig, setSyncConfig] = useState({
    actualYear: new Date().getFullYear(),
    selectedMonths: [],
    selectedBudget: "",
    selectedCapexBg: "",
    selectedCapexActual: "",
  });

  // Auto Fetch Details when Selection Changes or Date Filter Changes
  useEffect(() => {
    let isMounted = true;

    const fetchDetails = async () => {
      const idsToFetch =
        selectedLeaves.size > 0
          ? Array.from(selectedLeaves)
              .map((id) => id.split("|")[0])
              .filter((code) => code !== "")
          : [];

      if (isMounted) {
        setLoadingDetails(true);
        setLoadingActuals(true);
      }

      try {
        let currentSync = {};
        try {
          const configRes = await api.get("/budgets/configs");
          const configs = configRes.data || {};

          // const parseMonths = (val) => {
          //   if (Array.isArray(val)) return val;
          //   if (typeof val === "string") {
          //     try {
          //       return JSON.parse(val);
          //     } catch (e) {
          //       return [];
          //     }
          //   }
          //   return [];
          // };

          const parseMonths = (val) => {
            let months = [];
            if (Array.isArray(val)) months = val;
            else if (typeof val === "string") {
              try {
                const parsed = JSON.parse(val);
                months = Array.isArray(parsed) ? parsed : [parsed];
              } catch (e) {
                // ถ้าไม่ใช่ JSON string แต่เป็น string ธรรมดา เช่น "04"
                if (val) months = [val];
              }
            }
            // 🌟 แปลงทุกอย่างให้เป็น String 2 หลักเสมอ (เช่น "1" -> "01")
            return months.map((m) => String(m).padStart(2, "0"));
          };

          currentSync = {
            actualYear:
              configs.actualYear ||
              configs.actual_year ||
              syncConfig.actualYear ||
              new Date().getFullYear(),
            selectedMonths:
              parseMonths(configs.selectedMonths || configs.selected_months) ||
              syncConfig.selectedMonths ||
              [],
            selectedBudget:
              configs.selectedBudget ||
              configs.selected_budget ||
              syncConfig.selectedBudget ||
              "",
            selectedCapexBg:
              configs.selectedCapexBg ||
              configs.selected_capex_bg ||
              syncConfig.selectedCapexBg ||
              "",
            selectedCapexActual:
              configs.selectedCapexActual ||
              configs.selected_capex_actual ||
              syncConfig.selectedCapexActual ||
              "",
          };
          if (isMounted) setSyncConfig(currentSync);
        } catch (e) {
          console.error("Failed to fetch fallback configs in Detail", e);
          const cached = JSON.parse(
            localStorage.getItem("dm_lastSyncedConfig") || "{}"
          );
          if (cached.actualYear) {
            currentSync = { ...currentSync, ...cached };
            if (isMounted) setSyncConfig(currentSync);
          }
        }

        const payload = {
          conso_gls: idsToFetch,
          start_date: actualDateFilter.startDate,
          end_date: actualDateFilter.endDate,
          entities: selectedEntity ? [selectedEntity] : [],
          branches: selectedBranch ? [selectedBranch] : [],
          departments: selectedDepartment ? [selectedDepartment] : [],
          year: String(currentSync.actualYear),
          months: currentSync.selectedMonths,
          budget_file_id: currentSync.selectedBudget,
          capex_file_id: currentSync.selectedCapexBg,
          capex_actual_file_id: currentSync.selectedCapexActual,
          page: actualPage,
          limit: actualRowsPerPage,
        };

        // Fetch Budget
        api
          .post("/budgets/details", payload)
          .then((res) => {
            if (!isMounted) return;
            const rawBudget = res.data || [];
            const budgetMap = new Map();
            rawBudget.forEach((item) => {
              const key = `${item.conso_gl}|${item.gl_name}`;
              if (!budgetMap.has(key)) {
                budgetMap.set(key, {
                  ...item,
                  budget_amounts: [...(item.budget_amounts || [])],
                });
              } else {
                const existing = budgetMap.get(key);
                existing.year_total =
                  (parseFloat(existing.year_total) || 0) +
                  (parseFloat(item.year_total) || 0);
                const existingAmounts = existing.budget_amounts;
                item.budget_amounts?.forEach((newAmt) => {
                  const match = existingAmounts.find(
                    (ea) => ea.month === newAmt.month
                  );
                  if (match)
                    match.amount =
                      (parseFloat(match.amount) || 0) +
                      (parseFloat(newAmt.amount) || 0);
                  else existingAmounts.push({ ...newAmt });
                });
              }
            });
            setBudgetDetails(Array.from(budgetMap.values()));
          })
          .catch((err) => {
            console.error("Budget Details Fetch Failed", err);
            if (isMounted) setBudgetDetails([]);
          })
          .finally(() => {
            if (isMounted) setLoadingDetails(false);
          });

        // Fetch Actuals
        api
          .post("/budgets/actuals-transactions", payload)
          .then((res) => {
            if (!isMounted) return;
            const resMap = res.data || {};
            setActualDetails(resMap.data || []);
            setActualTotalCount(resMap.total_count || 0);
          })
          .catch((err) => {
            console.error("Actual Transactions Fetch Failed", err);
            if (isMounted) {
              setActualDetails([]);
              setActualTotalCount(0);
            }
          })
          .finally(() => {
            if (isMounted) setLoadingActuals(false);
          });
      } catch (err) {
        console.error("Fetch Setup Error", err);
        if (isMounted) {
          setLoadingDetails(false);
          setLoadingActuals(false);
        }
      }
    };

    const timeoutId = setTimeout(() => {
      fetchDetails();
    }, 300);

    return () => {
      isMounted = false;
      clearTimeout(timeoutId);
    };
  }, [
    selectedLeaves,
    getAllLeafIds,
    actualDateFilter,
    selectedEntity,
    selectedBranch,
    selectedDepartment,
    actualPage,
    actualRowsPerPage,
    forceRefresh,
  ]);
  // 🌟 อย่าลืมเติม forceRefresh ใน Array ด้านบนด้วย!

  // ==========================================
  // 🌟 ระบบจัดการ Filters เพื่อรองรับ Auto-Sync จากตะกร้า
  // ==========================================
  // const currentFilters = useMemo(() => ({
  //   entity: selectedEntity,
  //   branch: selectedBranch,
  //   department: selectedDepartment,
  //   year: String(syncConfig.actualYear),
  //   month: syncConfig.selectedMonths?.[0] || '',
  //   months: syncConfig.selectedMonths || []
  // }), [selectedEntity, selectedBranch, selectedDepartment, syncConfig]);

  const currentFilters = useMemo(() => {
    // หาเดือนที่จะโชว์: 1. จาก config 2. ถ้าไม่มีใช้เดือนปัจจุบัน
    const currentMonthFromConfig =
      syncConfig.selectedMonths && syncConfig.selectedMonths.length > 0
        ? syncConfig.selectedMonths[0]
        : String(new Date().getMonth() + 1).padStart(2, "0");

    return {
      entity: selectedEntity,
      branch: selectedBranch,
      department: selectedDepartment,
      year: String(syncConfig.actualYear),
      month: currentMonthFromConfig, // 👈 มั่นใจว่ามีค่าเสมอ
      months: syncConfig.selectedMonths || [currentMonthFromConfig],
    };
  }, [selectedEntity, selectedBranch, selectedDepartment, syncConfig]);

  const handleSetFilters = async (updater) => {
    // จำลองการทำงานแบบ setState(prev => next)
    const nextFilters =
      typeof updater === "function" ? updater(currentFilters) : updater;

    // อัปเดต State พื้นฐาน
    if (
      nextFilters.entity !== undefined &&
      nextFilters.entity !== selectedEntity
    )
      setSelectedEntity(nextFilters.entity);
    if (
      nextFilters.branch !== undefined &&
      nextFilters.branch !== selectedBranch
    )
      setSelectedBranch(nextFilters.branch);
    if (
      nextFilters.department !== undefined &&
      nextFilters.department !== selectedDepartment
    )
      setSelectedDepartment(nextFilters.department);

    // ถ้ามีการบังคับเปลี่ยนปี/เดือนจากตะกร้า (Auto-Sync)
    const yearChanged =
      nextFilters.year &&
      String(nextFilters.year) !== String(syncConfig.actualYear);
    const monthChanged =
      nextFilters.months &&
      JSON.stringify(nextFilters.months) !==
        JSON.stringify(syncConfig.selectedMonths);

    if (yearChanged || monthChanged) {
      const newYear = nextFilters.year
        ? parseInt(nextFilters.year, 10)
        : syncConfig.actualYear;
      const newMonths = nextFilters.months || syncConfig.selectedMonths;

      // 1. อัปเดต local state เพื่อให้ Component อื่น (เช่น FilterPane) รู้ตัว
      setSyncConfig((prev) => ({
        ...prev,
        actualYear: newYear,
        selectedMonths: newMonths,
      }));

      // 2. ยิง API ไปบันทึก Global Configs (เพื่อให้ useEffect โหลดข้อมูลได้ถูกเดือน)
      try {
        const configRes = await api.get("/budgets/configs");
        const currentConfigs = configRes.data || {};
        await api.post("/budgets/configs", {
          ...currentConfigs,
          actual_year: newYear,
          selected_months: newMonths,
        });

        // 3. กระตุ้นให้ useEffect ทำงานดึงข้อมูลใหม่ (Data Refresh)
        setForceRefresh((prev) => prev + 1);
      } catch (err) {
        console.error("Failed to auto-sync config with basket", err);
      }
    }
  };
  // ==========================================

  const handleBudgetExport = async () => {
    const idsToFetch =
      selectedLeaves.size > 0
        ? Array.from(selectedLeaves)
            .map((id) => id.split("|")[0])
            .filter((code) => code !== "")
        : [];

    let syncConfig = JSON.parse(
      localStorage.getItem("dm_lastSyncedConfig") || "{}"
    );
    const actualYear = syncConfig.actualYear || new Date().getFullYear();
    const selectedMonths = Array.isArray(syncConfig.selectedMonths)
      ? syncConfig.selectedMonths
      : [];

    const payload = {
      conso_gls: idsToFetch,
      entities: selectedEntity ? [selectedEntity] : [],
      branches: selectedBranch ? [selectedBranch] : [],
      departments: selectedDepartment ? [selectedDepartment] : [],
      year: String(actualYear),
      months: selectedMonths,
      budget_file_id: syncConfig.selectedBudget,
      capex_file_id: syncConfig.selectedCapexBg,
      capex_actual_file_id: syncConfig.selectedCapexActual,
      // No pagination for export
    };

    await downloadExcelFile(
      "/export-budget-detail",
      payload,
      `Budget_Detail_Report_${actualYear}.xlsx`
    );
  };

  const handleActualExport = async () => {
    const idsToFetch =
      selectedLeaves.size > 0
        ? Array.from(selectedLeaves)
            .map((id) => id.split("|")[0])
            .filter((code) => code !== "")
        : [];

    let syncConfig = JSON.parse(
      localStorage.getItem("dm_lastSyncedConfig") || "{}"
    );
    const actualYear = syncConfig.actualYear || new Date().getFullYear();
    const selectedMonths = Array.isArray(syncConfig.selectedMonths)
      ? syncConfig.selectedMonths
      : [];

    const payload = {
      conso_gls: idsToFetch,
      start_date: actualDateFilter.startDate,
      end_date: actualDateFilter.endDate,
      entities: selectedEntity ? [selectedEntity] : [],
      branches: selectedBranch ? [selectedBranch] : [],
      departments: selectedDepartment ? [selectedDepartment] : [],
      year: String(actualYear),
      months: selectedMonths,
      budget_file_id: syncConfig.selectedBudget,
      capex_file_id: syncConfig.selectedCapexBg,
      capex_actual_file_id: syncConfig.selectedCapexActual,
      // No pagination for export
    };

    await downloadExcelFile(
      "/export-actual-detail",
      payload,
      `Actual_Detail_Report_${actualYear}.xlsx`
    );
  };

  return (
    <Box
      sx={{
        p: 2,
        height: "100vh",
        display: "flex",
        flexDirection: "column",
        overflow: "hidden",
      }}
    >
      {/* Header & Filters */}
      <Box
        sx={{
          mb: 2,
          flexShrink: 0,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <Box
          sx={{ color: "primary.main", fontWeight: "bold", fontSize: "2rem" }}
        >
          รายงานรายละเอียด
        </Box>

        {/* Filter UI */}
        <Stack direction="row" spacing={2} sx={{ minWidth: 300 }}>
          <FormControl
            size="small"
            sx={{ minWidth: 200, bgcolor: "white", borderRadius: 1 }}
          >
            <InputLabel>Entity (บริษัท)</InputLabel>
            <Select
              value={selectedEntity}
              label="Entity (บริษัท)"
              onChange={(e) => {
                setSelectedEntity(e.target.value);
                setSelectedBranch("");
                setSelectedDepartment("");
              }}
            >
              <MenuItem value="">
                <em>All Entities</em>
              </MenuItem>
              {orgStructure.map((org) => (
                <MenuItem key={org.entity} value={org.entity}>
                  {org.entity}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          <FormControl
            size="small"
            sx={{ minWidth: 200, bgcolor: "white", borderRadius: 1 }}
          >
            <InputLabel>Branch (สาขา)</InputLabel>
            <Select
              value={selectedBranch}
              label="Branch (สาขา)"
              onChange={(e) => {
                setSelectedBranch(e.target.value);
                setSelectedDepartment("");
              }}
              disabled={!selectedEntity}
            >
              <MenuItem value="">
                <em>All Branches</em>
              </MenuItem>
              {availableBranches.map((branch) => (
                <MenuItem key={branch.name} value={branch.name}>
                  {branch.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          <FormControl
            size="small"
            sx={{ minWidth: 200, bgcolor: "white", borderRadius: 1 }}
          >
            <InputLabel>Department (แผนก)</InputLabel>
            <Select
              value={selectedDepartment}
              label="Department (แผนก)"
              onChange={(e) => setSelectedDepartment(e.target.value)}
            >
              <MenuItem value="">
                <em>All Departments</em>
              </MenuItem>
              {availableDepartments.map((dept) => (
                <MenuItem key={dept} value={dept}>
                  {dept}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Stack>
      </Box>

      {/* Main Grid */}
      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", md: "280px minmax(0, 1fr)" },
          gridTemplateRows: { xs: "320px 1fr", md: "1fr" },
          gap: 2,
          flexGrow: 1,
          overflow: "hidden",
          height: "100%",
          minHeight: 0,
        }}
      >
        {/* Left Pane */}
        <FilterPane />

        {/* Right Pane */}
        <Box
          sx={{
            display: "flex",
            flexDirection: "column",
            overflowX: "hidden",
            overflowY: { xs: "auto", md: "hidden" },
            height: "100%",
            minWidth: 0,
            gap: 2,
          }}
        >
          {/* Top: Budget Table */}
          <BudgetTable
            loading={loadingDetails}
            data={budgetDetails}
            selectedCount={selectedLeaves.size}
            onDownload={handleBudgetExport}
          />

          {/* Bottom: Actual Table */}
          <ActualTable
            loading={loadingActuals}
            data={actualDetails}
            dateFilter={actualDateFilter}
            onDateFilterChange={setActualDateFilter}
            page={actualPage}
            rowsPerPage={actualRowsPerPage}
            totalCount={actualTotalCount}
            onPageChange={setActualPage}
            onRowsPerPageChange={setActualRowsPerPage}
            onDownload={handleActualExport}
            // 🌟 พระเอกของเรา: โยน filters และ setFilters ลงไปให้ตารางใช้งาน
            filters={currentFilters}
            setFilters={handleSetFilters}
          />
        </Box>
      </Box>
    </Box>
  );
};

const DetailPage = () => {
  return (
    <BudgetProvider>
      <DetailContent />
    </BudgetProvider>
  );
};

export default DetailPage;
