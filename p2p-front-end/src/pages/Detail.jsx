import React, { useState, useEffect } from 'react';
import { Box } from '@mui/material';
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
          end_date: actualDateFilter.endDate
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
          console.error("Budget fetch failed:", results[0].reason);
        }

        // --- Process Actual (Transactions) ---
        if (results[1].status === 'fulfilled') {
          const rawActual = results[1].value.data || [];
          setActualDetails(rawActual);
        } else {
          console.error("Actual fetch failed:", results[1].reason);
          setActualDetails([]);
        }

      } catch (error) {
        console.error("Fetch details error:", error);
      } finally {
        if (isMounted) {
          setLoadingDetails(false);
          setLoadingActuals(false);
        }
      }
    };

    fetchDetails();

    return () => {
      isMounted = false;
    };
  }, [selectedLeaves, getAllLeafIds, actualDateFilter]);

  return (
    <Box sx={{ p: 2, height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>

      {/* Header */}
      <Box sx={{ mb: 1, flexShrink: 0 }}>
        <Box sx={{ color: 'primary.main', fontWeight: 'bold', fontSize: '1.5rem' }}>
          รายงานรายละเอียด
        </Box>
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
          overflowY: { xs: 'auto', md: 'hidden' }, // Scroll on mobile, hidden on desktop
          height: '100%',
          minWidth: 0,
          gap: 2 // Explicit gap between tables
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