import React, { useState, useEffect, useMemo } from 'react';
import { Box, Grid, Typography, Button, Container, Select, MenuItem, FormControl, InputLabel, Stack, Paper, OutlinedInput, Popover, InputAdornment } from '@mui/material';
import { useAuth } from '../hooks/useAuth';
import api from '../utils/api/axiosInstance';
import { BudgetProvider, useBudget } from '../contexts/BudgetContext';
import StatCard from '../components/Dashboard/StatCard';
import { TotalBudgetCard, RemainingBudgetCard, DepartmentAlertCard } from '../components/Dashboard/BudgetStatsCard';
import DepartmentTable from '../components/Dashboard/DepartmentTable';
import BudgetChart from '../components/Dashboard/BudgetChart';
import CapexSection from '../components/Dashboard/CapexSection';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import AttachMoneyIcon from '@mui/icons-material/AttachMoney';
import AssignmentTurnedInIcon from '@mui/icons-material/AssignmentTurnedIn';
import AccountBalanceIcon from '@mui/icons-material/AccountBalance';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import { useNavigate } from 'react-router-dom';
import ErrorBoundary from '../components/ErrorBoundary';
import FilterPane from '../components/Budget/FilterPane';

const gridItemStyle = { display: 'flex', flexDirection: 'column' }; // Re-add if missing or define inline

