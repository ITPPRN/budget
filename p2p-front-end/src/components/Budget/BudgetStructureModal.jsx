import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
    Dialog, DialogTitle, DialogContent, Typography, IconButton,
    Box, Grid, TextField, Button, Table, TableBody, TableCell,
    TableContainer, TableHead, TableRow, Paper, CircularProgress
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import DeleteIcon from '@mui/icons-material/Delete';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import AddIcon from '@mui/icons-material/Add';
import api from '../../utils/api/axiosInstance';
import { toast } from 'react-toastify';

// Memoized Table Component to prevent lag while typing
const StructureTable = React.memo(({ structures, loading, onDelete }) => (
    <TableContainer sx={{ flexGrow: 1, maxHeight: 500, overflow: 'auto' }}>
        <Table size="small" stickyHeader>
            <TableHead>
                <TableRow>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', width: 60 }}>ID</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5' }}>Group 1</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5' }}>Group 2</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5' }}>Group 3</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5' }}>Conso GL</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5' }}>Account Name</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', width: 60 }}>ลบ</TableCell>
                </TableRow>
            </TableHead>
            <TableBody>
                {loading ? (
                    <TableRow><TableCell colSpan={7} align="center"><CircularProgress size={30} sx={{ my: 4 }} /></TableCell></TableRow>
                ) : structures.map((row) => (
                    <TableRow key={row.id} hover>
                        <TableCell>{row.id}</TableCell>
                        <TableCell>{row.group1}</TableCell>
                        <TableCell>{row.group2}</TableCell>
                        <TableCell>{row.group3}</TableCell>
                        <TableCell>{row.conso_gl}</TableCell>
                        <TableCell sx={{ maxWidth: 150 }} noWrap title={row.account_name}>{row.account_name}</TableCell>
                        <TableCell>
                            <IconButton size="small" color="error" onClick={() => onDelete(row.id)}>
                                <DeleteIcon fontSize="small" />
                            </IconButton>
                        </TableCell>
                    </TableRow>
                ))}
                {!loading && structures.length === 0 && (
                    <TableRow><TableCell colSpan={7} align="center" sx={{ py: 4, color: '#999' }}>ไม่มีข้อมูล</TableCell></TableRow>
                )}
            </TableBody>
        </Table>
    </TableContainer>
));

