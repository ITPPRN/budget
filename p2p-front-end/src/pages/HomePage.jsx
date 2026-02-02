import React, { useState, useEffect, useMemo } from 'react';
import { Box, Grid, Typography, Button, Container, Select, MenuItem, FormControl, InputLabel, Stack } from '@mui/material';
import { useAuth } from '../hooks/useAuth';
import api from '../utils/api/axiosInstance';
import { STATIC_FILTER_OPTIONS } from '../constants/budgetData';
import StatCard from '../components/Dashboard/StatCard';
import DepartmentTable from '../components/Dashboard/DepartmentTable';
import BudgetChart from '../components/Dashboard/BudgetChart';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import AttachMoneyIcon from '@mui/icons-material/AttachMoney';
import AssignmentTurnedInIcon from '@mui/icons-material/AssignmentTurnedIn';
import AccountBalanceIcon from '@mui/icons-material/AccountBalance';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import { useNavigate } from 'react-router-dom';

const HomePage = () => {
  const { user } = useAuth();
  const navigate = useNavigate();

  const [loading, setLoading] = useState(true);
  const [totalBudget, setTotalBudget] = useState(0);
  const [chartData, setChartData] = useState([]);
  const [departmentData, setDepartmentData] = useState([]);

  // --- Filter State ---
  const [orgStructure, setOrgStructure] = useState([]);
  const [selectedEntity, setSelectedEntity] = useState('');
  const [selectedBranch, setSelectedBranch] = useState('');

  // Common Grid Item Styles
  const gridItemStyle = { minWidth: 0 };

  // 1. Fetch Organization Structure on Mount
  useEffect(() => {
    const fetchOrgStructure = async () => {
      try {
        const res = await api.get('/budgets/organization-structure');
        setOrgStructure(res.data || []);
      } catch (err) {
        console.error("Failed to fetch organization structure", err);
      }
    };
    fetchOrgStructure();
  }, []);

  // 2. Computed Branches based on Selected Entity
  const availableBranches = useMemo(() => {
    if (!selectedEntity) return [];
    const entityData = orgStructure.find(e => e.entity === selectedEntity);
    return entityData ? entityData.branches : [];
  }, [orgStructure, selectedEntity]);



  // 3. Build Map: Leaf ID -> Level 1 Name (Keep for fallback or categorization if needed)
  // Actually, user wants "Department" column to be the real Department name (e.g. HR), not Budget Group.
  // So we use item.department directly.

  // 4. Fetch Dashboard Data when Filters Change
  useEffect(() => {
    const fetchDashboardData = async () => {
      setLoading(true);
      try {
        // If filtering by Entity, we need to pass it.
        // If "All", we pass nothing.
        const payload = {
          conso_gls: [], // Empty acts as "All" or similar depending on backend? Backend checks length > 0.
          // Wait, previous code sent "leafToCategoryMap.allLeaves".
          // If backend treats empty "conso_gls" as "Fetch All", we are good.
          // Let's assume we need to send filters if specific filtering is actually needed by backend logic
          // OR if backend returns everything by default.
          // Current Backend: if len > 0 -> Where In. Else -> No filter (All).
          // So sending empty arrays is correct for "Select All".
          entities: selectedEntity ? [selectedEntity] : [],
          branches: selectedBranch ? [selectedBranch] : []
        };

        const res = await api.post('/budgets/details', payload);
        const data = res.data || [];

        // Aggregation
        let sumTotal = 0;
        const monthlySums = {
          "JAN": 0, "FEB": 0, "MAR": 0, "APR": 0, "MAY": 0, "JUN": 0,
          "JUL": 0, "AUG": 0, "SEP": 0, "OCT": 0, "NOV": 0, "DEC": 0
        };

        // Grouping Logic: Key = Unique ID for Table Row
        // User requirement: Group duplicates ONLY if same Entity & Branch.
        // Meaning: (Dept="HR", Entity="A", Branch="1") is different from (Dept="HR", Entity="A", Branch="2").
        // Key = `${dept}|${entity}|${branch}`
        const deptGroups = {};

        data.forEach(item => {
          const amt = parseFloat(item.year_total || 0);
          sumTotal += amt;

          // Monthly Sums (Global for Chart)
          item.budget_amounts?.forEach(m => {
            const mName = m.month.toUpperCase().substring(0, 3); // Ensure format JAN, FEB
            if (monthlySums.hasOwnProperty(mName)) {
              monthlySums[mName] += parseFloat(m.amount || 0);
            }
          });

          // Table Grouping
          // Display Name Logic:
          // If Filtered by Branch -> Just "Department Name"
          // If All Branches -> "Department Name (Branch)" or "Department Name (Entity - Branch)"
          const deptName = item.department || "(No Dept)";
          const entityName = item.entity || "-";
          const branchName = item.branch || "-";

          // Unique Key to separate rows
          const key = `${deptName}|${entityName}|${branchName}`;

          if (!deptGroups[key]) {
            // Construct Display Label
            let label = deptName;
            if (!selectedBranch) {
              // If viewing All or just Entity, show context to differentiate
              if (selectedEntity) {
                label = `${deptName} (${branchName})`;
              } else {
                label = `${deptName} (${entityName} - ${branchName})`;
              }
            }

            deptGroups[key] = {
              name: label,
              budget: 0,
              spending: 0,
              deptRaw: deptName // For sorting if needed
            };
          }
          deptGroups[key].budget += amt;
        });

        setTotalBudget(sumTotal);

        // Transform for Chart (Monthly)
        const monthsOrder = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];
        const monthlyChartData = monthsOrder.map(m => ({
          name: m,
          budget: monthlySums[m],
          actual: 0 // Placeholder
        }));
        setChartData(monthlyChartData);

        // Transform for Table
        const deptTableData = Object.values(deptGroups).sort((a, b) => b.budget - a.budget); // Sort by budget desc
        setDepartmentData(deptTableData);

      } catch (err) {
        console.error("Dashboard Fetch Error", err);
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, [selectedEntity, selectedBranch]); // dependencies

  // Format Helpers
  const formatMB = (val) => {
    const m = val / 1000000;
    return `${m.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} MB`;
  };

  return (
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
        <Stack direction="row" spacing={2} sx={{ minWidth: 300 }}>
          <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
            <InputLabel>Entity (บริษัท)</InputLabel>
            <Select
              value={selectedEntity}
              label="Entity (บริษัท)"
              onChange={(e) => {
                setSelectedEntity(e.target.value);
                setSelectedBranch(''); // Always reset branch when Entity changes
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
              onChange={(e) => setSelectedBranch(e.target.value)}
            >
              <MenuItem value=""><em>All Branches</em></MenuItem>
              {availableBranches.map((branch) => (
                <MenuItem key={branch} value={branch}>{branch}</MenuItem>
              ))}
            </Select>
          </FormControl>
        </Stack>
      </Box>

      {/* KPI Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={4} sx={gridItemStyle}>
          <StatCard
            title="Total Organization Budget"
            value={loading ? "..." : formatMB(totalBudget)}
            subValues={[`of ${formatMB(totalBudget)} (Approved)`]}
            icon={<AccountBalanceIcon />}
            color="primary.main"
          />
        </Grid>
        <Grid item xs={12} md={4} sx={gridItemStyle}>
          <StatCard
            title="Remaining Budget"
            value={loading ? "..." : formatMB(totalBudget - 0)}
            subValues={["100% Remaining"]}
            icon={<AttachMoneyIcon />}
            color="#36b9cc" // Teal
          />
        </Grid>
        <Grid item xs={12} md={4} sx={gridItemStyle}>
          <StatCard
            title="Department Status Alert"
            value="N/A"
            subValues={["No data available"]}
            icon={<AssignmentTurnedInIcon />}
            color="#e74a3b" // Red
          />
        </Grid>
      </Grid>

      {/* Charts & Tables */}
      <Grid container spacing={3}>
        <Grid item xs={12} md={4} sx={{ ...gridItemStyle, width: '100%' }}>
          {/* Department Status Table */}
          <Box sx={{ height: 450, width: '100%' }}>
            <DepartmentTable data={departmentData} />
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
  );
};

export default HomePage;