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

  // Filters State
  const [selectedEntity, setSelectedEntity] = useState('');
  const [selectedBranch, setSelectedBranch] = useState('');
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

  // Auto Fetch Details when Selection Changes or Date Filter Changes
  useEffect(() => {
    let isMounted = true;

    const fetchDetails = async () => {
      // Logic: If selection is empty -> Fetch ALL. If selection exists -> Fetch Selected.
      const idsToFetch = selectedLeaves.size > 0
        ? Array.from(selectedLeaves)
        : getAllLeafIds();

      // Ensure we have IDs to fetch (edge case: empty tree)
      if (idsToFetch.length === 0) {
        if (isMounted) {
          setBudgetDetails([]);
          setActualDetails([]);
        }
        return;
      }

      if (isMounted) {
        setLoadingDetails(true);
        setLoadingActuals(true);
      }

      try {
        const payload = {
          conso_gls: idsToFetch,
          start_date: actualDateFilter.startDate,
          end_date: actualDateFilter.endDate,
          entities: selectedEntity ? [selectedEntity] : [],     // Add Entity Filter
          branches: selectedBranch ? [selectedBranch] : []      // Add Branch Filter
        };

        // Parallel Fetch
        const results = await Promise.allSettled([
          api.post('/budgets/details', payload),
          api.post('/budgets/actuals-transactions', payload)
        ]);

        if (!isMounted) return;

        // --- Process Budget ---
        if (results[0].status === 'fulfilled') {
          const rawBudget = results[0].value.data || [];
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
        } else {
          console.error("Budget Details Fetch Failed", results[0].reason);
          setBudgetDetails([]);
        }

        // --- Process Actual (Transactions) ---
        if (results[1].status === 'fulfilled') {
          const rawActual = results[1].value.data || [];
          setActualDetails(rawActual);
        } else {
          console.error("Actual Transactions Fetch Failed", results[1].reason);
          setActualDetails([]);
        }

      } catch (err) {
        console.error("Fetch Details Error", err);
      } finally {
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
  }, [selectedLeaves, getAllLeafIds, actualDateFilter, selectedEntity, selectedBranch]); // Add dependencies

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
              onChange={(e) => setSelectedBranch(e.target.value)}
              disabled={!selectedEntity}
            >
              <MenuItem value=""><em>All Branches</em></MenuItem>
              {availableBranches.map((branch) => (
                <MenuItem key={branch} value={branch}>{branch}</MenuItem>
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