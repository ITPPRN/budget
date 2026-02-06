import React from 'react';
import { Paper, Typography, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Box, TablePagination, TableSortLabel } from '@mui/material';
import FlagIcon from '@mui/icons-material/Flag';

const DepartmentTable = ({ data, count, page, rowsPerPage, onPageChange, onRowsPerPageChange, orderBy, order, onRequestSort, selectedDept, onRowClick }) => {
    // Helper to format numbers (Always in MB)
    const formatMoney = (amount) => {
        const mb = amount / 1000000;
        return `${mb.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} MB`;
    };

    // Helper to determine status color based on usage
    const getStatusColor = (budget, used) => {
        if (budget === 0) return 'text.secondary';
        const ratio = used / budget;
        if (ratio > 1) return '#e74a3b'; // Red (Over)
        if (ratio > 0.8) return '#f6c23e'; // Yellow (Warning)
        return '#1cc88a'; // Green (Good)
    };

    const createSortHandler = (property) => (event) => {
        onRequestSort(property);
    };

    return (
        <Paper sx={{ p: 2, height: '100%', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                <Typography variant="h6" sx={{ fontWeight: 'bold', color: 'primary.main', bgcolor: '#e3f2fd', px: 2, py: 0.5, borderRadius: 2 }}>
                    Department Budget Status (Top Spenders)
                </Typography>
            </Box>

            <TableContainer sx={{ flexGrow: 1 }}>
                <Table stickyHeader size="small">
                    <TableHead>
                        <TableRow>
                            <TableCell sx={{ fontWeight: 'bold', bgcolor: '#eaecf4' }}>Status</TableCell>
                            <TableCell sx={{ fontWeight: 'bold', bgcolor: '#eaecf4' }}>Department</TableCell>

                            <TableCell align="right" sortDirection={orderBy === 'budget' ? order : false} sx={{ fontWeight: 'bold', bgcolor: '#eaecf4' }}>
                                <TableSortLabel
                                    active={orderBy === 'budget'}
                                    direction={orderBy === 'budget' ? order : 'asc'}
                                    onClick={createSortHandler('budget')}
                                >
                                    Budget
                                </TableSortLabel>
                            </TableCell>

                            <TableCell align="right" sortDirection={orderBy === 'actual' ? order : false} sx={{ fontWeight: 'bold', bgcolor: '#eaecf4' }}>
                                <TableSortLabel
                                    active={orderBy === 'actual'}
                                    direction={orderBy === 'actual' ? order : 'asc'}
                                    onClick={createSortHandler('actual')}
                                >
                                    Spend
                                </TableSortLabel>
                            </TableCell>

                            <TableCell align="right" sortDirection={orderBy === 'remaining' ? order : false} sx={{ fontWeight: 'bold', bgcolor: '#eaecf4' }}>
                                <TableSortLabel
                                    active={orderBy === 'remaining'}
                                    direction={orderBy === 'remaining' ? order : 'asc'}
                                    onClick={createSortHandler('remaining')}
                                >
                                    Remaining
                                </TableSortLabel>
                            </TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {data.map((row) => {
                            const spending = row.spending || 0;
                            const remaining = row.budget - spending;
                            const statusColor = getStatusColor(row.budget, spending);
                            const isSelected = selectedDept === row.deptRaw; // Check selection

                            return (
                                <TableRow
                                    key={row.name}
                                    hover
                                    onClick={() => onRowClick && onRowClick(row.deptRaw)}
                                    selected={isSelected} // MUI Selected Style
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
                                </TableRow>
                            );
                        })}
                    </TableBody>
                </Table>
            </TableContainer>

            <TablePagination
                component="div"
                count={count || 0}
                page={page || 0}
                onPageChange={onPageChange}
                rowsPerPage={rowsPerPage || 10}
                onRowsPerPageChange={onRowsPerPageChange}
                rowsPerPageOptions={[10, 20, 50, 100]}
            />
        </Paper >
    );
};

export default DepartmentTable;