const BudgetStructureModal = ({ open, onClose }) => {
    const [structures, setStructures] = useState([]);
    const [loading, setLoading] = useState(false);
    const fileInputRef = useRef(null);

    // Form State
    const [newRow, setNewRow] = useState({ group1: '', group2: '', group3: '', conso_gl: '', account_name: '' });

    const fetchStructures = useCallback(async () => {
        setLoading(true);
        try {
            const res = await api.get('/budgets/budget-structure-list');
            setStructures(res.data || []);
        } catch (error) {
            toast.error('Failed to load Budget Structures');
        } finally {
            setLoading(false);
        }
    }, []);

    // Reset state when modal closes
    useEffect(() => {
        if (!open) {
            setNewRow({ group1: '', group2: '', group3: '', conso_gl: '', account_name: '' });
        } else {
            fetchStructures();
        }
    }, [open, fetchStructures]);

    const handleAdd = async () => {
        if (!newRow.group1 || !newRow.group2 || !newRow.group3 || !newRow.conso_gl) {
            toast.warning('กรุณากรอกข้อมูล Group 1-3 และ Conso GL ให้ครบ');
            return;
        }

        const payload = {
            group1: newRow.group1,
            group2: newRow.group2,
            group3: newRow.group3,
            conso_gl: newRow.conso_gl,
            account_name: newRow.account_name
        };

        try {
            await api.post('/budgets/budget-structure', payload);
            toast.success('เพิ่มข้อมูลเรียบร้อย');
            setNewRow({ group1: '', group2: '', group3: '', conso_gl: '', account_name: '' });
            fetchStructures();
        } catch (error) {
            toast.error('ไม่สามารถเพิ่มข้อมูลได้');
        }
    };

    const handleDelete = useCallback(async (id) => {
        if (!window.confirm('ยืนยันการลบโครงสร้างนี้? ระบบอาจส่งผลกระทบต่อรายการที่อ้างอิงอยู่')) return;
        try {
            await api.delete(`/budgets/budget-structure/${id}`);
            toast.success('ลบข้อมูลเรียบร้อย');
            fetchStructures();
        } catch (error) {
            toast.error('ไม่สามารถลบข้อมูลได้');
        }
    }, [fetchStructures]);

    const handleFileSelect = async (e) => {
        // Obsolete
    };

    if (!open) return null;

    return (
        <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
            <DialogTitle sx={{ bgcolor: '#043478', color: 'white', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography sx={{ fontWeight: 'bold', color: 'white' }}>MANAGE FILTER PANE STRUCTURE (BUDGET STRUCTURE)</Typography>
                <IconButton onClick={onClose} sx={{ color: 'white' }}><CloseIcon /></IconButton>
            </DialogTitle>

            <DialogContent sx={{ p: 4, bgcolor: '#f8f9fc' }}>
                <Grid container spacing={3} wrap="nowrap" alignItems="flex-start">

                    {/* Left: Input Form & Upload */}
                    <Grid item sx={{ width: '300px', flexShrink: 0 }}>
                        <Paper sx={{ p: 2, borderRadius: '15px' }}>
                            <Typography variant="caption" sx={{ fontWeight: 'bold', mb: 1, display: 'block' }}>เพิ่มแผนกจัดกลุ่มใหม่</Typography>
                            <TextField
                                size="small" fullWidth label="Group 1" sx={{ mb: 1, '& .MuiInputBase-input': { fontSize: '0.75rem', py: 1 } }}
                                value={newRow.group1} onChange={(e) => setNewRow({ ...newRow, group1: e.target.value })}
                            />
                            <TextField
                                size="small" fullWidth label="Group 2" sx={{ mb: 1, '& .MuiInputBase-input': { fontSize: '0.75rem', py: 1 } }}
                                value={newRow.group2} onChange={(e) => setNewRow({ ...newRow, group2: e.target.value })}
                            />
                            <TextField
                                size="small" fullWidth label="Group 3" sx={{ mb: 1, '& .MuiInputBase-input': { fontSize: '0.75rem', py: 1 } }}
                                value={newRow.group3} onChange={(e) => setNewRow({ ...newRow, group3: e.target.value })}
                            />
                            <TextField
                                size="small" fullWidth label="Conso GL" sx={{ mb: 1, '& .MuiInputBase-input': { fontSize: '0.75rem', py: 1 } }}
                                value={newRow.conso_gl} onChange={(e) => setNewRow({ ...newRow, conso_gl: e.target.value })}
                            />
                            <TextField
                                size="small" fullWidth label="Account Name" sx={{ mb: 1, '& .MuiInputBase-input': { fontSize: '0.75rem', py: 1 } }}
                                value={newRow.account_name} onChange={(e) => setNewRow({ ...newRow, account_name: e.target.value })}
                            />
                            <Button
                                variant="contained" fullWidth startIcon={<AddIcon sx={{ fontSize: 18 }} />}
                                sx={{ borderRadius: '10px', bgcolor: '#043478', textTransform: 'none', py: 1.2, fontSize: '0.85rem', fontWeight: 'bold' }}
                                onClick={handleAdd}
                            >
                                เพิ่มโครงสร้าง
                            </Button>
                        </Paper>
                    </Grid>

                    {/* Right: Data Table */}
                    <Grid item sx={{ flexGrow: 1, minWidth: 0 }}>
                        <Paper sx={{ p: 3, borderRadius: '15px', height: '100%', display: 'flex', flexDirection: 'column' }}>
                            <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>
                                รายการทั้งหมด ({structures.length})
                            </Typography>

                            <StructureTable
                                structures={structures}
                                loading={loading}
                                onDelete={handleDelete}
                            />
                        </Paper>
                    </Grid>

                </Grid>
            </DialogContent>
        </Dialog>
    );
};

export default BudgetStructureModal;
