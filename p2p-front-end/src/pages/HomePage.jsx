import React, { useState, useEffect, useMemo } from 'react';
import { Box, Grid, Typography, Button, Container, Select, MenuItem, FormControl, InputLabel, Stack, Paper } from '@mui/material';
import { useAuth } from '../hooks/useAuth';
import api from '../utils/api/axiosInstance';
import { STATIC_FILTER_OPTIONS } from '../constants/budgetData';
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
import { useNavigate } from 'react-router-dom';
import ErrorBoundary from '../components/ErrorBoundary';

const gridItemStyle = { display: 'flex', flexDirection: 'column' }; // Re-add if missing or define inline

const HomePage = () => {
  const { user } = useAuth();
  const navigate = useNavigate();

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
      const payload = {
        entities: selectedEntity ? [selectedEntity] : [],
        branches: selectedBranch ? [selectedBranch] : [],
        departments: contextDept ? [contextDept] : [], // Level 1: Master
        nav_codes: isChartFocus && focusNavCode ? [focusNavCode] : [], // Level 2: Sub-Dept (NavCode)
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
  }, [selectedEntity, selectedBranch, selectedDepartment, page, rowsPerPage, orderBy, order]);

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
      // Level 1: Drill Down to Master -> Sub List
      setSelectedDepartment(deptName);
      setPage(0);
    } else {
      // Level 2: Select Sub-Item for Graph
      if (selectedSubDept === deptName) {
        setSelectedSubDept(''); // Deselect (Revert to Master Aggregation)
      } else {
        setSelectedSubDept(deptName);
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
          {/* ... Box content same as before ... */}
          <Box>
            <Typography variant="h4" sx={{ fontWeight: 'bold', color: 'primary.main' }}>
              Dashboard
            </Typography>
            <Typography variant="subtitle1" color="text.secondary">
              ภาพรวมงบประมาณ
            </Typography>
          </Box>

          {/* Filter Controls */}
          <Stack direction="row" spacing={2} sx={{ minWidth: 300, alignItems: 'center' }}>


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

export default HomePage;