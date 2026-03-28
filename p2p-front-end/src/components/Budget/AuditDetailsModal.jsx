import React, { useState, useEffect } from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    Table,
    TableHead,
    TableRow,
    TableCell,
    TableBody,
    Typography,
    Box,
    TableContainer,
    CircularProgress
} from '@mui/material';
import VisibilityIcon from '@mui/icons-material/Visibility';
import api from '../../utils/api/axiosInstance';

const AuditDetailsModal = ({ open, onClose, log }) => {
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (open && log?.id) {
            const fetchItems = async () => {
                setLoading(true);
                try {
                    const res = await api.get(`/budgets/audit/logs/${log.id}/items`);
                    setItems(res.data || []);
                } catch (err) {
                    console.error("Fetch Rejected Items Error:", err);
                    setItems([]);
                } finally {
                    setLoading(false);
                }
            };
            fetchItems();
        } else {
            setItems([]);
        }
    }, [open, log]);

    if (!log) return null;

    return (
        <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
            <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1, bgcolor: '#043478', color: 'white' }}>
                <VisibilityIcon />
                <Typography variant="h6">Rejected Items Details</Typography>
            </DialogTitle>
            <DialogContent sx={{ p: 0 }}>
                <Box sx={{ p: 2, display: 'flex', justifyContent: 'space-between', bgcolor: '#f5f5f5' }}>
                    <Box>
                        <Typography variant="subtitle2" color="textSecondary">Department</Typography>
                        <Typography variant="body1" fontWeight="bold">{log.department_code || log.department}</Typography>
                    </Box>
                    <Box>
                        <Typography variant="subtitle2" color="textSecondary">Period</Typography>
                        <Typography variant="body1" fontWeight="bold">{log.month}/{log.year}</Typography>
                    </Box>
                    <Box>
                        <Typography variant="subtitle2" color="textSecondary">Owner</Typography>
                        <Typography variant="body1" fontWeight="bold">{log.owner_name}</Typography>
                    </Box>
                </Box>

                {loading ? (
                    <Box sx={{ p: 5, textAlign: 'center' }}>
                        <CircularProgress />
                    </Box>
                ) : (
                    <TableContainer sx={{ maxHeight: '60vh' }}>
                        <Table stickyHeader size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>GL Code</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Account Name</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Doc No.</TableCell>
                                    <TableCell align="right" sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Amount</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Vendor</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Description</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Date</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {items.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={7} align="center" sx={{ py: 5 }}>
                                            No rejected items found or snapshots not available.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    items.map((row, index) => (
                                        <TableRow key={index} hover>
                                            <TableCell>{row.conso_gl}</TableCell>
                                            <TableCell>{row.gl_account_name}</TableCell>
                                            <TableCell>{row.doc_no}</TableCell>
                                            <TableCell align="right" sx={{ fontWeight: 'bold', color: parseFloat(row.amount) < 0 ? 'red' : 'green' }}>
                                                {parseFloat(row.amount || 0).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                                            </TableCell>
                                            <TableCell>{row.vendor || "-"}</TableCell>
                                            <TableCell>{row.description}</TableCell>
                                            <TableCell>{row.posting_date}</TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </TableContainer>
                )}
            </DialogContent>
            <DialogActions sx={{ p: 2 }}>
                <Button onClick={onClose} variant="contained" color="primary">ปิดหน้าต่าง (Close)</Button>
            </DialogActions>
        </Dialog>
    );
};

export default AuditDetailsModal;
