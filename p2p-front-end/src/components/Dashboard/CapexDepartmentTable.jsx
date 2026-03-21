import React from 'react';
import { Paper, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Box, TablePagination, TableSortLabel, IconButton } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import FlagIcon from '@mui/icons-material/Flag';

const CapexDepartmentTable = ({ data, count, page, rowsPerPage, onPageChange, onRowsPerPageChange, orderBy, order, onRequestSort, selectedDept, onRowClick, onDownload }) => {
    // Helper to format numbers (Always in MB)
    const formatMoney = (amount) => {
        const mb = amount / 1000000;
        return `${mb.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} MB`;
    };

    // Helper to determine status color based on usage
    const getStatusColor = (budget, used) => {
        if (budget === 0) {
            if (used > 0) return '#e74a3b';
            return 'text.secondary';
        }
        const remaining = budget - used;
        const remainingPercentage = (remaining / budget) * 100;

        if (remaining < 0) return '#e74a3b';
        if (remainingPercentage <= 20) return '#f6c23e';
        return '#1cc88a';
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
                <IconButton size="small" onClick={onDownload} sx={{ color: 'primary.main' }} disabled={!onDownload || !data || data.length === 0}>
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
                            <TableCell align="right" sortDirection={orderBy === 'remaining_pct' ? order : false} sx={{ fontWeight: 'bold', bgcolor: '#eaecf4', width: '10%', minWidth: '60px', padding: '6px 4px' }}>
                                <TableSortLabel
                                    active={orderBy === 'remaining_pct'}
                                    direction={orderBy === 'remaining_pct' ? order : 'asc'}
                                    onClick={createSortHandler('remaining_pct')}
                                >
                                    %
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
                                    <TableCell align="right" sx={{ color: remaining < 0 ? 'error.main' : 'success.main', fontWeight: 'bold', fontSize: '0.75rem' }}>
                                        {remainingPct.toFixed(2)}%
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
