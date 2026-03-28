import React, { useState, useEffect, useRef, useCallback } from 'react';
import {
    Dialog, DialogTitle, DialogContent, Typography, IconButton,
    Box, Grid, TextField, Button, Table, TableBody, TableCell,
    TableContainer, TableHead, TableRow, Paper, CircularProgress, Divider
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import DeleteIcon from '@mui/icons-material/Delete';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import AddIcon from '@mui/icons-material/Add';
import api from '../../utils/api/axiosInstance';
import { toast } from 'react-toastify';

// Memoized Table Component
const GroupingTable = React.memo(({ groupings, loading, onDelete }) => (
    <TableContainer sx={{ flexGrow: 1, maxHeight: 500, overflow: 'auto' }}>
        <Table size="small" stickyHeader>
            <TableHead>
                <TableRow>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>Group 1</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>Group 2</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>Group 3</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>Entity</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>Entity GL</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>Conso GL</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>Account Name</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', width: 50 }}>ลบ</TableCell>
                </TableRow>
            </TableHead>
            <TableBody>
                {loading ? (
                    <TableRow><TableCell colSpan={8} align="center"><CircularProgress size={30} sx={{ my: 4 }} /></TableCell></TableRow>
                ) : groupings.map((row) => (
                    <TableRow key={row.id} hover>
                        <TableCell sx={{ fontSize: '0.7rem' }}>{row.group1}</TableCell>
                        <TableCell sx={{ fontSize: '0.7rem' }}>{row.group2}</TableCell>
                        <TableCell sx={{ fontSize: '0.7rem' }}>{row.group3}</TableCell>
                        <TableCell sx={{ fontSize: '0.7rem' }}>{row.entity}</TableCell>
                        <TableCell sx={{ fontSize: '0.7rem' }}>{row.entity_gl}</TableCell>
                        <TableCell sx={{ fontSize: '0.7rem' }}>{row.conso_gl}</TableCell>
                        <TableCell sx={{ fontSize: '0.7rem', maxWidth: 120 }} noWrap title={row.account_name}>{row.account_name}</TableCell>
                        <TableCell>
                            <IconButton size="small" color="error" onClick={() => onDelete(row.id)}>
                                <DeleteIcon fontSize="inherit" />
                            </IconButton>
                        </TableCell>
                    </TableRow>
                ))}
                {!loading && groupings.length === 0 && (
                    <TableRow><TableCell colSpan={8} align="center" sx={{ py: 4, color: '#999' }}>ไม่มีข้อมูล แนะนำให้อัปโหลดไฟล์ .xlsx (7 คอลัมน์)</TableCell></TableRow>
                )}
            </TableBody>
        </Table>
    </TableContainer>
));

const GLGroupingModal = ({ open, onClose }) => {
    const [groupings, setGroupings] = useState([]);
    const [loading, setLoading] = useState(false);
    const [uploading, setUploading] = useState(false);
    const fileInputRef = useRef(null);

    // Form State (7 Fields)
    const [newRow, setNewRow] = useState({ 
        entity: '', 
        entity_gl: '', 
        conso_gl: '', 
        account_name: '',
        group1: '',
        group2: '',
        group3: ''
    });

    const fetchGroupings = useCallback(async () => {
        setLoading(true);
        try {
            const res = await api.get('/budgets/gl-grouping-list');
            setGroupings(res.data || []);
        } catch (error) {
            toast.error('Failed to load GL Groupings');
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        if (open) {
            fetchGroupings();
        } else {
            setNewRow({ entity: '', entity_gl: '', conso_gl: '', account_name: '', group1: '', group2: '', group3: '' });
        }
    }, [open, fetchGroupings]);

    const handleAdd = async () => {
        if (!newRow.entity || !newRow.entity_gl || !newRow.conso_gl || !newRow.group1) {
            toast.warning('กรุณากรอกข้อมูล Entity, GL, Conso และ Group 1 ให้ครบ');
            return;
        }
        try {
            await api.post('/budgets/gl-groupings', newRow);
            toast.success('เพิ่มข้อมูลเรียบร้อย');
            setNewRow({ entity: '', entity_gl: '', conso_gl: '', account_name: '', group1: '', group2: '', group3: '' });
            fetchGroupings();
        } catch (error) {
            toast.error('ไม่สามารถเพิ่มข้อมูลได้');
        }
    };

    const handleDelete = useCallback(async (id) => {
        if (!window.confirm('ยืนยันการลบข้อมูลนี้?')) return;
        try {
            await api.delete(`/budgets/gl-groupings/${id}`);
            toast.success('ลบข้อมูลเรียบร้อย');
            fetchGroupings();
        } catch (error) {
            toast.error('ไม่สามารถลบข้อมูลได้');
        }
    }, [fetchGroupings]);

    const handleFileSelect = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        const formData = new FormData();
        formData.append('file', file);

        setUploading(true);
        try {
            await api.post('/budgets/gl-groupings/import', formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            toast.success('อัปโหลดไฟล์สำเร็จ');
            fetchGroupings();
        } catch (error) {
            const errorMsg = error.response?.data?.message || error.response?.data?.error || 'อัปโหลดไฟล์ล้มเหลว';
            toast.error(errorMsg);
        } finally {
            setUploading(false);
            e.target.value = null; // reset input
        }
    };

    if (!open) return null;

    return (
        <Dialog open={open} onClose={onClose} maxWidth="xl" fullWidth>
            <DialogTitle sx={{ bgcolor: '#043478', color: 'white', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography sx={{ fontWeight: 'bold', color: 'white' }}>MANAGE GL GROUPING (UNIFIED)</Typography>
                <IconButton onClick={onClose} sx={{ color: 'white' }}><CloseIcon /></IconButton>
            </DialogTitle>

            <DialogContent sx={{ p: 4, bgcolor: '#f8f9fc' }}>
                <Grid container spacing={3} wrap="nowrap" alignItems="flex-start">

                    {/* Left: Input Form & Upload */}
                    <Grid item sx={{ width: '300px', flexShrink: 0 }}>
                        <Paper sx={{ p: 2, borderRadius: '15px', mb: 2 }}>
                            <Typography variant="caption" sx={{ fontWeight: 'bold', mb: 1, display: 'block' }}>เพิ่มข้อมูลใหม่</Typography>
                            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                                <TextField
                                    size="small" fullWidth label="Group 1" variant="filled"
                                    value={newRow.group1} onChange={(e) => setNewRow({ ...newRow, group1: e.target.value })}
                                />
                                <TextField
                                    size="small" fullWidth label="Group 2" variant="filled"
                                    value={newRow.group2} onChange={(e) => setNewRow({ ...newRow, group2: e.target.value })}
                                />
                                <TextField
                                    size="small" fullWidth label="Group 3" variant="filled"
                                    value={newRow.group3} onChange={(e) => setNewRow({ ...newRow, group3: e.target.value })}
                                />
                                <Divider sx={{ my: 0.5 }} />
                                <TextField
                                    size="small" fullWidth label="Entity" variant="filled"
                                    value={newRow.entity} onChange={(e) => setNewRow({ ...newRow, entity: e.target.value })}
                                />
                                <TextField
                                    size="small" fullWidth label="Entity GL" variant="filled"
                                    value={newRow.entity_gl} onChange={(e) => setNewRow({ ...newRow, entity_gl: e.target.value })}
                                />
                                <TextField
                                    size="small" fullWidth label="Conso GL" variant="filled"
                                    value={newRow.conso_gl} onChange={(e) => setNewRow({ ...newRow, conso_gl: e.target.value })}
                                />
                                <TextField
                                    size="small" fullWidth label="Account Name" variant="filled"
                                    value={newRow.account_name} onChange={(e) => setNewRow({ ...newRow, account_name: e.target.value })}
                                />
                                <Button
                                    variant="contained" fullWidth startIcon={<AddIcon />}
                                    sx={{ borderRadius: '10px', bgcolor: '#043478', mt: 1, fontWeight: 'bold' }}
                                    onClick={handleAdd}
                                >
                                    เพิ่มรายการ
                                </Button>
                            </Box>
                        </Paper>

                        <Paper sx={{ p: 2, borderRadius: '15px', textAlign: 'center' }}>
                            <Typography variant="caption" sx={{ fontWeight: 'bold', mb: 0.5, display: 'block' }}>อัปโหลด .xlsx (7 คอลัมน์)</Typography>
                            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block', mb: 1.5, fontSize: '0.6rem', lineHeight: 1.2 }}>
                                Group1, Group2, Group3, Entity, Entity GL, Conso GL, Account Name
                            </Typography>
                            <input type="file" ref={fileInputRef} hidden onChange={handleFileSelect} accept=".xlsx" />
                            <Button
                                variant="outlined" fullWidth startIcon={uploading ? <CircularProgress size={18} /> : <CloudUploadIcon />}
                                sx={{ borderRadius: '10px', fontWeight: 'bold' }}
                                onClick={() => fileInputRef.current.click()} disabled={uploading}
                            >
                                {uploading ? 'กำลังอัปโหลด...' : 'อัปโหลดไฟล์'}
                            </Button>
                        </Paper>
                    </Grid>

                    {/* Right: Data Table */}
                    <Grid item sx={{ flexGrow: 1, minWidth: 0 }}>
                        <Paper sx={{ p: 3, borderRadius: '15px', height: '100%', display: 'flex', flexDirection: 'column' }}>
                            <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>
                                รายการทั้งหมด ({groupings.length})
                            </Typography>

                            <GroupingTable
                                groupings={groupings}
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

export default GLGroupingModal;
