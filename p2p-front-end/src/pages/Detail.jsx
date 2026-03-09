import React, { useState, useEffect, useMemo } from 'react';
import { Box, FormControl, InputLabel, Select, MenuItem, Stack } from '@mui/material';
import api from '../utils/api/axiosInstance';
import { BudgetProvider, useBudget } from '../contexts/BudgetContext';
import FilterPane from '../components/Budget/FilterPane';
import BudgetTable from '../components/Budget/BudgetTable';
import ActualTable from '../components/Budget/ActualTable';

// Inner component that consumes the context
const DetailContent = () => {
  const { selectedLeaves, getAllLeafIds } = useBudget();

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

  // Filters State
  const [selectedEntity, setSelectedEntity] = useState('');
  const [selectedBranch, setSelectedBranch] = useState('');
  const [selectedDepartment, setSelectedDepartment] = useState(''); // New State
  const [orgStructure, setOrgStructure] = useState([]);

  // Fetch Filter Options
  useEffect(() => {
    const fetchFilters = async () => {
      try {
        const res = await api.get('/budgets/organization-structure');
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
    const entityObj = orgStructure.find(o => o.entity === selectedEntity);
    return entityObj ? entityObj.branches : [];
  }, [selectedEntity, orgStructure]);

  // Derived state for departments
  const availableDepartments = useMemo(() => {
    if (selectedBranch) {
      const branchObj = availableBranches.find(b => b.name === selectedBranch);
      return branchObj ? branchObj.departments : [];
    } else {
      // Flatten all unique departments across entire org if no branch selected
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
      // Logic: If selection is empty -> Fetch ALL (send empty list to backend to optimize)
      const idsToFetch = selectedLeaves.size > 0
        ? Array.from(selectedLeaves).map(id => id.split('|')[0]).filter(code => code !== "")
        : []; // Optimizing: Don't send 200+ IDs if not specifically filtered

      if (isMounted) {
        setLoadingDetails(true);
        setLoadingActuals(true);
      }

      try {
        // Get the synced Actuals configuration from localStorage
        const syncConfig = JSON.parse(localStorage.getItem('dm_lastSyncedConfig') || '{}');
        const actualYear = syncConfig.actualYear || new Date().getFullYear();
        const selectedMonths = syncConfig.selectedMonths || [];

        const payload = {
          conso_gls: idsToFetch,
          start_date: actualDateFilter.startDate,
          end_date: actualDateFilter.endDate,
          entities: selectedEntity ? [selectedEntity] : [],
          branches: selectedBranch ? [selectedBranch] : [],
          departments: selectedDepartment ? [selectedDepartment] : [],
          year: String(actualYear),
          months: selectedMonths,
          page: actualPage,
          limit: actualRowsPerPage
        };

        // Fetch Budget (Fast)
        api.post('/budgets/details', payload)
          .then(res => {
            if (!isMounted) return;
            const rawBudget = res.data || [];
            const budgetMap = new Map();
            rawBudget.forEach(item => {
              const key = `${item.conso_gl}|${item.gl_name}`;
              if (!budgetMap.has(key)) {
                budgetMap.set(key, { ...item, budget_amounts: [...(item.budget_amounts || [])] });
              } else {
                const existing = budgetMap.get(key);
                existing.year_total = (parseFloat(existing.year_total) || 0) + (parseFloat(item.year_total) || 0);
                const existingAmounts = existing.budget_amounts;
                item.budget_amounts?.forEach(newAmt => {
                  const match = existingAmounts.find(ea => ea.month === newAmt.month);
                  if (match) match.amount = (parseFloat(match.amount) || 0) + (parseFloat(newAmt.amount) || 0);
                  else existingAmounts.push({ ...newAmt });
                });
              }
            });
            setBudgetDetails(Array.from(budgetMap.values()));
          })
          .catch(err => {
            console.error("Budget Details Fetch Failed", err);
            if (isMounted) setBudgetDetails([]);
          })
          .finally(() => {
            if (isMounted) setLoadingDetails(false);
          });

        // Fetch Actuals (May be slower due to large DB)
        api.post('/budgets/actuals-transactions', payload)
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

    // Debounce slightly to avoid rapid re-fetches if selection changes fast
    const timeoutId = setTimeout(() => {
      fetchDetails();
    }, 300);

    return () => {
      isMounted = false;
      clearTimeout(timeoutId);
    };
  }, [selectedLeaves, getAllLeafIds, actualDateFilter, selectedEntity, selectedBranch, selectedDepartment, actualPage, actualRowsPerPage]); // Add dependencies

  return (
    <Box sx={{ p: 2, height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>

      {/* Header & Filters */}
      <Box sx={{ mb: 2, flexShrink: 0, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Box sx={{ color: 'primary.main', fontWeight: 'bold', fontSize: '1.5rem' }}>
          รายงานรายละเอียด
        </Box>

        {/* Filter UI */}
        <Stack direction="row" spacing={2} sx={{ minWidth: 300 }}>
          <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
            <InputLabel>Entity (บริษัท)</InputLabel>
            <Select
              value={selectedEntity}
              label="Entity (บริษัท)"
              onChange={(e) => {
                setSelectedEntity(e.target.value);
                setSelectedBranch(''); // Reset branch when entity changes
                setSelectedDepartment(''); // Reset department
              }}
            >
              <MenuItem value=""><em>All Entities</em></MenuItem>
              {orgStructure.map((org) => (
                <MenuItem key={org.entity} value={org.entity}>{org.entity}</MenuItem>
              ))}
            </Select>
          </FormControl>

          <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
            <InputLabel>Branch (สาขา)</InputLabel>
            <Select
              value={selectedBranch}
              label="Branch (สาขา)"
              onChange={(e) => {
                setSelectedBranch(e.target.value);
                setSelectedDepartment(''); // Reset Department
              }}
              disabled={!selectedEntity}
            >
              <MenuItem value=""><em>All Branches</em></MenuItem>
              {availableBranches.map((branch) => (
                <MenuItem key={branch.name} value={branch.name}>{branch.name}</MenuItem>
              ))}
            </Select>
          </FormControl>

          {/* Department Filter (New) */}
          <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
            <InputLabel>Department (แผนก)</InputLabel>
            <Select
              value={selectedDepartment}
              label="Department (แผนก)"
              onChange={(e) => setSelectedDepartment(e.target.value)}
            >
              <MenuItem value=""><em>All Departments</em></MenuItem>
              {availableDepartments.map((dept) => (
                <MenuItem key={dept} value={dept}>{dept}</MenuItem>
              ))}
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
          {/* Top: Budget Table */}
          <BudgetTable
            loading={loadingDetails}
            data={budgetDetails}
            selectedCount={selectedLeaves.size}
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
          />
        </Box>

      </Box>
    </Box>
  );
};

// Main Page Component wraps content in Provider
const DetailPage = () => {
  return (
    <BudgetProvider>
      <DetailContent />
    </BudgetProvider>
  );
};

export default DetailPage;