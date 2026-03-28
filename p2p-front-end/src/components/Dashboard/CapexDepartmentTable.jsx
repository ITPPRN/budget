import React from 'react';
import { Paper, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Box, TablePagination, TableSortLabel, IconButton } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import FlagIcon from '@mui/icons-material/Flag';

const CapexDepartmentTable = ({ data, count, page, rowsPerPage, onPageChange, onRowsPerPageChange, orderBy, order, onRequestSort, selectedDept, onRowClick, onDownload, onSettings, thresholds }) => {
    // Helper to format numbers (Truncate MB to 2 decimals)
    const formatMoney = (amount) => {
        if (!amount) return "0.00 MB";
        const mb = amount / 1000000;
        // Truncate to 2 decimal places (no rounding)
        const truncated = Math.floor(mb * 100) / 100;
        return `${truncated.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} MB`;
    };

    // Helper to determine status color based on usage
    const getStatusColor = (budget, used) => {
        const redLimit = Number(thresholds?.red) || 100;
        const yellowLimit = Number(thresholds?.yellow) || 80;

        if (budget === 0) {
            if (used > 0) return '#e74a3b'; // Red
            return '#1cc88a'; // Green
        }
        const spendPct = (used / budget) * 100;

        if (spendPct >= redLimit) return '#e74a3b'; // Red
        if (spendPct >= yellowLimit) return '#f6c23e'; // Yellow
        return '#1cc88a'; // Green
    };

    const createSortHandler = (property) => (event) => {
        onRequestSort(property);
    };

    return (
        <Paper sx={{ p: 2, height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                <Typography variant="h6" sx={{ fontWeight: 'bold', color: 'secondary.main', bgcolor: '#f3e5f5', px: 2, py: 0.5, borderRadius: 2 }}>
                    Capex Department Status
                </Typography>
                <Box sx={{ display: 'flex', gap: 0.5 }}>
                    <IconButton size="small" onClick={onSettings} sx={{ color: 'text.secondary' }}>
                        <Typography sx={{ fontSize: '10px', mr: 0.5, fontWeight: 'bold' }}>SET</Typography>
                    </IconButton>
                    <IconButton size="small" onClick={onDownload} sx={{ color: 'primary.main' }} disabled={!onDownload || !data || data.length === 0}>
                        <DownloadIcon sx={{ fontSize: 20 }} />
                    </IconButton>
                </Box>
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
                                    Capex_BG
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
                            <TableCell align="right" sortDirection={orderBy === 'actual'} sx={{ fontWeight: 'bold', bgcolor: '#eaecf4', width: '10%', minWidth: '60px', padding: '6px 4px' }}>
                                <TableSortLabel
                                    active={orderBy === 'actual'}
                                    direction={orderBy === 'actual' ? order : 'asc'}
                                    onClick={createSortHandler('actual')}
                                >
                                    %spend
                                </TableSortLabel>
                            </TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {(data || []).map((row) => {
                            const budget = row.budget || 0;
                            const spending = row.actual || 0; // Note: using row.actual from backend DTO
                            const remaining = budget - spending;
                            const remainingPct = budget > 0 ? (remaining / budget) * 100 : 0;
                            const statusColor = getStatusColor(budget, spending);
                            const isSelected = selectedDept === row.department; // Note: using row.department from backend DTO

                            return (
                                <TableRow
                                    key={row.department}
                                    hover
                                    onClick={() => onRowClick && onRowClick(row.department)}
                                    sx={{
                                        cursor: 'pointer',
                                        bgcolor: isSelected ? 'rgba(156, 39, 176, 0.08) !important' : 'inherit'
                                    }}
                                >
                                    <TableCell>
                                        <FlagIcon sx={{ color: statusColor, fontSize: '1rem' }} />
                                    </TableCell>
                                    <TableCell sx={{ fontWeight: 'bold', fontSize: '0.75rem' }}>{row.department}</TableCell>
                                    <TableCell align="right" sx={{ fontSize: '0.75rem' }}>{formatMoney(budget)}</TableCell>
                                    <TableCell align="right" sx={{ fontSize: '0.75rem' }}>{formatMoney(spending)}</TableCell>
                                    <TableCell align="right" sx={{ color: remaining < 0 ? 'error.main' : 'success.main', fontWeight: 'bold', fontSize: '0.75rem' }}>
                                        {formatMoney(remaining)}
                                    </TableCell>
                                    <TableCell align="right" sx={{ color: statusColor, fontWeight: 'bold', fontSize: '0.75rem' }}>
                                        {budget > 0 ? (spending / budget * 100).toFixed(2) : (spending > 0 ? '100.00' : '0.00')}%
                                    </TableCell>
                                </TableRow>
                            );
                        })}
                    </TableBody>
                </Table>
            </TableContainer>

            <Box sx={{ flexShrink: 0 }}>
                <TablePagination
                    rowsPerPageOptions={[10, 20, 50, 100]}
                    component="div"
                    count={count || 0}
                    rowsPerPage={rowsPerPage}
                    page={page}
                    onPageChange={onPageChange}
                    onRowsPerPageChange={onRowsPerPageChange}
                />
            </Box>
        </Paper >
    );
};

export default CapexDepartmentTable;
