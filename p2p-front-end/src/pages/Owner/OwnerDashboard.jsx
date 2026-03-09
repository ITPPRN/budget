import React, { useState, useEffect } from 'react';
import { Box, Grid, Typography, FormControl, Select, MenuItem, Stack, Paper, CircularProgress, Button, IconButton, Tooltip, LinearProgress, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, InputLabel } from '@mui/material';
import MoreHorizIcon from '@mui/icons-material/MoreHoriz';
import SyncIcon from '@mui/icons-material/Sync';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import ShoppingBagIcon from '@mui/icons-material/ShoppingBag';
import PieChartIcon from '@mui/icons-material/PieChart';
import TrendingUpIcon from '@mui/icons-material/TrendingUp';
import FilterListIcon from '@mui/icons-material/FilterList';
import DownloadIcon from '@mui/icons-material/Download';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import Inventory2Icon from '@mui/icons-material/Inventory2';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';
import LinkIcon from '@mui/icons-material/Link';
import FilterAltIcon from '@mui/icons-material/FilterAlt';
import ArrowBackIosNewIcon from '@mui/icons-material/ArrowBackIosNew';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import PeopleIcon from '@mui/icons-material/People';
import { useAuth } from '../../hooks/useAuth';
import api from '../../utils/api/axiosInstance';
import BudgetChart from '../../components/Dashboard/OwnerBudgetChart';
import DonutChart, { COLORS } from '../../components/Dashboard/DonutChart';
import ErrorBoundary from '../../components/ErrorBoundary';
import { toast } from 'react-toastify';

