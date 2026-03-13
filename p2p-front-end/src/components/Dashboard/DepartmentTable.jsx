import React from 'react';
import { Paper, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Box, TablePagination, TableSortLabel, IconButton, Button } from '@mui/material';
import FlagIcon from '@mui/icons-material/Flag';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import DownloadIcon from '@mui/icons-material/Download';

const DepartmentTable = ({ data, count, page, rowsPerPage, onPageChange, onRowsPerPageChange, orderBy, order, onRequestSort, selectedDept, onRowClick, onBack }) => {
    // Helper to format numbers (Always in MB)
    const formatMoney = (amount) => {
        const mb = amount / 1000000;
        return `${mb.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} MB`;
    };

    // Helper to determine status color based on usage
    const getStatusColor = (budget, used) => {
        // 1. Invalid/No Budget
        if (budget === 0) {
            // If spending occurs without budget -> Red (Critical)
            if (used > 0) return '#e74a3b';
            // No budget, no spend -> Grey (Inactive)
            return 'text.secondary';
        }

        const remaining = budget - used;
        const remainingPercentage = (remaining / budget) * 100;

        // 2. Red (Over Budget): Spend > Budget (Remaining < 0)
        if (remaining < 0) return '#e74a3b';

        // 3. Yellow (Warning): Remaining is Low (e.g. < 20% left)
        if (remainingPercentage <= 20) return '#f6c23e';

        // 4. Green (Healthy): Safe zone
        return '#1cc88a';
    };

    const createSortHandler = (property) => (event) => {
        onRequestSort(property);
    };

    return (
        <Paper sx={{ p: 2, height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    {selectedDept === 'None' && (
                        <IconButton size="small" onClick={onBack} sx={{ bgcolor: '#f5f5f5' }}>
                            <ArrowBackIcon fontSize="small" />
                        </IconButton>
                    )}
                    <Typography variant="h6" sx={{ fontWeight: 'bold', color: 'primary.main', bgcolor: '#e3f2fd', px: 2, py: 0.5, borderRadius: 2 }}>
                        {selectedDept ? `Department: ${selectedDept}` : 'Department Budget Status (Top Spenders)'}
                    </Typography>
                </Box>
                <IconButton size="small" sx={{ color: 'primary.main' }}>
                    <DownloadIcon sx={{ fontSize: 20 }} />
                </IconButton>
            </Box>

            <TableContainer sx={{ flexGrow: 1, overflow: 'auto' }}>
                <Table stickyHeader size="small">
                    <TableHead>
                        <TableRow>
                            <TableCell sx={{ fontWeight: 'bold', bgcolor: '#eaecf4', width: '30px', padding: '6px 4px' }}>Status</TableCell>
                            <TableCell sx={{ fontWeight: 'bold', bgcolor: '#eaecf4' }}>Department</TableCell>

                            <TableCell align="right" sortDirection={orderBy === 'budget' ? order : false} sx={{ fontWeight: 'bold', bgcolor: '#eaecf4', width: '15%' }}>
                                <TableSortLabel
                                    active={orderBy === 'budget'}
                                    direction={orderBy === 'budget' ? order : 'asc'}
                                    onClick={createSortHandler('budget')}
                                >
                                    Budget
                                </TableSortLabel>
                            </TableCell>

                            <TableCell align="right" sortDirection={orderBy === 'actual' ? order : false} sx={{ fontWeight: 'bold', bgcolor: '#eaecf4', width: '15%' }}>
                                <TableSortLabel
                                    active={orderBy === 'actual'}
                                    direction={orderBy === 'actual' ? order : 'asc'}
                                    onClick={createSortHandler('actual')}
                                >
                                    Spend
                                </TableSortLabel>
                            </TableCell>

                            <TableCell align="right" sortDirection={orderBy === 'remaining' ? order : false} sx={{ fontWeight: 'bold', bgcolor: '#eaecf4', width: '15%' }}>
                                <TableSortLabel
                                    active={orderBy === 'remaining'}
                                    direction={orderBy === 'remaining' ? order : 'asc'}
                                    onClick={createSortHandler('remaining')}
                                >
                                    Remaining
                                </TableSortLabel>
                            </TableCell>
                            <TableCell align="right" sortDirection={orderBy === 'remaining_pct' ? order : false} sx={{ fontWeight: 'bold', bgcolor: '#eaecf4', width: '10%', minWidth: '60px', padding: '6px 4px' }}>
                                <TableSortLabel
                                    active={orderBy === 'remaining_pct'}
                                    direction={orderBy === 'remaining_pct' ? order : 'asc'}
                                    onClick={createSortHandler('remaining_pct')}
                                >
                                    ( % )
                                </TableSortLabel>
                            </TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {(data || []).map((row) => {
                            const budget = row.budget || 0;
                            const spending = row.spending || 0;
                            const remaining = budget - spending;
                            const remainingPct = budget > 0 ? (remaining / budget) * 100 : 0;
                            const statusColor = getStatusColor(budget, spending);

                            const isSelected = selectedDept === row.deptRaw;

                            return (
                                <TableRow
                                    key={row.name}
                                    hover
                                    onClick={() => onRowClick && onRowClick(row.deptRaw)}
                                    // Remove 'selected' prop to avoid default gray background
                                    sx={{
                                        cursor: 'pointer',
                                        bgcolor: isSelected ? 'rgba(25, 118, 210, 0.08) !important' : 'inherit'
                                    }}
                                >
                                    <TableCell>
                                        <FlagIcon sx={{ color: statusColor, fontSize: '1rem' }} />
                                    </TableCell>
                                    <TableCell sx={{ fontWeight: 'bold', fontSize: '0.75rem' }}>{row.name}</TableCell>
                                    <TableCell align="right" sx={{ fontSize: '0.75rem' }}>{formatMoney(row.budget)}</TableCell>
                                    <TableCell align="right" sx={{ fontSize: '0.75rem' }}>{formatMoney(spending)}</TableCell>
                                    <TableCell align="right" sx={{ color: remaining < 0 ? 'error.main' : 'success.main', fontWeight: 'bold', fontSize: '0.75rem' }}>
                                        {formatMoney(remaining)}
                                    </TableCell>
                                    <TableCell align="right" sx={{ color: remaining < 0 ? 'error.main' : 'success.main', fontWeight: 'bold', fontSize: '0.75rem' }}>
                                        ({remainingPct.toFixed(2)}%)
                                    </TableCell>
                                </TableRow>
                            );
                        })}
                    </TableBody>
                </Table>
            </TableContainer>

            <Box sx={{ flexShrink: 0 }}>
                <TablePagination
                    component="div"
                    count={count || 0}
                    page={page || 0}
                    onPageChange={onPageChange}
                    rowsPerPage={rowsPerPage || 10}
                    onRowsPerPageChange={onRowsPerPageChange}
                    rowsPerPageOptions={[10, 20, 50, 100]}
                />
            </Box>
        </Paper >
    );
};

export default DepartmentTable;