const HomePageContent = () => {
  const { user } = useAuth();
  const navigate = useNavigate();
  const { filterOptions, isLoading: filterLoading, selectedLeaves } = useBudget();

  const [anchorEl, setAnchorEl] = useState(null);
  const openFilter = Boolean(anchorEl);

  const handleOpenFilter = (event) => setAnchorEl(event.currentTarget);
  const handleCloseFilter = () => setAnchorEl(null);

  const [isSyncing, setIsSyncing] = useState(false); // New state

  const [loading, setLoading] = useState(true);
  const [totalBudget, setTotalBudget] = useState(0);
  const [totalActual, setTotalActual] = useState(0); // New State
  const [chartData, setChartData] = useState([]);
  const [departmentData, setDepartmentData] = useState([]);

  // Filters State
  const [selectedEntity, setSelectedEntity] = useState('');
  const [selectedBranch, setSelectedBranch] = useState('');
  const [selectedDepartment, setSelectedDepartment] = useState(''); // Level 1: Master Dept (Controls Table Context)
  const [selectedSubDept, setSelectedSubDept] = useState(''); // Level 2: Sub Dept (Controls Chart & Highlight)
  const [orgStructure, setOrgStructure] = useState([]);

  // Alert Counts State
  const [alertCounts, setAlertCounts] = useState({ over: 0, near: 0 });

  // Pagination State
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [totalCount, setTotalCount] = useState(0);

  // Sorting State
  const [order, setOrder] = useState('desc');
  const [orderBy, setOrderBy] = useState('actual');

  // Derived state for branches
  const availableBranches = useMemo(() => {
    if (!selectedEntity) return [];
    const entityObj = orgStructure.find(o => o.entity === selectedEntity);
    return entityObj ? entityObj.branches : [];
  }, [selectedEntity, orgStructure]);

  // Derived state for departments - REMOVED

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

  // Handlers
  const handleChangePage = (event, newPage) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleRequestSort = (property) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
    setPage(0); // Reset to first page on sort change
  };

  // 4. Fetch Dashboard Data
  // Strategy:
  // - Context Fetch: Depends on selectedEntity, selectedBranch, selectedDepartment (Master)
  //   Updates: departmentData (Table), totalBudget, alertCounts.
  //   Also updates: chartData (Default aggregate) IF no sub-dept selected.
  // - Focus Fetch: Depends on selectedSubDept.
  //   Updates: chartData (Specific).

  const fetchDashboardData = async (overrideParams = {}) => {
    // Determine effective filters
    // If fetching for Table Context: use selectedDepartment
    // If fetching for Chart Focus: use selectedSubDept (if exists) OR selectedDepartment

    const isChartFocus = overrideParams.isChartFocus;
    // Context Dept (Level 1) - Always the Master Department
    const contextDept = selectedDepartment;
    // Focus Dept (Level 2) - Specific Sub-Department (NavCode)
    const focusNavCode = overrideParams.targetDept;

    setLoading(true);
    try {
      // Get the synced Actuals configuration from localStorage
      let syncConfig = JSON.parse(localStorage.getItem('dm_lastSyncedConfig') || '{}');

      // If localStorage is empty, try fetching from Backend directly (Single Source of Truth)
      if (!syncConfig.actualYear) {
        try {
          const configRes = await api.get('/budgets/configs');
          const configs = configRes.data || {};
          if (configs.actualYear) {
            syncConfig = {
              actualYear: configs.actualYear,
              selectedMonths: JSON.parse(configs.selectedMonths || '[]')
            };
            // Cache it for subsequent loads
            localStorage.setItem('dm_lastSyncedConfig', JSON.stringify(syncConfig));
          }
        } catch (e) { console.error("Failed to fetch fallback configs", e); }
      }

      const actualYear = syncConfig.actualYear || new Date().getFullYear();
      const selectedMonths = syncConfig.selectedMonths || [];

      const payload = {
        entities: selectedEntity ? [selectedEntity] : [],
        branches: selectedBranch ? [selectedBranch] : [],
        departments: contextDept ? [contextDept] : [], // Level 1: Master
        nav_codes: isChartFocus && focusNavCode ? [focusNavCode] : [], // Level 2: Sub-Dept (NavCode)
        budget_gls: Array.from(selectedLeaves).map(id => id.split('|')[0]).filter(code => code !== ""),
        year: String(actualYear),
        months: selectedMonths,
        budget_file_id: syncConfig.selectedBudget,
        capex_file_id: syncConfig.selectedCapexBg,
        capex_actual_file_id: syncConfig.selectedCapexActual,
        page: page + 1,
        limit: rowsPerPage,
        sort_by: orderBy,
        sort_order: order
      };

      console.log(`Dashboard Fetch (${isChartFocus ? 'Focus' : 'Context'}):`, payload);
      const { data } = await api.post('/budgets/dashboard-summary', payload);

      if (isChartFocus) {
        // Update Chart & Totals (but NOT Table or TotalCount)
        setTotalBudget(data.total_budget || 0);
        setTotalActual(data.total_actual || 0);
        setAlertCounts({
          over: data.over_budget_count || 0,
          near: data.near_limit_count || 0
        });

        const mappedChart = (data.chart_data || []).map(item => ({
          name: item.month,
          budget: item.budget,
          actual: item.actual
        }));
        setChartData(mappedChart);
      } else {
        // Update Context (Table, Totals, Default Chart)
        setTotalBudget(data.total_budget || 0);
        setTotalActual(data.total_actual || 0);
        setTotalCount(data.total_count || 0);
        setAlertCounts({
          over: data.over_budget_count || 0,
          near: data.near_limit_count || 0
        });

        // Chart (Default)
        if (!selectedSubDept) {
          const mappedChart = (data.chart_data || []).map(item => ({
            name: item.month,
            budget: item.budget,
            actual: item.actual
          }));
          setChartData(mappedChart);
        }

        // Table
        const transformedTable = (data.department_data || []).map(d => ({
          name: d.department,
          budget: d.budget,
          spending: d.actual,
          deptRaw: d.department
        }));
        setDepartmentData(transformedTable);
      }

    } catch (err) {
      console.error("Dashboard Fetch Error", err);
    } finally {
      setLoading(false);
    }
  };

  // Effect: Context Change (Global Filters, Pagination, Master Drill-Down)
  useEffect(() => {
    fetchDashboardData({ isChartFocus: false });
  }, [selectedEntity, selectedBranch, selectedDepartment, selectedLeaves, page, rowsPerPage, orderBy, order]);


  // Effect: Sub-Dept Change (Chart Focus)
  useEffect(() => {
    if (selectedSubDept) {
      fetchDashboardData({ isChartFocus: true, targetDept: selectedSubDept });
    } else {
      // If cleared, revert to context chart (handled by context effect if dependencies change, or we trigger manual?)
      // If selectedSubDept becomes empty, we need to revert chart.
      // Re-running context fetch is safest/easiest.
      if (selectedDepartment) {
        fetchDashboardData({ isChartFocus: false });
      }
    }
  }, [selectedSubDept]);


  // Handle Row Click
  const handleDepartmentClick = (deptName) => {
    console.log("Clicked:", deptName);

    if (!selectedDepartment) {
      // 1. Root Level (Top Spenders) -> Drill Down
      setSelectedDepartment(deptName);
      setPage(0);
    } else {
      // 2. Currently inside a drill-down view
      if (selectedDepartment === "None") {
        // Special logic for "None": 
        if (selectedSubDept === deptName) {
          // Click same sub-item again -> CLEAR and JUMP BACK to root table
          setSelectedSubDept('');
          setSelectedDepartment(''); 
          setPage(0);
        } else {
          // Click first time or click different sub-item -> Select it for graph but STAY in view
          setSelectedSubDept(deptName);
        }
      } else if (selectedDepartment === deptName) {
        // Toggle logic for normal departments: clicking the same row again returns to root list
        setSelectedDepartment('');
        setSelectedSubDept('');
        setPage(0);
      } else {
        // Toggle highlight for sub-items (if any)
        setSelectedSubDept(prev => prev === deptName ? '' : deptName);
      }
    }
  };

  // Format Helpers
  const formatMB = (val) => {
    const m = val / 1000000;
    return `${m.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} MB`;
  };

  return (
    <ErrorBoundary>
      <Box sx={{ mt: 4, mb: 4, px: 3, width: '100%' }}>
        {/* Header & Filters */}
        <Box sx={{ display: 'flex', flexDirection: { xs: 'column', md: 'row' }, justifyContent: 'space-between', alignItems: 'flex-start', mb: 3, gap: 2 }}>
          <Box>
            <Typography variant="h4" sx={{ fontWeight: 'bold', color: 'primary.main' }}>
              Dashboard
            </Typography>
            <Typography variant="subtitle1" color="text.secondary">
              ภาพรวมงบประมาณ
            </Typography>
          </Box>

          {/* Filter Controls */}
          <Stack direction="row" sx={{ minWidth: 300, alignItems: 'center', flexWrap: 'wrap', gap: 2 }}>
            <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
              <InputLabel>Entity (บริษัท)</InputLabel>
              <Select
                value={selectedEntity}
                label="Entity (บริษัท)"
                onChange={(e) => {
                  setSelectedEntity(e.target.value);
                  setSelectedBranch('');
                  setSelectedDepartment(''); // Reset
                }}
              >
                <MenuItem value=""><em>All Entities</em></MenuItem>
                {orgStructure.map((org) => (
                  <MenuItem key={org.entity} value={org.entity}>{org.entity}</MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }} disabled={!selectedEntity}>
              <InputLabel>Branch (สาขา)</InputLabel>
              <Select
                value={selectedBranch}
                label="Branch (สาขา)"
                onChange={(e) => {
                  setSelectedBranch(e.target.value);
                  setSelectedDepartment(''); // Reset
                }}
              >
                <MenuItem value=""><em>All Branches</em></MenuItem>
                {availableBranches.map((branch) => (
                  <MenuItem key={branch.name} value={branch.name}>{branch.name}</MenuItem>
                ))}
              </Select>
            </FormControl>

            <FormControl size="small" sx={{ minWidth: 250, flexGrow: 1, bgcolor: 'white', borderRadius: 1 }}>
              <InputLabel shrink={selectedLeaves.size > 0 || openFilter}>
                Account Filter
              </InputLabel>
              <OutlinedInput
                notched={selectedLeaves.size > 0 || openFilter}
                label="Account Filter"
                readOnly
                onClick={handleOpenFilter}
                value={selectedLeaves.size > 0 ? `เลือกแล้ว ${selectedLeaves.size} รายการ` : ""}
                placeholder={selectedLeaves.size === 0 && !openFilter ? " " : " "}
                endAdornment={
                  <InputAdornment position="end">
                    <ArrowDropDownIcon color="action" />
                  </InputAdornment>
                }
                sx={{
                  height: '40px',
                  bgcolor: 'white',
                  cursor: 'pointer',
                  '& input': {
                    cursor: 'pointer',
                  }
                }}
              />
              <Popover
                open={openFilter}
                anchorEl={anchorEl}
                onClose={handleCloseFilter}
                anchorOrigin={{
                  vertical: 'bottom',
                  horizontal: 'left',
                }}
                transformOrigin={{
                  vertical: 'top',
                  horizontal: 'left',
                }}
                PaperProps={{
                  sx: {
                    width: anchorEl ? anchorEl.clientWidth : 350,
                    height: 400,
                    mt: 1,
                    p: 0, // FilterPane already has internal padding
                    overflow: 'hidden'
                  }
                }}
              >
                {openFilter && <FilterPane compact={true} />}
              </Popover>
            </FormControl>
          </Stack>
        </Box>

        {/* 1. Summary Cards */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          {/* Total Budget Card */}
          <Grid item xs={12} md={4}>
            <TotalBudgetCard totalBudget={totalBudget} totalActual={totalActual} />
          </Grid>

          {/* Remaining Budget Card */}
          <Grid item xs={12} md={4}>
            <RemainingBudgetCard totalBudget={totalBudget} totalActual={totalActual} />
          </Grid>

          {/* Department Alert Card */}
          <Grid item xs={12} md={4}>
            <DepartmentAlertCard overBudgetCount={alertCounts.over} nearLimitCount={alertCounts.near} />
          </Grid>
        </Grid>

        {/* Charts & Tables */}
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={3} sx={{ width: '100%', mb: 4 }}>
          {/* Department Status Table Container */}
          <Box sx={{ flex: 1, minWidth: 0, height: 500 }}>
            <DepartmentTable
              data={departmentData}
              count={totalCount}
              page={page}
              rowsPerPage={rowsPerPage}
              onPageChange={handleChangePage}
              onRowsPerPageChange={handleChangeRowsPerPage}
              orderBy={orderBy}
              order={order}
              onRequestSort={handleRequestSort}
              selectedDept={selectedSubDept || selectedDepartment}
              onRowClick={handleDepartmentClick}
              onBack={() => {
                setSelectedDepartment('');
                setSelectedSubDept('');
                setPage(0);
              }}
            />
          </Box>

          {/* Budget Chart Container */}
          <Box sx={{ flex: 1, minWidth: 0, height: 500 }}>
            <BudgetChart
              data={chartData}
              title="Budget vs Actual"
              selectedDept={selectedSubDept || selectedDepartment || "ALL"}
            />
          </Box>
        </Stack>

        {/* CAPEX Section */}
        <CapexSection
          globalEntity={selectedEntity}
          orgStructure={orgStructure}
        />
      </Box>
    </ErrorBoundary>
  );
};

const HomePage = () => {
  return (
    <BudgetProvider>
      <HomePageContent />
    </BudgetProvider>
  );
};

export default HomePage;