// Calculate formatter
const formatCurrency = (value) => {
    // If value is 0 or null/undefined, return "0"
    if (!value) return "0";

    const absValue = Math.abs(value);
    const sign = value < 0 ? "-" : "";

    if (absValue >= 1000000) return `${sign}${(absValue / 1000000).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} MB`;
    if (absValue >= 1000) return `${sign}${(absValue / 1000).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} K`;

    return value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

const MetricCard = ({ title, value, icon: Icon, color = '#4d6eff' }) => (
    <Paper
        sx={{
            p: { xs: 2.5, md: 3 },
            borderRadius: '16px',
            minHeight: { xs: 120, md: 150 },
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center',
            bgcolor: color,
            color: 'white',
            boxShadow: '0 8px 32px rgba(0,0,0,0.08)',
            position: 'relative',
            transition: 'transform 0.2s',
            '&:hover': { transform: 'translateY(-2px)' }
        }}
    >
        <Stack direction="row" spacing={1.5} alignItems="center" sx={{ mb: 1.5 }}>
            <Icon sx={{ fontSize: 20, color: 'white', opacity: 0.9 }} />
            <Typography variant="body2" sx={{ fontWeight: 600, color: 'white', letterSpacing: '0.01em', opacity: 0.9 }}>
                {title}
            </Typography>
        </Stack>
        <Box>
            <Typography variant="h4" sx={{ fontWeight: 800, color: 'white', letterSpacing: '-0.02em', fontSize: { xs: '1.75rem', md: '2.1rem' } }}>
                {value}
            </Typography>
        </Box>
    </Paper>
);

const RemainingBudgetCard = ({ value }) => (
    <Paper
        sx={{
            p: { xs: 2.5, md: 3 },
            borderRadius: '16px',
            minHeight: { xs: 120, md: 150 },
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center',
            bgcolor: '#4d6eff',
            color: 'white',
            boxShadow: '0 8px 32px rgba(0,0,0,0.08)',
            position: 'relative',
            overflow: 'hidden',
        }}
    >
        <Stack direction="row" spacing={1.5} alignItems="center" sx={{ mb: 1.5, position: 'relative', zIndex: 1, color: 'rgba(255,255,255,0.8)' }}>
            <LinkIcon sx={{ fontSize: 20 }} />
            <Typography variant="body2" sx={{ fontWeight: 800, letterSpacing: '0.02em', textTransform: 'uppercase', color: 'white' }}>
                Remaining Budget
            </Typography>
        </Stack>
        <Box sx={{ position: 'relative', zIndex: 1 }}>
            <Typography variant="h4" sx={{ fontWeight: 900, color: 'white', letterSpacing: '-0.03em', fontSize: { xs: '1.75rem', md: '2.4rem' } }}>
                {value}
            </Typography>
        </Box>

    </Paper>
);

const UsageCard = ({ usagePercent }) => (
    <Paper
        sx={{
            p: '10px 24px',
            borderRadius: '40px',
            display: 'flex',
            alignItems: 'center',
            gap: 2,
            bgcolor: 'white',
            boxShadow: '0 4px 12px rgba(0,0,0,0.04)',
            height: 'fit-content',
            minHeight: 56
        }}
    >
        <Typography variant="body2" sx={{ fontWeight: 700, color: '#333', minWidth: 'fit-content', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
            Usage
        </Typography>
        <Box sx={{
            flexGrow: 1,
            height: 12,
            bgcolor: '#f1f5f9',
            borderRadius: '10px',
            position: 'relative',
            overflow: 'hidden',
            minWidth: 120,
            border: '1px solid #e2e8f0'
        }}>
            <Box sx={{
                height: '100%',
                width: `${Math.min(usagePercent, 100)}%`,
                bgcolor: usagePercent > 100 ? '#ef5350' : '#4d6eff',
                borderRadius: '10px',
                transition: 'width 1s ease-in-out'
            }} />
        </Box>
        <Typography variant="body1" sx={{ fontWeight: 600, color: '#666' }}>
            {usagePercent.toFixed(0)}%
        </Typography>
    </Paper>
);

const OwnerDashboard = () => {
    const { user } = useAuth();
    const [loading, setLoading] = useState(true); // Filter initialization
    const [dataLoading, setDataLoading] = useState(false); // Background data fetch
    const [chartMode, setChartMode] = useState('monthly'); // 'monthly' | 'accumulated'
    const [isTopExpenseExpanded, setIsTopExpenseExpanded] = useState(false);

    // Filter Options State (Reset on load)
    const [selectedCompany, setSelectedCompany] = useState('');
    const [selectedBranch, setSelectedBranch] = useState('');
    const [selectedDept, setSelectedDept] = useState('');
    const [selectedYear, setSelectedYear] = useState('');
    const [selectedAccount, setSelectedAccount] = useState(''); // New State

    const [summary, setSummary] = useState({
        totalBudget: 0,
        totalActual: 0,
        capexBudget: 0,
        capexActual: 0,
        chartData: [],
        topExpenses: [],
        overBudgetCount: 0,
        nearLimitCount: 0
    });

    // Filter Options State (Cascading)
    const [orgStructure, setOrgStructure] = useState([]);
    const [filterYears, setFilterYears] = useState([]);
    const [accountFilters, setAccountFilters] = useState([]); // Storage from GetBudgetFilterOptions

    // Derived Branches based on Entity Selection
    const availableBranches = React.useMemo(() => {
        if (!selectedCompany || selectedCompany === 'All' || !Array.isArray(orgStructure)) return [];
        const entityObj = orgStructure.find(o => o.entity === selectedCompany);
        return entityObj ? entityObj.branches : [];
    }, [selectedCompany, orgStructure]);

    // Flatten all unique departments from orgStructure for easy selection
    const allDepartments = React.useMemo(() => {
        if (!Array.isArray(orgStructure)) return [];
        const depts = new Set();
        orgStructure.forEach(entity => {
            entity.branches?.forEach(branch => {
                branch.departments?.forEach(dept => {
                    const deptName = typeof dept === 'string' ? dept : dept.name;
                    if (deptName) depts.add(deptName);
                });
            });
        });
        return Array.from(depts).sort();
    }, [orgStructure]);

    // Derived Accounts for the dropdown based on selected departments
    const availableAccounts = React.useMemo(() => {
        if (!Array.isArray(accountFilters)) return [];
        let filtered = accountFilters;
        if (selectedDept && selectedDept !== 'All') {
            filtered = filtered.filter(f => f.department === selectedDept || f.nav_code === selectedDept);
        }
        // Extract unique Level 3 Names based on current filter or all if none selected
        const accounts = new Set();
        // Since budget filter options returns raw GLs, we just want unique GL names or we need to wait for Phase 11 for the full tree
        // The user asked to "Connect Filter Account dropdown to API". The Owner API has a `/owner/budget-filters` endpoint yielding BudgetFactEntities.
        filtered.forEach(f => {
            if (f.gl_name) accounts.add(f.gl_name); // Or whatever we group by. Let's use gl_name for now to match the payload requirement
        });
        return Array.from(accounts).sort();
    }, [accountFilters, selectedDept]);

    // 2. Fetch Dashboard Data
    const fetchDashboardData = async () => {
        if (!selectedYear) return;
        setDataLoading(true);
        try {
            // Find the conso_gls matching the selected account name
            let targetGls = [];
            if (selectedAccount && selectedAccount !== 'All') {
                targetGls = accountFilters.filter(f => f.gl_name === selectedAccount).map(f => f.conso_gl);
            }

            const payload = {
                year: selectedYear,
                entities: selectedCompany && selectedCompany !== 'All' ? [selectedCompany] : [],
                branches: selectedBranch && selectedBranch !== 'All' ? [selectedBranch] : [],
                departments: selectedDept && selectedDept !== 'All' ? [selectedDept] : [],
                conso_gls: targetGls // Pass to backend
            };



            const { data } = await api.post('/owner/dashboard-summary', payload);

            setSummary({
                totalBudget: data.total_budget || 0,
                totalActual: data.total_actual || 0,
                capexBudget: data.capex_budget || 0,
                capexActual: data.capex_actual || 0,
                chartData: (data.chart_data || []).map(item => ({
                    name: item.month,
                    budget: item.budget,
                    actual: item.actual
                })),
                topExpenses: (data.top_expenses || []).map(item => ({
                    name: item.name,
                    value: Number(item.amount) || 0
                })),
                overBudgetCount: data.over_budget_count || 0,
                nearLimitCount: data.near_limit_count || 0
            });

        } catch (err) {
            console.error("Owner Dashboard Fetch Error", err);
            // toast.error("Failed to load dashboard data"); 
        } finally {
            setDataLoading(false);
        }
    };

    // 4. Sequential Initialization (Sync -> Filters -> Data)
    useEffect(() => {
        const initDashboard = async () => {
            if (!user) return;

            // Prevent double-init if already running
            // (Strict mode in Dev might cause double run, but logic should be idempotent enough)

            console.log("Initializing Dashboard...");
            // setLoading(true); // Ensure loading is on

            // Step 2: Fetch Filters Parallelized (Promise.all)
            try {
                console.log("3. Fetching Filters (Parallel)...");
                const [structRes, listRes, accountRes] = await Promise.all([
                    api.get('/owner/organization-structure'),
                    api.get('/owner/filter-lists'),
                    api.get('/owner/budget-filters') // Fetching the GL options
                ]);

                const structData = structRes.data || [];
                setOrgStructure(structData);

                const years = listRes.data.years || [];
                setFilterYears(years);
                console.log("4. Filters Fetched. Years:", years);

                setAccountFilters(accountRes.data || []);

                // Smart Default logic
                if (!selectedYear && years.length > 0) {
                    if (years.includes('All')) {
                        setSelectedYear('All');
                    } else {
                        const sortedYears = [...years].sort((a, b) => b.localeCompare(a));
                        const defaultYear = sortedYears[0];
                        setSelectedYear(defaultYear);
                    }
                }

                // Smart Auto-Selection logic (only if not already set)
                if (!selectedCompany && structData.length === 1) {
                    const firstEntity = structData[0];
                    setSelectedCompany(firstEntity.entity);
                    if (!selectedBranch && firstEntity.branches?.length === 1) {
                        const firstBranch = firstEntity.branches[0];
                        setSelectedBranch(firstBranch.name);
                        if (!selectedDept && firstBranch.departments?.length === 1) {
                            setSelectedDept(firstBranch.departments[0]);
                        }
                    }
                }

            } catch (err) {
                console.error("Filter Fetch Error", err);
                toast.error("Failed to load filters");
            } finally {
                setLoading(false);
            }
        };

        if (user) {
            initDashboard();
        }
    }, [user]); // Run when User is ready. Empty deps [] might miss 'user' if it loads late.

    // 3. Fetch Data on Filter Change
    useEffect(() => {
        if (selectedYear) {
            console.log("Triggering Fetch Dashboard (Filter Changed)");
            fetchDashboardData();
        }
    }, [selectedYear, selectedCompany, selectedBranch, selectedDept, selectedAccount]);


    const usagePercent = summary.totalBudget > 0
        ? (summary.totalActual / summary.totalBudget) * 100
        : (summary.totalActual > 0 ? 100 : 0);

    const statusLabel =
        summary.totalBudget === 0 && summary.totalActual > 0 ? 'Over Budget' :
            usagePercent > 100 ? 'Over Budget' :
                usagePercent > 80 ? 'Near Limit' : 'In Budget';

    // Calculate chart data based on selected mode (Accumulated vs Monthly)
    const displayChartData = React.useMemo(() => {
        if (!summary.chartData || summary.chartData.length === 0) return [];

        if (chartMode === 'accumulated') {
            let accBudget = 0;
            let accActual = 0;
            return summary.chartData.map(item => {
                accBudget += (item.budget || 0);
                accActual += (item.actual || 0);
                return {
                    name: item.name,
                    budget: accBudget,
                    actual: accActual
                };
            });
        }
        return summary.chartData;
    }, [summary.chartData, chartMode]);

    // Optimized Loading State: Only block IF we have absolutely NO data and NO year yet.
    // If we have a year (from localStorage), let the UI render.
    if (loading && !selectedYear) {
        return (<Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}> <CircularProgress /> </Box>);
    }

    return (
        <ErrorBoundary>
            <Box sx={{
                width: '100%',
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                bgcolor: '#f8f9fc',
                overflow: 'hidden'
            }}>
                {/* Top Loading Indicator (Granular UX) */}
                {dataLoading && (
                    <LinearProgress
                        sx={{
                            position: 'absolute',
                            top: 0,
                            left: 0,
                            right: 0,
                            zIndex: 9999,
                            height: 4,
                            bgcolor: 'rgba(77, 110, 255, 0.1)',
                            '& .MuiLinearProgress-bar': { bgcolor: '#4d6eff' }
                        }}
                    />
                )}

                {/* Fluid Wrapper - Edge to Edge */}
                <Box sx={{
                    width: '100%',
                    maxWidth: 'none !important', // Strictly no max-width
                    margin: '0 !important',
                    px: { xs: 2, md: 4, lg: 4 }, // Added some breathing room back
                    py: { xs: 3, md: 5 },
                    flex: 1,
                    display: 'flex',
                    flexDirection: 'column'
                }}>

                    {/* 1. Integrated Header & Filter Bar */}
                    <Box sx={{ mb: 6 }}>
                        <Stack
                            direction={{ xs: 'column', lg: 'row' }}
                            justifyContent="space-between"
                            alignItems={{ xs: 'flex-start', lg: 'flex-end' }}
                            spacing={3}
                        >
                            <Box>
                                <Typography variant="h4" sx={{
                                    fontWeight: 700,
                                    color: '#043478',
                                    letterSpacing: '-0.02em',
                                    mb: 0.5
                                }}>
                                    Dashboard
                                </Typography>
                                <Typography variant="body1" sx={{ color: '#64748b', fontWeight: 500 }}>
                                    Welcome {user?.name || 'Owner'}
                                </Typography>
                            </Box>

                            {/* Filter Section (Admin Theme Match) */}
                            <Stack direction="row" spacing={2} sx={{ minWidth: 300, flexWrap: 'wrap', justifyContent: { xs: 'flex-start', lg: 'flex-end' } }}>
                                {/* Department Filter */}
                                <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
                                    <InputLabel>Department (แผนก)</InputLabel>
                                    <Select
                                        value={selectedDept}
                                        label="Department (แผนก)"
                                        onChange={(e) => setSelectedDept(e.target.value)}
                                    >
                                        <MenuItem value=""><em>All Departments</em></MenuItem>
                                        {allDepartments.map((deptName) => (
                                            <MenuItem key={deptName} value={deptName}>{deptName}</MenuItem>
                                        ))}
                                    </Select>
                                </FormControl>

                                {/* Account Name Filter */}
                                <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
                                    <InputLabel>Account (หมวดหมู่)</InputLabel>
                                    <Select
                                        value={selectedAccount}
                                        label="Account (หมวดหมู่)"
                                        onChange={(e) => setSelectedAccount(e.target.value)}
                                    >
                                        <MenuItem value=""><em>All Accounts</em></MenuItem>
                                        {availableAccounts.map((accountName) => (
                                            <MenuItem key={accountName} value={accountName}>{accountName}</MenuItem>
                                        ))}
                                    </Select>
                                </FormControl>

                                {/* Company Filter */}
                                <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
                                    <InputLabel>Entity (บริษัท)</InputLabel>
                                    <Select
                                        value={selectedCompany}
                                        label="Entity (บริษัท)"
                                        onChange={(e) => { setSelectedCompany(e.target.value); setSelectedBranch(''); }}
                                    >
                                        <MenuItem value=""><em>All Entities</em></MenuItem>
                                        {orgStructure.map((org) => <MenuItem key={org.entity} value={org.entity}>{org.entity}</MenuItem>)}
                                    </Select>
                                </FormControl>

                                {/* Branch Filter */}
                                <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
                                    <InputLabel>Branch (สาขา)</InputLabel>
                                    <Select
                                        value={selectedBranch}
                                        label="Branch (สาขา)"
                                        onChange={(e) => setSelectedBranch(e.target.value)}
                                        disabled={!selectedCompany}
                                    >
                                        <MenuItem value=""><em>All Branches</em></MenuItem>
                                        {availableBranches.map((b) => <MenuItem key={b.name} value={b.name}>{b.name}</MenuItem>)}
                                    </Select>
                                </FormControl>

                                {/* Year Filter (Owner specific but styling match) */}
                                <FormControl size="small" sx={{ minWidth: 120, bgcolor: 'white', borderRadius: 1 }}>
                                    <InputLabel>Year</InputLabel>
                                    <Select
                                        value={selectedYear}
                                        label="Year"
                                        onChange={(e) => setSelectedYear(e.target.value)}
                                    >
                                        <MenuItem value="" disabled><em>Year</em></MenuItem>
                                        {filterYears.map((y) => <MenuItem key={y} value={y}>{y.replace('FY', '')}</MenuItem>)}
                                    </Select>
                                </FormControl>
                            </Stack>
                        </Stack>
                    </Box>

                    {/* 2. Stats Cards (Redesigned to match screenshot) */}
                    <Grid container spacing={4} sx={{ mb: 6 }} alignItems="stretch">
                        <Grid item xs={12} sm={6} md={3}>
                            <MetricCard
                                title="Approved Expense Budget"
                                value={formatCurrency(summary.totalBudget)}
                                icon={Inventory2Icon}
                                color="#4d6eff"
                            />
                        </Grid>
                        <Grid item xs={12} sm={6} md={3}>
                            <MetricCard
                                title="Actual Spending"
                                value={formatCurrency(summary.totalActual)}
                                icon={LocalOfferIcon}
                                color="#4d6eff"
                            />
                        </Grid>
                        <Grid item xs={12} sm={6} md={3}>
                            <RemainingBudgetCard
                                value={formatCurrency(summary.totalBudget - summary.totalActual)}
                            />
                        </Grid>
                        <Grid item xs={12} md={3}>
                            <UsageCard usagePercent={usagePercent} />
                        </Grid>
                    </Grid>

                    {/* 3. Detailed Analysis (Side-by-Side Row) */}
                    <Grid
                        container
                        spacing={3}
                        sx={{
                            width: '100%',
                            margin: 0,
                            display: 'flex',
                            flexWrap: 'nowrap',
                            flexDirection: { xs: 'column', lg: 'row' }
                        }}
                    >
                        {/* LEFT - Budget vs Actual (Fluid Growth) */}
                        <Grid
                            item
                            sx={{
                                flexGrow: 1,
                                minWidth: 0,
                                width: { xs: '100%', lg: 'auto' }
                            }}>
                            <Paper sx={{
                                p: 3,
                                borderRadius: '24px',
                                height: 550,
                                display: 'flex',
                                flexDirection: 'column',
                                bgcolor: 'rgba(255,255,255,0.7)',
                                boxShadow: '0 10px 40px rgba(0,0,0,0.02)'
                            }}
                            >
                                {/* Header Grouping */}
                                <Stack
                                    direction={{ xs: 'column', sm: 'row' }}
                                    justifyContent="space-between"
                                    alignItems={{ xs: 'flex-start', sm: 'center' }}
                                    spacing={2}
                                    sx={{ mb: 3 }}
                                >
                                    <Box sx={{ display: 'flex', gap: 4, alignItems: 'center' }}>
                                        <Typography variant="h6" sx={{ fontWeight: 800, color: '#333' }}>
                                            Budget vs Actual
                                        </Typography>
                                    </Box>

                                    <Stack direction="row" spacing={2} alignItems="center">
                                        <Paper sx={{ bgcolor: '#000', borderRadius: '4px', p: 0.5, display: 'flex', overflow: 'hidden' }}>
                                            <Button
                                                size="small"
                                                onClick={() => setChartMode('monthly')}
                                                sx={{
                                                    bgcolor: chartMode === 'monthly' ? '#4d6eff' : 'transparent',
                                                    color: 'white',
                                                    borderRadius: '4px',
                                                    fontWeight: 800,
                                                    textTransform: 'none', px: 2,
                                                    fontSize: '0.85rem',
                                                    '&:hover': { bgcolor: chartMode === 'monthly' ? '#4d6eff' : 'rgba(255,255,255,0.1)' }
                                                }}
                                            >
                                                Monthly
                                            </Button>
                                            <Button
                                                size="small"
                                                onClick={() => setChartMode('accumulated')}
                                                sx={{
                                                    bgcolor: chartMode === 'accumulated' ? '#4d6eff' : 'transparent',
                                                    color: 'white',
                                                    borderRadius: '4px',
                                                    fontWeight: 800,
                                                    textTransform: 'none', px: 2,
                                                    fontSize: '0.85rem',
                                                    '&:hover': { bgcolor: chartMode === 'accumulated' ? '#4d6eff' : 'rgba(255,255,255,0.1)' }
                                                }}
                                            >
                                                Accumulated
                                            </Button>
                                        </Paper>

                                        <Box sx={{ display: 'flex', gap: 1 }}>
                                            <IconButton size="small" sx={{ bgcolor: '#4d6eff', color: 'white', '&:hover': { bgcolor: '#3d59cc' } }}>
                                                <FilterAltIcon sx={{ fontSize: 18 }} />
                                            </IconButton>
                                            <IconButton size="small" sx={{ bgcolor: '#4d6eff', color: 'white', '&:hover': { bgcolor: '#3d59cc' } }}>
                                                <DownloadIcon sx={{ fontSize: 18 }} />
                                            </IconButton>
                                        </Box>
                                    </Stack>
                                </Stack>

                                <Box sx={{ flexGrow: 1, width: '100%', mt: 2 }}>
                                    <BudgetChart data={displayChartData} />
                                </Box>

                                {/* Legend at Bottom */}
                                <Stack direction="row" spacing={4} justifyContent="center" sx={{ mt: 2, borderTop: '1px solid #f1f5f9', pt: 3 }}>
                                    <Stack direction="row" spacing={1.5} alignItems="center">
                                        <Box sx={{ width: 12, height: 12, borderRadius: '50%', bgcolor: '#4d6eff' }} />
                                        <Typography variant="body2" sx={{ fontWeight: 700, color: '#64748b' }}>Budget</Typography>
                                    </Stack>
                                    <Stack direction="row" spacing={1.5} alignItems="center">
                                        <Box sx={{ width: 12, height: 12, borderRadius: '50%', bgcolor: '#64b5f6' }} />
                                        <Typography variant="body2" sx={{ fontWeight: 700, color: '#64748b' }}>Actual</Typography>
                                    </Stack>
                                </Stack>
                            </Paper>
                        </Grid>

                        {/* RIGHT - Stats Sidebar (Fluid Expandable) */}
                        <Grid item xs={12} lg={isTopExpenseExpanded ? 6 : 2.4} sx={{ pr: { lg: 5 }, transition: 'all 0.4s ease-in-out' }}>
                            <Stack spacing={4} sx={{ height: '100%', width: '100%' }}>
                                {/* Top Expense */}
                                <Paper
                                    sx={{
                                        p: 3,
                                        borderRadius: '12px',
                                        flexGrow: 1,
                                        display: 'flex',
                                        flexDirection: 'column',
                                        width: '100%',
                                        boxShadow: '0 4px 12px rgba(0,0,0,0.03)'
                                    }}
                                >
                                    <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 2 }}>
                                        <Typography variant="h6" sx={{ fontWeight: 800, color: '#333', flexGrow: 1 }}>
                                            Top expense
                                        </Typography>
                                        <IconButton size="small" onClick={() => setIsTopExpenseExpanded(!isTopExpenseExpanded)} sx={{ bgcolor: '#4d6eff', color: 'white', '&:hover': { bgcolor: '#3d59cc' } }}>
                                            {isTopExpenseExpanded ? <ChevronRightIcon sx={{ fontSize: 20 }} /> : <ChevronLeftIcon sx={{ fontSize: 20 }} />}
                                        </IconButton>
                                        <IconButton size="small" sx={{ bgcolor: '#4d6eff', color: 'white', '&:hover': { bgcolor: '#3d59cc' } }}>
                                            <DownloadIcon sx={{ fontSize: 20 }} />
                                        </IconButton>
                                    </Stack>
                                    <Box sx={{ flexGrow: 1, width: '100%', position: 'relative', minHeight: 280, display: 'flex', gap: 2 }}>
                                        <Box sx={{ flex: 1 }}>
                                            <DonutChart data={summary.topExpenses} showLegend={!isTopExpenseExpanded} />
                                        </Box>
                                        {isTopExpenseExpanded && (
                                            <Box sx={{ flex: 1, overflowY: 'auto', maxHeight: 280, pr: 1 }}>
                                                <TableContainer>
                                                    <Table size="small" stickyHeader>
                                                        <TableHead>
                                                            <TableRow>
                                                                <TableCell sx={{ fontWeight: 800, color: '#64748b' }}>Account Name</TableCell>
                                                                <TableCell align="right" sx={{ fontWeight: 800, color: '#64748b' }}>Amount (THB)</TableCell>
                                                            </TableRow>
                                                        </TableHead>
                                                        <TableBody>
                                                            {summary.topExpenses.map((expense, index) => (
                                                                <TableRow key={index} hover sx={{ '&:last-child td, &:last-child th': { border: 0 } }}>
                                                                    <TableCell component="th" scope="row" sx={{ color: '#334155', fontWeight: 600 }}>
                                                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                                                            <Box sx={{ flexShrink: 0, width: 12, height: 12, borderRadius: '50%', bgcolor: COLORS[index % COLORS.length] }} />
                                                                            <span>{expense.name || 'Unknown Account'}</span>
                                                                        </Box>
                                                                    </TableCell>
                                                                    <TableCell align="right" sx={{ color: '#0f172a', fontWeight: 700 }}>
                                                                        {new Intl.NumberFormat('en-US', { minimumFractionDigits: 2 }).format(expense.value)}
                                                                    </TableCell>
                                                                </TableRow>
                                                            ))}
                                                            {summary.topExpenses.length === 0 && (
                                                                <TableRow>
                                                                    <TableCell colSpan={2} align="center" sx={{ py: 3, color: '#94a3b8' }}>
                                                                        No expenses recorded
                                                                    </TableCell>
                                                                </TableRow>
                                                            )}
                                                        </TableBody>
                                                    </Table>
                                                </TableContainer>
                                            </Box>
                                        )}
                                    </Box>
                                </Paper>

                                {/* CAPEX Budget */}
                                <Paper
                                    sx={{
                                        p: 2.5,
                                        borderRadius: '12px',
                                        background: 'linear-gradient(135deg, #4d6eff 0%, #3d59cc 100%)', // Blue shade
                                        color: 'white',
                                        width: '100%',
                                        boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
                                        position: 'relative',
                                        overflow: 'hidden'
                                    }}
                                >
                                    <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1, position: 'relative', zIndex: 1 }}>
                                        <Stack direction="row" spacing={1} alignItems="center">
                                            <Box sx={{ bgcolor: 'rgba(255,255,255,0.2)', p: 0.5, borderRadius: '4px', display: 'flex' }}>
                                                <SyncIcon sx={{ fontSize: 18 }} />
                                            </Box>
                                            <Typography variant="body1" sx={{ fontWeight: 800, color: 'white' }}>
                                                CAPEX Budget
                                            </Typography>
                                        </Stack>
                                        <IconButton size="small" sx={{ color: 'white' }}>
                                            <DownloadIcon sx={{ fontSize: 22 }} />
                                        </IconButton>
                                    </Stack>

                                    <Box sx={{ position: 'relative', zIndex: 1 }}>
                                        <Typography variant="h4" sx={{ fontWeight: 800, letterSpacing: '-0.01em', color: 'white' }}>
                                            {(summary.capexActual / 1000000).toFixed(2)} MB
                                        </Typography>
                                        <Box sx={{ textAlign: 'right', mt: 1 }}>
                                            <Typography variant="body2" sx={{ fontWeight: 700, opacity: 0.9, color: 'white' }}>
                                                of {(summary.capexBudget / 1000000).toFixed(2)} MB
                                            </Typography>
                                        </Box>
                                    </Box>
                                </Paper>
                            </Stack>
                        </Grid>
                    </Grid>
                </Box>
            </Box >
        </ErrorBoundary >
    );
};

export default OwnerDashboard;