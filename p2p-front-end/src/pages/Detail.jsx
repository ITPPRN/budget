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

  // Auto Fetch Details when Selection Changes
  useEffect(() => {
    let isMounted = true;

    const fetchDetails = async () => {
      // Logic: If selection is empty -> Fetch ALL. If selection exists -> Fetch Selected.
      const idsToFetch = selectedLeaves.size > 0
        ? Array.from(selectedLeaves)
        : getAllLeafIds();

      // Ensure we have IDs to fetch (edge case: empty tree)
      if (idsToFetch.length === 0) {
        if (isMounted) setBudgetDetails([]);
        return;
      }

      if (isMounted) setLoadingDetails(true);
      try {
        const payload = { conso_gls: idsToFetch };
        const res = await api.post('/budgets/details', payload);

        if (!isMounted) return;

        const rawData = res.data || [];

        // Aggregate Data by GL Code + GL Name
        const aggregatedMap = new Map();

        rawData.forEach(item => {
          const key = `${item.conso_gl}|${item.gl_name}`;

          if (!aggregatedMap.has(key)) {
            // Clone to avoid mutating original if needed, distinct object
            aggregatedMap.set(key, { ...item, budget_amounts: [...(item.budget_amounts || [])] });
          } else {
            const existing = aggregatedMap.get(key);

            // Sum Year Total
            existing.year_total = (parseFloat(existing.year_total) || 0) + (parseFloat(item.year_total) || 0);

            // Sum Monthly Amounts
            // Assuming budget_amounts is array of { month: 'JAN', amount: 100 }
            const existingAmounts = existing.budget_amounts;
            const newAmounts = item.budget_amounts || [];

            newAmounts.forEach(newAmt => {
              const matchIndex = existingAmounts.findIndex(ea => ea.month === newAmt.month);
              if (matchIndex > -1) {
                existingAmounts[matchIndex].amount = (parseFloat(existingAmounts[matchIndex].amount) || 0) + (parseFloat(newAmt.amount) || 0);
              } else {
                existingAmounts.push({ ...newAmt });
              }
            });
          }
        });

        if (isMounted) setBudgetDetails(Array.from(aggregatedMap.values()));
      } catch (error) {
        console.error("Fetch details error:", error);
      } finally {
        if (isMounted) setLoadingDetails(false);
      }
    };

    fetchDetails();

    return () => {
      isMounted = false;
    };
  }, [selectedLeaves, getAllLeafIds]);

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
          minWidth: 0
        }}>
          {/* Top: Budget Table */}
          <BudgetTable
            loading={loadingDetails}
            data={budgetDetails}
            selectedCount={selectedLeaves.size}
          />

          {/* Bottom: Actual Table */}
          <ActualTable />
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