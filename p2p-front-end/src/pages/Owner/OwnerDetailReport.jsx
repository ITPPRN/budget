import React, { useState, useEffect, useMemo } from 'react';
import { Box, FormControl, InputLabel, Select, MenuItem, Stack, Typography } from '@mui/material';
import api from '../../utils/api/axiosInstance';
import { BudgetProvider, useBudget } from '../../contexts/BudgetContext';
import FilterPane from '../../components/Budget/FilterPane';
import BudgetTable from '../../components/Budget/BudgetTable';
import ActualTable from '../../components/Budget/ActualTable';
import { useAuth } from '../../hooks/useAuth';
import { downloadExcelFile } from '../../utils/exportUtils';

// Inner component for Owner Report
const OwnerDetailContent = () => {
    const { selectedLeaves, getAllLeafIds } = useBudget();
    const { user } = useAuth();

    // Determine the user's constraints based on their profile data
    const userEntity = user?.company || '';
    const userBranch = user?.branch || '';
    const userDepartment = user?.department || user?.department_code || '';

    // Data Fetching State
    const [budgetDetails, setBudgetDetails] = useState([]);
    const [loadingDetails, setLoadingDetails] = useState(false);
    const [actualDetails, setActualDetails] = useState([]);
    const [loadingActuals, setLoadingActuals] = useState(false);

    // Date Filter State for Actuals
    const [actualDateFilter, setActualDateFilter] = useState({ startDate: '', endDate: '' });

    // Pagination State for Actuals
    const [actualPage, setActualPage] = useState(0);
    const [actualRowsPerPage, setActualRowsPerPage] = useState(10);
    const [actualTotalCount, setActualTotalCount] = useState(0);

    // Filters State (Initialized to 'All' by default)
    const [selectedEntity, setSelectedEntity] = useState('');
    const [selectedBranch, setSelectedBranch] = useState('');
    const [selectedDepartment, setSelectedDepartment] = useState('');
    const [selectedActualYear, setSelectedActualYear] = useState('');
    const [orgStructure, setOrgStructure] = useState([]);
    const [actualYears, setActualYears] = useState([]);

    // Sync Config State
    const [syncConfig, setSyncConfig] = useState({
        actualYear: new Date().getFullYear(),
        selectedMonths: [],
        selectedBudget: "",
        selectedCapexBg: "",
        selectedCapexActual: ""
    });

    // Helper: derive month filter override from actualDateFilter
    const MONTH_ABBR = ['JAN','FEB','MAR','APR','MAY','JUN','JUL','AUG','SEP','OCT','NOV','DEC'];
    const getMonthOverride = () => {
        if (!actualDateFilter.startDate) return null;
        const parts = actualDateFilter.startDate.split('-'); // "2026-04-01"
        const monthIdx = parseInt(parts[1], 10) - 1;
        const year = parseInt(parts[0], 10);
        const lastDay = new Date(year, monthIdx + 1, 0).getDate();
        return {
            months: [MONTH_ABBR[monthIdx]],
            start_date: `${parts[0]}-${parts[1]}-01`,
            end_date: `${parts[0]}-${parts[1]}-${String(lastDay).padStart(2, '0')}`
        };
    };



    // Fetch Filter Options
    useEffect(() => {
        const fetchFilters = async () => {
            try {
                const [orgRes, yearRes] = await Promise.all([
                    api.get('/owner/organization-structure'),
                    api.get('/owner/actual-years')
                ]);
                setOrgStructure(orgRes.data || []);
                setActualYears(yearRes.data || []);

            } catch (err) {
                console.error("Filter Fetch Error", err);
            }
        };
        fetchFilters();
    }, []);

    // Derived state for branches
    const availableBranches = useMemo(() => {
        if (!selectedEntity) return [];
        const entityObj = orgStructure.find(o => o.entity === selectedEntity);
        return entityObj ? entityObj.branches : [];
    }, [selectedEntity, orgStructure]);

    // Derived state for departments based on branch selection, otherwise ALL departments
    const availableDepartments = useMemo(() => {
        if (selectedBranch) {
            const branchObj = availableBranches.find(b => b.name === selectedBranch);
            return branchObj ? branchObj.departments : [];
        } else {
            // Flatten all unique departments across entire org
            const depts = new Set();
            orgStructure.forEach(entity => {
                entity.branches?.forEach(branch => {
                    branch.departments?.forEach(dept => {
                        depts.add(dept);
                    });
                });
            });
            return Array.from(depts).sort();
        }
    }, [selectedBranch, availableBranches, orgStructure]);

    // Auto Fetch Details when Selection Changes or Date Filter Changes
    useEffect(() => {
        let isMounted = true;

        const fetchDetails = async () => {
            const idsToFetch = selectedLeaves.size > 0
                ? Array.from(selectedLeaves).map(id => id.split('|')[0]).filter(code => code !== "")
                : []; // Optimized: Empty list means 'All' to the backend

            if (isMounted) {
                setLoadingDetails(true);
                setLoadingActuals(true);
            }

            try {
                // Robust Fetch: Always fetch from Backend to guarantee true Global synchronization
                let activeSync = { ...syncConfig };

                try {
                    const configRes = await api.get('/budgets/configs');
                    const configs = configRes.data || {};
                    
                    const parseMonths = (val) => {
                        if (Array.isArray(val)) return val;
                        if (typeof val === 'string') {
                            try { return JSON.parse(val); } catch (e) { return []; }
                        }
                        return [];
                    };

                    activeSync = {
                        actualYear: configs.actualYear || configs.actual_year || syncConfig.actualYear || new Date().getFullYear(),
                        selectedMonths: parseMonths(configs.selectedMonths || configs.selected_months) || syncConfig.selectedMonths || [],
                        selectedBudget: configs.selectedBudget || configs.selected_budget || syncConfig.selectedBudget || "",
                        selectedCapexBg: configs.selectedCapexBg || configs.selected_capex_bg || syncConfig.selectedCapexBg || "",
                        selectedCapexActual: configs.selectedCapexActual || configs.selected_capex_actual || syncConfig.selectedCapexActual || ""
                    };
                    if (isMounted) setSyncConfig(activeSync);
                } catch (e) {
                    console.error("Failed to fetch fallback configs in OwnerDetail", e);
                    const cached = JSON.parse(localStorage.getItem('dm_lastSyncedConfig') || '{}');
                    if (cached.actualYear) {
                        activeSync = { ...activeSync, ...cached };
                        if (isMounted) setSyncConfig(activeSync);
                    }
                }

                const actualYear = activeSync.actualYear || new Date().getFullYear();
                const monthOvr = getMonthOverride();

                const payload = {
                    conso_gls: idsToFetch,
                    start_date: monthOvr ? monthOvr.start_date : actualDateFilter.startDate,
                    end_date: monthOvr ? monthOvr.end_date : actualDateFilter.endDate,
                    entities: selectedEntity ? [selectedEntity] : [],
                    branches: selectedBranch ? [selectedBranch] : [],
                    departments: selectedDepartment ? [selectedDepartment] : availableDepartments,
                    year: String(actualYear),
                    months: monthOvr ? monthOvr.months : (Array.isArray(activeSync.selectedMonths) ? activeSync.selectedMonths : []),
                    page: actualPage + 1,
                    limit: actualRowsPerPage
                };

                // Fetch Budget (Fast)
                api.post('/budgets/details', payload)
                    .then(res => {
                        if (!isMounted) return;
                        setBudgetDetails(res.data || []);
                    })
                    .catch(err => {
                        console.error("Budget Details Fetch Failed", err);
                        if (isMounted) setBudgetDetails([]);
                    })
                    .finally(() => {
                        if (isMounted) setLoadingDetails(false);
                    });

                // Fetch Actuals (May be slower)
                api.post('/owner/actual-transactions', payload)
                    .then(res => {
                        if (!isMounted) return;
                        const resMap = res.data || {};
                        setActualDetails(resMap.data || []);
                        setActualTotalCount(resMap.total_count || 0);
                    })
                    .catch(err => {
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
    }, [selectedLeaves, getAllLeafIds, actualDateFilter, selectedEntity, selectedBranch, selectedDepartment, actualPage, actualRowsPerPage, selectedActualYear]);

    // --- Export Handlers ---
    const handleBudgetExport = async () => {
        const idsToFetch = selectedLeaves.size > 0
            ? Array.from(selectedLeaves).map(id => id.split('|')[0]).filter(code => code !== "")
            : [];
            
        let syncConfig = JSON.parse(localStorage.getItem('dm_lastSyncedConfig') || '{}');
        const actualYear = syncConfig.actualYear || new Date().getFullYear();

        const payload = {
            conso_gls: idsToFetch,
            entities: selectedEntity ? [selectedEntity] : [],
            branches: selectedBranch ? [selectedBranch] : [],
            departments: selectedDepartment ? [selectedDepartment] : availableDepartments,
            year: String(actualYear),
            months: Array.isArray(syncConfig.selectedMonths) ? syncConfig.selectedMonths : [],
            budget_file_id: syncConfig.selectedBudget,
            capex_file_id: syncConfig.selectedCapexBg,
            capex_actual_file_id: syncConfig.selectedCapexActual,
        };

        await downloadExcelFile('/owner/export-budget-detail', payload, `Owner_Budget_Detail_Report_${actualYear}.xlsx`);
    };

    const handleActualExport = async () => {
        const idsToFetch = selectedLeaves.size > 0
            ? Array.from(selectedLeaves).map(id => id.split('|')[0]).filter(code => code !== "")
            : [];

        let syncConfig = JSON.parse(localStorage.getItem('dm_lastSyncedConfig') || '{}');
        const actualYear = syncConfig.actualYear || new Date().getFullYear();
        const monthOvr = getMonthOverride();

        const payload = {
            conso_gls: idsToFetch,
            start_date: monthOvr ? monthOvr.start_date : actualDateFilter.startDate,
            end_date: monthOvr ? monthOvr.end_date : actualDateFilter.endDate,
            entities: selectedEntity ? [selectedEntity] : [],
            branches: selectedBranch ? [selectedBranch] : [],
            departments: selectedDepartment ? [selectedDepartment] : availableDepartments,
            year: String(actualYear),
            months: monthOvr ? monthOvr.months : (Array.isArray(syncConfig.selectedMonths) ? syncConfig.selectedMonths : []),
            budget_file_id: syncConfig.selectedBudget,
            capex_file_id: syncConfig.selectedCapexBg,
            capex_actual_file_id: syncConfig.selectedCapexActual,
        };

        await downloadExcelFile('/owner/export-actual-detail', payload, `Owner_Actual_Detail_Report_${actualYear}.xlsx`);
    };

    if (!user) {
        return <Box sx={{ p: 3, display: 'flex', justifyContent: 'center' }}><Typography>Loading user data...</Typography></Box>;
    }

    return (
        <Box sx={{ p: 2, height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
            {/* Header & Filters */}
            <Box sx={{ mb: 2, flexShrink: 0, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Box sx={{ color: 'primary.main', fontWeight: 'bold', fontSize: '1.5rem' }}>
                    รายงานรายละเอียด
                </Box>

                {/* Filter UI */}
                <Stack direction="row" spacing={2}>
                    {/* Independent Year Filter Removed as per User Request (Use Global Database Actuals) */}

                    {/* Entity Filter (Primary) */}
                    <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
                        <InputLabel id="entity-label">Entity (บริษัท)</InputLabel>
                        <Select
                            labelId="entity-label"
                            value={selectedEntity}
                            label="Entity (บริษัท)"
                            onChange={(e) => {
                                setSelectedEntity(e.target.value);
                                setSelectedBranch('');
                            }}
                            disabled={!!userEntity} // Locked if user has entity
                        >
                            <MenuItem value="">All Entities</MenuItem>
                            {orgStructure.map((org) => (
                                <MenuItem key={org.entity} value={org.entity}>{org.entity}</MenuItem>
                            ))}
                        </Select>
                    </FormControl>

                    {/* Branch Filter (Secondary) */}
                    <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
                        <InputLabel id="branch-label">Branch (สาขา)</InputLabel>
                        <Select
                            labelId="branch-label"
                            value={selectedBranch}
                            label="Branch (สาขา)"
                            onChange={(e) => {
                                setSelectedBranch(e.target.value);
                            }}
                            disabled={!!userBranch || (!selectedEntity && !userEntity)} // Locked if user has branch
                        >
                            <MenuItem value="">All Branches</MenuItem>
                            {availableBranches.map((branch) => (
                                <MenuItem key={branch.name} value={branch.name}>{branch.name}</MenuItem>
                            ))}
                        </Select>
                    </FormControl>

                    {/* Department Filter (Tertiary) */}
                    <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
                        <InputLabel id="dept-label">Department (แผนก)</InputLabel>
                        <Select
                            labelId="dept-label"
                            value={selectedDepartment}
                            label="Department (แผนก)"
                            onChange={(e) => setSelectedDepartment(e.target.value)}
                        >
                            <MenuItem value="">All Departments</MenuItem>
                            {(() => {
                                // 🛠️ Simplified: Use only availableDepartments (already filtered by Backend permissions)
                                const sortedDepts = Array.from(new Set(availableDepartments)).sort();
                                return sortedDepts.map((dept) => (
                                    <MenuItem key={dept} value={dept}>{dept}</MenuItem>
                                ));
                            })()}
                        </Select>
                    </FormControl>
                </Stack>
            </Box>

            {/* Main Grid */}
            <Box sx={{
                display: 'grid',
                gridTemplateColumns: { xs: '1fr', md: '280px minmax(0, 1fr)' },
                gridTemplateRows: { xs: '320px 1fr', md: '1fr' },
                gap: 2,
                flexGrow: 1,
                overflow: 'hidden',
                height: '100%',
                minHeight: 0
            }}>
                {/* Left Pane */}
                <FilterPane />

                {/* Right Pane */}
                <Box sx={{
                    display: 'flex',
                    flexDirection: 'column',
                    overflowX: 'hidden',
                    overflowY: { xs: 'auto', md: 'hidden' },
                    height: '100%',
                    minWidth: 0,
                    gap: 2
                }}>
                    <BudgetTable
                        loading={loadingDetails}
                        data={budgetDetails}
                        selectedCount={selectedLeaves.size}
                        onDownload={handleBudgetExport}
                    />
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
                        filters={{
                            entity: selectedEntity,
                            branch: selectedBranch,
                            department: selectedDepartment,
                            year: String(syncConfig.actualYear || ""),
                            month: (syncConfig.selectedMonths && syncConfig.selectedMonths[0]) || "",
                            months: syncConfig.selectedMonths,
                            conso_gls: selectedLeaves.size > 0
                                ? Array.from(selectedLeaves).map(id => id.split('|')[0]).filter(code => code !== "")
                                : [],
                            start_date: actualDateFilter.startDate,
                            end_date: actualDateFilter.endDate
                        }}
                    />
                </Box>
            </Box>
        </Box>
    );
};

// Main Page Component
const OwnerDetailReport = () => {
    return (
        <BudgetProvider>
            <OwnerDetailContent />
        </BudgetProvider>
    );
};

export default OwnerDetailReport;
