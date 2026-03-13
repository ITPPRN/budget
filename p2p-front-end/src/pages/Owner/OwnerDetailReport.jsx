import React, { useState, useEffect, useMemo } from 'react';
import { Box, FormControl, InputLabel, Select, MenuItem, Stack, Typography } from '@mui/material';
import api from '../../utils/api/axiosInstance';
import { BudgetProvider, useBudget } from '../../contexts/BudgetContext';
import FilterPane from '../../components/Budget/FilterPane';
import BudgetTable from '../../components/Budget/BudgetTable';
import ActualTable from '../../components/Budget/ActualTable';
import { useAuth } from '../../hooks/useAuth';

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
                const payload = {
                    conso_gls: idsToFetch,
                    start_date: actualDateFilter.startDate,
                    end_date: actualDateFilter.endDate,
                    entities: selectedEntity ? [selectedEntity] : [],
                    branches: selectedBranch ? [selectedBranch] : [],
                    departments: selectedDepartment ? [selectedDepartment] : [],
                    year: selectedActualYear, // Use the independent year filter
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
                <Stack direction="row" spacing={2} sx={{ minWidth: 800 }}>
                    {/* Independent Year Filter (Actual Only) */}
                    <FormControl size="small" sx={{ minWidth: 120, bgcolor: 'white', borderRadius: 1 }}>
                        <InputLabel id="actual-year-label" shrink={!!selectedActualYear}>Year (ปี)</InputLabel>
                        <Select
                            labelId="actual-year-label"
                            value={selectedActualYear}
                            label="Year (ปี)"
                            onChange={(e) => setSelectedActualYear(e.target.value)}
                        >
                            <MenuItem value="">FY All</MenuItem>
                            {actualYears.map((year) => (
                                <MenuItem key={year} value={year}>{`FY ${year}`}</MenuItem>
                            ))}
                        </Select>
                    </FormControl>

                    {/* Entity Filter (Primary) */}
                    <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
                        <InputLabel id="entity-label" shrink={!!selectedEntity}>Entity (บริษัท)</InputLabel>
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
                        <InputLabel id="branch-label" shrink={!!selectedBranch}>Branch (สาขา)</InputLabel>
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
                        <InputLabel id="dept-label" shrink={!!selectedDepartment}>Department (แผนก)</InputLabel>
                        <Select
                            labelId="dept-label"
                            value={selectedDepartment}
                            label="Department (แผนก)"
                            onChange={(e) => setSelectedDepartment(e.target.value)}
                        >
                            <MenuItem value="">All Departments</MenuItem>
                            {(() => {
                                // 🛠️ Robust Logic: Merge Depts from OrgStructure with User's Allowed Permissions
                                const deptsSet = new Set(availableDepartments);
                                user?.permissions?.forEach(p => {
                                    if (p.is_active && p.department_code) deptsSet.add(p.department_code.trim());
                                });
                                if (userDepartment) deptsSet.add(userDepartment.trim());

                                const sortedDepts = Array.from(deptsSet).sort();
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
                    />
                    <ActualTable
                        loading={loadingActuals}
                        data={actualDetails}
                        dateFilter={actualDateFilter}
                        onDateFilterChange={setActualDateFilter}
                        yearFilter={''}
                        onYearFilterChange={() => {}}
                        years={[]}
                        page={actualPage}
                        rowsPerPage={actualRowsPerPage}
                        totalCount={actualTotalCount}
                        onPageChange={setActualPage}
                        onRowsPerPageChange={setActualRowsPerPage}
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
