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
    TableBody
} from '@mui/material';

const ActualTable = () => {
    return (
        <Paper sx={{
            p: 2,
            borderRadius: 2,
            height: '300px', // Fixed height for consistency
            flexShrink: 0, // Prevent shrinking
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden'
        }}>
            <Typography variant="h6" sx={{ color: 'white', bgcolor: '#ff9800', p: 1, borderRadius: 1, mb: 2, flexShrink: 0 }}>
                Actual Detail
            </Typography>
            <Box sx={{ flexGrow: 1, overflow: 'hidden', width: '100%', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
                <TableContainer sx={{ flexGrow: 1, width: '100%', overflowX: 'auto', overflowY: 'auto' }}>
                    <Table stickyHeader size="small" sx={{ minWidth: '100%' }}>
                        <TableHead>
                            <TableRow>
                                <TableCell sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold', whiteSpace: 'nowrap' }}>GL Code</TableCell>
                                <TableCell sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold', minWidth: 200 }}>GL Name</TableCell>
                                <TableCell sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold' }}>Document No.</TableCell>
                                <TableCell align="right" sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold' }}>Amount</TableCell>
                                <TableCell sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold' }}>Vendor</TableCell>
                                <TableCell sx={{ bgcolor: '#4e73df', color: 'white', fontWeight: 'bold', minWidth: 150 }}>Description</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            <TableRow>
                                <TableCell colSpan={6} align="center" sx={{ py: 5, color: 'text.secondary' }}>
                                    Coming Soon...
                                </TableCell>
                            </TableRow>
                        </TableBody>
                    </Table>
                </TableContainer>
            </Box>
        </Paper>
    );
};

export default ActualTable;
