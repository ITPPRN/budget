import React from 'react';
import {
    Paper,
    Typography,
    Box,
    TableContainer,
    Table,
    TableHead,
    TableRow,
    TableCell,
    TableBody,
    Skeleton
} from '@mui/material';

const months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];

const BudgetTable = ({ loading, data, selectedCount }) => {
    return (
        <Paper sx={{
            p: 2,
            display: 'flex',
            flexDirection: 'column',
            borderRadius: 2,
            mb: 2,
            flexGrow: 1,
            minHeight: { xs: '400px', md: 0 }, // Mobile gets fixed minHeight, Desktop fits available space
            overflow: 'hidden'
        }}>
            <Typography variant="h6" sx={{ color: 'white', bgcolor: '#1976d2', p: 1, borderRadius: 1, mb: 2, flexShrink: 0 }}>
                Budget Detail
            </Typography>

            <Box sx={{ flexGrow: 1, overflow: 'hidden', width: '100%', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
                <TableContainer sx={{ flexGrow: 1, width: '100%', overflowX: 'auto', overflowY: 'auto' }}>
                    <Table stickyHeader size="small" sx={{ minWidth: '100%' }}>
                        <TableHead>
                            <TableRow>
                                <TableCell sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold', whiteSpace: 'nowrap', position: 'sticky', left: 0, zIndex: 3 }}>GL Code</TableCell>
                                <TableCell sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold', minWidth: 200, whiteSpace: 'nowrap', position: 'sticky', left: 80, zIndex: 3 }}>GL Name</TableCell>
                                {months.map(m => (
                                    <TableCell key={m} align="right" sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold', whiteSpace: 'nowrap' }}>{m}</TableCell>
                                ))}
                                <TableCell align="right" sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold', whiteSpace: 'nowrap' }}>Total</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {loading ? (
                                // Skeleton Loading State
                                Array.from(new Array(5)).map((_, index) => (
                                    <TableRow key={index}>
                                        <TableCell sx={{ position: 'sticky', left: 0, bgcolor: 'background.paper', zIndex: 1 }}><Skeleton variant="text" width={80} /></TableCell>
                                        <TableCell sx={{ position: 'sticky', left: 80, bgcolor: 'background.paper', zIndex: 1 }}><Skeleton variant="text" width={150} /></TableCell>
                                        {months.map(m => (
                                            <TableCell key={m}><Skeleton variant="text" /></TableCell>
                                        ))}
                                        <TableCell><Skeleton variant="text" width={60} /></TableCell>
                                    </TableRow>
                                ))
                            ) : data.length > 0 ? (
                                data.map((row) => {
                                    const amountMap = {};
                                    row.budget_amounts?.forEach(a => {
                                        amountMap[a.month] = a.amount;
                                    });

                                    return (
                                        <TableRow key={row.id} hover>
                                            <TableCell sx={{ whiteSpace: 'nowrap', position: 'sticky', left: 0, bgcolor: 'background.paper', zIndex: 1 }}>{row.conso_gl || "-"}</TableCell>
                                            <TableCell sx={{ whiteSpace: 'nowrap', position: 'sticky', left: 80, bgcolor: 'background.paper', zIndex: 1 }}>{row.gl_name}</TableCell>
                                            {months.map(m => (
                                                <TableCell key={m} align="right" sx={{ whiteSpace: 'nowrap' }}>
                                                    {parseFloat(amountMap[m] || 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                                                </TableCell>
                                            ))}
                                            <TableCell align="right" sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', whiteSpace: 'nowrap' }}>
                                                {parseFloat(row.year_total || 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                                            </TableCell>
                                        </TableRow>
                                    );
                                })
                            ) : (
                                <TableRow>
                                    <TableCell colSpan={15} align="center" sx={{ py: 5, color: 'text.secondary' }}>
                                        {selectedCount === 0 ? "Select items from the Filter Pane to view details" : "No data found for selected filters"}
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </TableContainer>
            </Box>
        </Paper>
    );
};

export default BudgetTable;
