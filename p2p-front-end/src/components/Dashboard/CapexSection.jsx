import React, { useState, useEffect } from 'react';
import { Box, Grid, Typography, Stack, FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import api from '../../utils/api/axiosInstance';
import { TotalBudgetCard, RemainingBudgetCard, DepartmentAlertCard } from '../Dashboard/BudgetStatsCard';
import CapexDepartmentTable from '../Dashboard/CapexDepartmentTable';
import BudgetChart from '../Dashboard/BudgetChart';
import { downloadExcelFile } from '../../utils/exportUtils';

const CapexSection = ({ globalEntity, orgStructure }) => {
    const [loading, setLoading] = useState(true);
    const [selectedEntity, setSelectedEntity] = useState(globalEntity || ''); // Local State

    // Sync with global if global changes (Optional, but good UX if user uses top filter)
    // User can then override it locally.
    useEffect(() => {
        if (globalEntity) {
            setSelectedEntity(globalEntity);
        }
    }, [globalEntity]);

    const [totalBudget, setTotalBudget] = useState(0);
    const [totalActual, setTotalActual] = useState(0);
    const [chartData, setChartData] = useState([]);
    const [departmentData, setDepartmentData] = useState([]);
    const [alertCounts, setAlertCounts] = useState({ over: 0, near: 0 });

    // Pagination & Sort State
    const [page, setPage] = useState(0);
    const [rowsPerPage, setRowsPerPage] = useState(10);
    const [totalCount, setTotalCount] = useState(0);
    const [order, setOrder] = useState('desc');
    const [orderBy, setOrderBy] = useState('actual');
    const [selectedDepartment, setSelectedDepartment] = useState('');

    useEffect(() => {
        const fetchCapexData = async () => {
            setLoading(true);
            try {
                const payload = {
                    entities: selectedEntity ? [selectedEntity] : [],
                    // Note: CAPEX ignores Branch filter as per requirement
                    departments: selectedDepartment ? [selectedDepartment] : [],
                    page: page + 1,
                    limit: rowsPerPage,
                    sort_by: orderBy,
                    sort_order: order
                };

                const { data } = await api.post('/capex/dashboard-summary', payload);

                setTotalBudget(data.total_budget || 0);
                setTotalActual(data.total_actual || 0);
                setTotalCount(data.total_count || 0);
                setAlertCounts({
                    over: data.over_budget_count || 0,
                    near: data.near_limit_count || 0
                });

                // Map Chart Data
                const mappedChart = (data.chart_data || []).map(item => ({
                    name: item.month,
                    budget: item.budget,
                    actual: item.actual
                }));
                setChartData(mappedChart);

                // Map Table Data (structure matches DTO from backend)
                setDepartmentData(data.department_data || []);

            } catch (err) {
                console.error("Capex Fetch Error", err);
            } finally {
                setLoading(false);
            }
        };

        fetchCapexData();
    }, [selectedEntity, selectedDepartment, page, rowsPerPage, orderBy, order]);

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
        setPage(0);
    };

    const handleDepartmentClick = (deptName) => {
        setSelectedDepartment(prev => prev === deptName ? '' : deptName);
        setPage(0);
    };

    // --- Export Handlers ---
    const handleCapexDeptStatusExport = async () => {
        const payload = {
            entities: selectedEntity ? [selectedEntity] : [],
            departments: selectedDepartment ? [selectedDepartment] : [],
            sort_by: orderBy,
            sort_order: order
        };
        
        await downloadExcelFile('/export-capex-department-status-admin', payload, `Capex_Dept_Status_Report.xlsx`);
    };

    const handleCapexVsActualExport = async () => {
        const payload = {
            entities: selectedEntity ? [selectedEntity] : [],
            departments: selectedDepartment ? [selectedDepartment] : [],
            sort_by: orderBy,
            sort_order: order
        };
        
        await downloadExcelFile('/export-capex-budget-vs-actual-admin', payload, `Capex_Vs_Actual_Report.xlsx`);
    };

    return (
        <Box sx={{ mt: 6, mb: 4, width: '100%', borderTop: '1px dashed #ccc', pt: 4 }}>
            {/* Header & Filter */}
            <Box sx={{ mb: 3, display: 'flex', flexDirection: { xs: 'column', sm: 'row' }, justifyContent: 'space-between', alignItems: 'flex-start', gap: 2 }}>
                <Box>
                    <Typography variant="h5" sx={{ fontWeight: 'bold', color: 'secondary.main' }}>
                        CAPEX Dashboard
                    </Typography>
                    <Typography variant="subtitle2" color="text.secondary">
                        ภาพรวมงบลงทุน ( CAPEX )
                    </Typography>
                </Box>

                {/* Local Entity Filter */}
                <FormControl size="small" sx={{ minWidth: 200, bgcolor: 'white', borderRadius: 1 }}>
                    <InputLabel color="secondary">Entity (บริษัท)</InputLabel>
                    <Select
                        value={selectedEntity}
                        label="Entity (บริษัท)"
                        onChange={(e) => {
                            setSelectedEntity(e.target.value);
                            setSelectedDepartment('');
                            setPage(0);
                        }}
                        color="secondary"
                    >
                        <MenuItem value=""><em>All Entities</em></MenuItem>
                        {(orgStructure || []).map((org) => (
                            <MenuItem key={org.entity} value={org.entity}>{org.entity}</MenuItem>
                        ))}
                    </Select>
                </FormControl>
            </Box>

            {/* Summary Cards */}
            <Grid container spacing={3} sx={{ mb: 4 }}>
                <Grid item xs={12} md={4}>
                    <TotalBudgetCard totalBudget={totalBudget} totalActual={totalActual} title="Total CAPEX Budget" />
                </Grid>
                <Grid item xs={12} md={4}>
                    <RemainingBudgetCard totalBudget={totalBudget} totalActual={totalActual} />
                </Grid>
                <Grid item xs={12} md={4}>
                    <DepartmentAlertCard overBudgetCount={alertCounts.over} nearLimitCount={alertCounts.near} />
                </Grid>
            </Grid>

            {/* Charts & Tables */}
            <Stack direction={{ xs: 'column', md: 'row' }} spacing={3} sx={{ width: '100%', mb: 4 }}>
                <Box sx={{ flex: 1, minWidth: 0, height: 500 }}>
                    <CapexDepartmentTable
                        data={departmentData}
                        count={totalCount}
                        page={page}
                        rowsPerPage={rowsPerPage}
                        onPageChange={handleChangePage}
                        onRowsPerPageChange={handleChangeRowsPerPage}
                        orderBy={orderBy}
                        order={order}
                        onRequestSort={handleRequestSort}
                        selectedDept={selectedDepartment}
                        onRowClick={handleDepartmentClick}
                        onDownload={handleCapexDeptStatusExport}
                    />
                </Box>
                <Box sx={{ flex: 1, minWidth: 0, height: 500 }}>
                    <BudgetChart
                        data={chartData}
                        title="CAPEX Budget vs Actual"
                        selectedDept={selectedDepartment || "ALL"}
                        onDownload={handleCapexVsActualExport}
                    />
                </Box>
            </Stack>
        </Box>
    );
};

export default CapexSection;
