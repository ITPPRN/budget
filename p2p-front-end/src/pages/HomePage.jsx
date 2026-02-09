import React, { useState, useEffect, useMemo } from 'react';
import { Box, Grid, Typography, Button, Container, Select, MenuItem, FormControl, InputLabel, Stack, Paper } from '@mui/material';
import { useAuth } from '../hooks/useAuth';
import api from '../utils/api/axiosInstance';
import { STATIC_FILTER_OPTIONS } from '../constants/budgetData';
import StatCard from '../components/Dashboard/StatCard';
import { TotalBudgetCard, RemainingBudgetCard, DepartmentAlertCard } from '../components/Dashboard/BudgetStatsCard';
import DepartmentTable from '../components/Dashboard/DepartmentTable';
import BudgetChart from '../components/Dashboard/BudgetChart';
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
  const [selectedDepartment, setSelectedDepartment] = useState(''); // Restored for Table Click
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

  // 4. Fetch Dashboard Data when Filters or Pagination Change
  useEffect(() => {
    const fetchDashboardData = async () => {
      // Prevent loading flash if just selecting department? Maybe keep it for feedback.
      setLoading(true);
      try {
        const payload = {
          entities: selectedEntity ? [selectedEntity] : [],
          branches: selectedBranch ? [selectedBranch] : [],
          // departments: [], // Removed
          entities: selectedEntity ? [selectedEntity] : [],
          branches: selectedBranch ? [selectedBranch] : [],
          departments: selectedDepartment ? [selectedDepartment] : [], // Pass Department Filter
          page: page + 1, // Backend is 1-based
          limit: rowsPerPage,
          sort_by: orderBy,
          sort_order: order
        };

        console.log("Dashboard: Fetching Summary...", payload);
        const { data } = await api.post('/budgets/dashboard-summary', payload);
        console.log("Dashboard: Summary Received", data);

        setTotalBudget(data.total_budget || 0);
        setTotalActual(data.total_actual || 0);
        // Only update Table Data if NOT filtering by department (to keep the list visible)
        // OR better: Always update data, but if filtering by dept, the chart updates. 
        // User wants: "Select 1 Dept -> Graph shows only that Dept".
        // BUT: Does user want table to filter to 1 row too? Usually yes ("Drill down"). 
        // If user wants table to stay giving context but CHART to filter, we need separate API calls or client side filtering.
        // User said: "แสดงเฉพาะกราฟของที่เราเลือก" (Show graph of what we selected).
        // If we filter API, the table will also shrink to 1 row. This is usually acceptable drill-down behavior.
        // Let's stick to API filtering for consistency of "Total Budget Card" etc.

        setTotalCount(data.total_count || 0);

        // Update Alert Counts from API (Global)
        setAlertCounts({
          over: data.over_budget_count || 0,
          near: data.near_limit_count || 0
        });

        setOrgStructure(prev => prev.length ? prev : []); // Keep structure

        const mappedChart = (data.chart_data || []).map(item => ({
          name: item.month,
          budget: item.budget,
          actual: item.actual
        }));
        setChartData(mappedChart);

        const transformedTable = (data.department_data || []).map(d => ({
          name: d.department,
          budget: d.budget,
          spending: d.actual,
          deptRaw: d.department
        }));
        setDepartmentData(transformedTable);

      } catch (err) {
        console.error("Dashboard Fetch Error", err);
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardData();
    fetchDashboardData();
  }, [selectedEntity, selectedBranch, selectedDepartment, page, rowsPerPage, orderBy, order]); // Added selectedDepartment

  // Handle Row Click
  const handleDepartmentClick = (deptName) => {
    // Toggle selection: If clicking same dept, unselect.
    console.log("Toggling Department:", deptName);
    setSelectedDepartment(prev => prev === deptName ? '' : deptName);
    setPage(0); // Reset pagination
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
        <Grid container spacing={3}>
          <Grid item xs={12} md={4} sx={{ ...gridItemStyle, width: '100%' }}>
            {/* Department Status Table */}
            <Box sx={{ height: 500, width: '100%' }}>
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
                selectedDept={selectedDepartment} // Prop
                onRowClick={handleDepartmentClick} // Prop
              />
            </Box>
          </Grid>
          <Grid item xs={12} md={8} sx={{ ...gridItemStyle, width: '100%' }}>
            {/* Main Chart */}
            <Box sx={{ height: 450, width: '100%' }}>
              <BudgetChart data={chartData} />
            </Box>
          </Grid>
        </Grid>
      </Box>
    </ErrorBoundary>
  );
};

export default HomePage;