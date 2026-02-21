import React, { useState, useEffect } from 'react';
import { Box, Grid, Typography, FormControl, Select, MenuItem, Stack, Paper, CircularProgress, Button, IconButton, Tooltip, LinearProgress } from '@mui/material';
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
import { useAuth } from '../../hooks/useAuth';
import api from '../../utils/api/axiosInstance';
import BudgetChart from '../../components/Dashboard/OwnerBudgetChart';
import DonutChart from '../../components/Dashboard/DonutChart';
import ErrorBoundary from '../../components/ErrorBoundary';
import { toast } from 'react-toastify';

// Calculate formatter
const formatCurrency = (value) => {
    // If value is 0 or null/undefined, return "0"
    if (!value) return "0";
    if (value >= 1000000) return `${(value / 1000000).toFixed(2)} MB`;
    if (value >= 1000) return `${(value / 1000).toFixed(2)} K`;
    return value.toLocaleString();
};

const MetricCard = ({ title, value, color, icon: Icon, subTitle, trend }) => (
    <Paper sx={{
        p: 3, borderRadius: '24px', height: '100%',
        bgcolor: '#ffffff',
        boxShadow: '0 10px 40px rgba(0,0,0,0.05)',
        display: 'flex', flexDirection: 'column', justifyContent: 'space-between',
        transition: 'all 0.3s ease',
        '&:hover': { transform: 'translateY(-5px)', boxShadow: '0 15px 50px rgba(0,0,0,0.1)' }
    }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <Box>
                <Typography variant="body2" sx={{ color: '#90a4ae', fontWeight: 'bold', mb: 1, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                    {title}
                </Typography>
                <Typography variant="h4" sx={{ fontWeight: '800', color: '#263238', letterSpacing: '-0.02em' }}>
                    {formatCurrency(value)}
                </Typography>
            </Box>
            <Box sx={{
                width: 48, height: 48, borderRadius: '16px',
                background: `linear-gradient(135deg, ${color} 0%, ${color}aa 100%)`,
                color: 'white',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                boxShadow: `0 8px 16px ${color}40`
            }}>
                <Icon fontSize="medium" />
            </Box>
        </Box>
        <Box sx={{ mt: 3, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <Typography variant="caption" sx={{ color: '#607d8b', fontWeight: 600, bgcolor: '#f5f7f9', px: 1.5, py: 0.5, borderRadius: '8px' }}>
                {subTitle || 'Fiscal Year'}
            </Typography>
            {trend && (
                <Typography variant="caption" sx={{ color: trend > 0 ? '#ef5350' : '#4caf50', fontWeight: 'bold' }}>
                    {trend > 0 ? '+' : ''}{trend}%
                </Typography>
            )}
        </Box>
    </Paper>
);

const UsageCard = ({ usagePercent, statusLabel }) => {
    const statusColor = statusLabel === 'In Budget' ? '#00e676' : (statusLabel === 'Near Limit' ? '#ff9100' : '#ff1744');

    return (
        <Paper sx={{
            p: 3, borderRadius: '24px', height: '100%',
            bgcolor: '#ffffff',
            boxShadow: '0 10px 40px rgba(0,0,0,0.05)',
            display: 'flex', flexDirection: 'column', justifyContent: 'center'
        }}>
            <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
                <Typography variant="body2" sx={{ color: '#90a4ae', fontWeight: 'bold', textTransform: 'uppercase' }}>
                    Budget Usage
                </Typography>
                <Box sx={{ px: 1.5, py: 0.5, bgcolor: `${statusColor}15`, borderRadius: '8px', color: statusColor }}>
                    <Typography variant="caption" fontWeight="bold">{statusLabel}</Typography>
                </Box>
            </Stack>

            <Box sx={{ position: 'relative', display: 'flex', justifyContent: 'center', mb: 1 }}>
                <CircularProgress
                    variant="determinate"
                    value={100}
                    size={100}
                    thickness={4}
                    sx={{ color: '#f0f2f5', position: 'absolute' }}
                />
                <CircularProgress
                    variant="determinate"
                    value={Math.min(usagePercent, 100)}
                    size={100}
                    thickness={4}
                    sx={{ color: statusColor, '& .MuiCircularProgress-circle': { strokeLinecap: 'round' } }}
                />
                <Box sx={{
                    top: 0, left: 0, bottom: 0, right: 0,
                    position: 'absolute', display: 'flex', alignItems: 'center', justifyContent: 'center',
                    flexDirection: 'column'
                }}>
                    <Typography variant="h6" component="div" color="text.primary" fontWeight="bold">
                        {usagePercent.toFixed(0)}%
                    </Typography>
                </Box>
            </Box>
            <Typography variant="caption" align="center" sx={{ color: '#607d8b', mt: 1 }}>
                of Total Budget
            </Typography>
        </Paper>
    );
};

const OwnerDashboard = () => {
    const { user } = useAuth();
    const [loading, setLoading] = useState(true);
    const [chartMode, setChartMode] = useState('monthly'); // 'monthly' | 'accumulated'

    // Filter Options State
    const [selectedCompany, setSelectedCompany] = useState('');
    const [selectedBranch, setSelectedBranch] = useState('');
    const [selectedYear, setSelectedYear] = useState('');

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

    // Derived Branches based on Entity Selection
    const availableBranches = React.useMemo(() => {
        if (!selectedCompany || selectedCompany === 'All' || !Array.isArray(orgStructure)) return [];
        const entityObj = orgStructure.find(o => o.entity === selectedCompany);
        return entityObj ? entityObj.branches : [];
    }, [selectedCompany, orgStructure]);

    // 2. Fetch Dashboard Data
    const fetchDashboardData = async () => {
        if (!selectedYear) return;
        setLoading(true);
        try {
            const payload = {
                year: selectedYear,
                entities: selectedCompany && selectedCompany !== 'All' ? [selectedCompany] : [],
                branches: selectedBranch && selectedBranch !== 'All' ? [selectedBranch] : []
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
                    value: item.amount
                })),
                overBudgetCount: data.over_budget_count || 0,
                nearLimitCount: data.near_limit_count || 0
            });

        } catch (err) {
            console.error("Owner Dashboard Fetch Error", err);
            // toast.error("Failed to load dashboard data"); 
        } finally {
            setLoading(false);
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

            // Step 1: Auto Sync
            try {
                console.log("1. Starting Auto-Sync...");
                await api.post('/owner/sync-actuals');
                console.log("2. Auto-Sync Completed.");
            } catch (err) {
                console.error("Auto-Sync Failed (continuing to filters)", err);
                toast.warning("Auto-sync failed, showing cached data.");
            }

            // Step 2: Fetch Filters (After Sync to ensure we get new years if any)
            try {
                console.log("3. Fetching Filters...");
                const structRes = await api.get('/budgets/organization-structure');
                setOrgStructure(structRes.data || []);

                const listRes = await api.get('/owner/filter-lists');
                const years = listRes.data.years || [];
                setFilterYears(years);
                console.log("4. Filters Fetched. Years:", years);

                if (years.length > 0) {
                    // Sort descending
                    const sortedYears = [...years].sort((a, b) => b.localeCompare(a));
                    const defaultYear = sortedYears[0];

                    // Select Year (this will trigger the Data Fetch Effect)
                    if (selectedYear !== defaultYear) {
                        console.log(`5. Setting Year to ${defaultYear}`);
                        setSelectedYear(defaultYear);
                    } else {
                        // If year is SAME as before (e.g. re-mount), we MUST force a re-fetch 
                        // because the data might have changed due to Sync.
                        console.log("5. Year unchanged, forcing data refresh...");
                        fetchDashboardData();
                    }
                } else {
                    console.warn("No years found.");
                    setLoading(false); // Stop loading if no data
                }

            } catch (err) {
                console.error("Filter Fetch Error", err);
                toast.error("Failed to load filters");
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
    }, [selectedYear, selectedCompany, selectedBranch]);


    const usagePercent = summary.totalBudget > 0
        ? (summary.totalActual / summary.totalBudget) * 100
        : (summary.totalActual > 0 ? 100 : 0);

    const statusLabel =
        summary.totalBudget === 0 && summary.totalActual > 0 ? 'Over Budget' :
            usagePercent > 100 ? 'Over Budget' :
                usagePercent > 80 ? 'Near Limit' : 'In Budget';

    if (loading && summary.totalBudget === 0 && !selectedYear) {
        return (<Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}> <CircularProgress /> </Box>);
    }

    return (
        <ErrorBoundary>
            <Box sx={{ p: 4, minHeight: '100vh', bgcolor: '#f4f6f8' }}>

                {/* 1. Header & Filters */}
                <Stack direction={{ xs: 'column', md: 'row' }} justifyContent="space-between" alignItems="center" sx={{ mb: 4 }}>
                    <Box>
                        <Typography variant="h4" sx={{ fontWeight: '800', color: '#1a237e', letterSpacing: '-0.02em', mb: 1 }}>
                            Dashboard
                        </Typography>
                        <Typography variant="subtitle1" sx={{ color: '#546e7a', fontWeight: 500 }}>
                            Welcome back, {user?.name || 'Owner'}
                        </Typography>
                    </Box>

                    <Paper sx={{ p: 1, borderRadius: '16px', display: 'flex', gap: 2, alignItems: 'center', bgcolor: 'white', boxShadow: '0 4px 12px rgba(0,0,0,0.03)' }}>
                        <FormControl size="small" sx={{ minWidth: 150 }}>
                            <Select
                                value={selectedCompany}
                                onChange={(e) => { setSelectedCompany(e.target.value); setSelectedBranch(''); }}
                                displayEmpty
                                variant="standard"
                                disableUnderline
                                sx={{ px: 2, py: 0.5, borderRadius: '8px', bgcolor: 'transparent', fontWeight: 600, color: '#37474f' }}
                            >
                                <MenuItem value="">All Entities</MenuItem>
                                {orgStructure.map((org) => <MenuItem key={org.entity} value={org.entity}>{org.entity}</MenuItem>)}
                            </Select>
                        </FormControl>
                        <Box sx={{ width: 1, height: 24, bgcolor: '#cfd8dc' }} /> {/* Divider */}
                        <FormControl size="small" sx={{ minWidth: 150 }} disabled={!selectedCompany}>
                            <Select
                                value={selectedBranch}
                                onChange={(e) => setSelectedBranch(e.target.value)}
                                displayEmpty
                                variant="standard"
                                disableUnderline
                                sx={{ px: 2, py: 0.5, borderRadius: '8px', bgcolor: 'transparent', fontWeight: 600, color: '#37474f' }}
                            >
                                <MenuItem value="">All Branches</MenuItem>
                                {availableBranches.map((b) => <MenuItem key={b.name} value={b.name}>{b.name}</MenuItem>)}
                            </Select>
                        </FormControl>
                        <Box sx={{ width: 1, height: 24, bgcolor: '#cfd8dc' }} /> {/* Divider */}
                        <FormControl size="small" sx={{ minWidth: 100 }}>
                            <Select
                                value={selectedYear}
                                onChange={(e) => setSelectedYear(e.target.value)}
                                displayEmpty
                                variant="standard"
                                disableUnderline
                                sx={{ px: 2, py: 0.5, borderRadius: '8px', bgcolor: 'transparent', fontWeight: 600, color: '#37474f' }}
                            >
                                <MenuItem value="" disabled>Year</MenuItem>
                                {filterYears.map((y) => <MenuItem key={y} value={y}>{y}</MenuItem>)}
                            </Select>
                        </FormControl>
                    </Paper>
                </Stack>

                {/* 2. Stats Cards */}
                <Grid container spacing={3} sx={{ mb: 4 }}>
                    <Grid item xs={12} sm={6} md={3}>
                        <MetricCard
                            title="Approved Budget"
                            value={summary.totalBudget}
                            color="#2979ff" // Blue
                            icon={AccountBalanceWalletIcon}
                        />
                    </Grid>
                    <Grid item xs={12} sm={6} md={3}>
                        <MetricCard
                            title="Actual Spending"
                            value={summary.totalActual}
                            color="#ff4081" // Pink
                            icon={ShoppingBagIcon}
                            subTitle={`${usagePercent.toFixed(1)}% Used`}
                        />
                    </Grid>
                    <Grid item xs={12} sm={6} md={3}>
                        <MetricCard
                            title="Remaining"
                            value={summary.totalBudget - summary.totalActual}
                            color="#00e5ff" // Cyan
                            icon={PieChartIcon}
                            subTitle="Available"
                        />
                    </Grid>
                    <Grid item xs={12} sm={6} md={3}>
                        <UsageCard usagePercent={usagePercent} statusLabel={statusLabel} />
                    </Grid>
                </Grid>

                {/* 3. Charts Section */}
                <Grid container spacing={3}>
                    {/* Left: Line Chart */}
                    <Grid item xs={12} lg={8}>
                        <Paper sx={{ p: 4, borderRadius: '24px', boxShadow: '0 10px 40px rgba(0,0,0,0.05)', height: 500 }}>
                            <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 4 }}>
                                <Box>
                                    <Typography variant="h6" sx={{ fontWeight: '800', color: '#263238' }}>Budget Performance</Typography>
                                    <Typography variant="caption" sx={{ color: '#90a4ae' }}>Monthly breakdown of expenses</Typography>
                                </Box>

                                <Paper sx={{ bgcolor: '#eceff1', borderRadius: '12px', p: 0.5, display: 'flex' }}>
                                    <Button
                                        size="small"
                                        onClick={() => setChartMode('monthly')}
                                        sx={{
                                            bgcolor: chartMode === 'monthly' ? 'white' : 'transparent',
                                            color: chartMode === 'monthly' ? '#1a237e' : '#78909c',
                                            borderRadius: '8px',
                                            fontWeight: 'bold',
                                            boxShadow: chartMode === 'monthly' ? '0 2px 8px rgba(0,0,0,0.05)' : 'none',
                                            textTransform: 'none', px: 2
                                        }}
                                    >
                                        Monthly
                                    </Button>
                                    <Button
                                        size="small"
                                        onClick={() => setChartMode('accumulated')}
                                        sx={{
                                            bgcolor: chartMode === 'accumulated' ? 'white' : 'transparent',
                                            color: chartMode === 'accumulated' ? '#1a237e' : '#78909c',
                                            borderRadius: '8px',
                                            fontWeight: 'bold',
                                            boxShadow: chartMode === 'accumulated' ? '0 2px 8px rgba(0,0,0,0.05)' : 'none',
                                            textTransform: 'none', px: 2
                                        }}
                                    >
                                        Accumulated
                                    </Button>
                                </Paper>
                            </Stack>

                            <Box sx={{ height: '85%', width: '100%' }}>
                                <BudgetChart data={summary.chartData} />
                            </Box>
                        </Paper>
                    </Grid>

                    {/* Right: Donut + Capex */}
                    <Grid item xs={12} lg={4}>
                        <Stack spacing={3} sx={{ height: '100%' }}>
                            {/* Top Expense */}
                            <Paper sx={{ p: 4, borderRadius: '24px', boxShadow: '0 10px 40px rgba(0,0,0,0.05)', flex: 1, minHeight: 400 }}>
                                <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 3 }}>
                                    <Box>
                                        <Typography variant="h6" sx={{ fontWeight: '800', color: '#263238' }}>Top Expenses</Typography>
                                        <Typography variant="caption" sx={{ color: '#90a4ae' }}>By Account Category</Typography>
                                    </Box>
                                    <IconButton size="small" sx={{ color: '#cfd8dc' }}><MoreHorizIcon /></IconButton>
                                </Stack>
                                <Box sx={{ height: 280, width: '100%', position: 'relative' }}>
                                    <DonutChart data={summary.topExpenses} />
                                </Box>
                            </Paper>

                            {/* CAPEX Card */}
                            <Paper sx={{
                                p: 3, borderRadius: '24px',
                                background: 'linear-gradient(135deg, #1a237e 0%, #0d47a1 100%)', // Deep Blue
                                color: 'white',
                                boxShadow: '0 10px 30px rgba(26, 35, 126, 0.4)'
                            }}>
                                <Stack direction="row" justifyContent="space-between" alignItems="center">
                                    <Box>
                                        <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
                                            <Paper sx={{ width: 24, height: 24, borderRadius: '8px', bgcolor: 'rgba(255,255,255,0.2)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                                                <TrendingUpIcon sx={{ fontSize: 16, color: 'white' }} />
                                            </Paper>
                                            <Typography variant="body2" fontWeight="bold" sx={{ opacity: 0.9 }}>CAPEX Status</Typography>
                                        </Stack>
                                        <Typography variant="h5" fontWeight="bold">{formatCurrency(summary.capexBudget)}</Typography>
                                        <Typography variant="caption" sx={{ opacity: 0.7, mt: 0.5, display: 'block' }}>
                                            Actual: {formatCurrency(summary.capexActual)}
                                        </Typography>
                                    </Box>
                                    <IconButton sx={{ color: 'white', bgcolor: 'rgba(255,255,255,0.1)' }}><ChevronRightIcon /></IconButton>
                                </Stack>
                            </Paper>
                        </Stack>
                    </Grid>
                </Grid>

            </Box>
        </ErrorBoundary>
    );
};

export default OwnerDashboard;
