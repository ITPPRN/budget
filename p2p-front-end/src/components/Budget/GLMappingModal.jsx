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
const MappingTable = React.memo(({ mappings, loading, onDelete }) => (
    <TableContainer sx={{ flexGrow: 1, maxHeight: 500, overflow: 'auto' }}>
        <Table size="small" stickyHeader>
            <TableHead>
                <TableRow>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5' }}>Entity</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5' }}>Entity GL</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5' }}>Conso GL</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5' }}>Account Name</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', width: 60 }}>ลบ</TableCell>
                </TableRow>
            </TableHead>
            <TableBody>
                {loading ? (
                    <TableRow><TableCell colSpan={5} align="center"><CircularProgress size={30} sx={{ my: 4 }} /></TableCell></TableRow>
                ) : mappings.map((row) => (
                    <TableRow key={row.id} hover>
                        <TableCell>{row.entity}</TableCell>
                        <TableCell>{row.entity_gl}</TableCell>
                        <TableCell>{row.conso_gl}</TableCell>
                        <TableCell sx={{ maxWidth: 150 }} noWrap title={row.account_name}>{row.account_name}</TableCell>
                        <TableCell>
                            <IconButton size="small" color="error" onClick={() => onDelete(row.id)}>
                                <DeleteIcon fontSize="small" />
                            </IconButton>
                        </TableCell>
                    </TableRow>
                ))}
                {!loading && mappings.length === 0 && (
                    <TableRow><TableCell colSpan={5} align="center" sx={{ py: 4, color: '#999' }}>ไม่มีข้อมูล แนะนำให้อัปโหลดไฟล์ .xlsx</TableCell></TableRow>
                )}
            </TableBody>
        </Table>
    </TableContainer>
));

const GLMappingModal = ({ open, onClose }) => {
    const [mappings, setMappings] = useState([]);
    const [loading, setLoading] = useState(false);
    const [uploading, setUploading] = useState(false);
    const fileInputRef = useRef(null);

    // Form State
    const [newRow, setNewRow] = useState({ entity: '', entity_gl: '', conso_gl: '', account_name: '' });

    const fetchMappings = useCallback(async () => {
        setLoading(true);
        try {
            const res = await api.get('/budgets/gl-mappings');
            setMappings(res.data || []);
        } catch (error) {
            toast.error('Failed to load GL Mappings');
        } finally {
            setLoading(false);
        }
    }, []);

    // Reset state when modal closes
    useEffect(() => {
        if (!open) {
            setNewRow({ entity: '', entity_gl: '', conso_gl: '', account_name: '' });
        } else {
            fetchMappings();
        }
    }, [open, fetchMappings]);

    const handleAdd = async () => {
        if (!newRow.entity || !newRow.entity_gl || !newRow.conso_gl) {
            toast.warning('กรุณากรอกข้อมูล Entity, Entity GL, และ Conso GL ให้ครบ');
            return;
        }
        try {
            await api.post('/budgets/gl-mappings', newRow);
            toast.success('เพิ่มข้อมูลเรียบร้อย');
            setNewRow({ entity: '', entity_gl: '', conso_gl: '', account_name: '' });
            fetchMappings();
        } catch (error) {
            toast.error('ไม่สามารถเพิ่มข้อมูลได้');
        }
    };

    const handleDelete = useCallback(async (id) => {
        if (!window.confirm('ยืนยันการลบข้อมูลนี้?')) return;
        try {
            await api.delete(`/budgets/gl-mappings/${id}`);
            toast.success('ลบข้อมูลเรียบร้อย');
            fetchMappings();
        } catch (error) {
            toast.error('ไม่สามารถลบข้อมูลได้');
        }
    }, [fetchMappings]);

    const handleFileSelect = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        const formData = new FormData();
        formData.append('file', file);

        setUploading(true);
        try {
            await api.post('/budgets/gl-mappings/import', formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            toast.success('อัปโหลดไฟล์สำเร็จ');
            fetchMappings();
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
        <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
            <DialogTitle sx={{ bgcolor: '#043478', color: 'white', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography sx={{ fontWeight: 'bold', color: 'white' }}>MANAGE GL MAPPING</Typography>
                <IconButton onClick={onClose} sx={{ color: 'white' }}><CloseIcon /></IconButton>
            </DialogTitle>

            <DialogContent sx={{ p: 4, bgcolor: '#f8f9fc' }}>
                <Grid container spacing={3} wrap="nowrap" alignItems="flex-start">

                    {/* Left: Input Form & Upload */}
                    <Grid item sx={{ width: '300px', flexShrink: 0 }}>
                        <Paper sx={{ p: 2, borderRadius: '15px', mb: 2 }}>
                            <Typography variant="caption" sx={{ fontWeight: 'bold', mb: 1, display: 'block' }}>เพิ่มข้อมูลใหม่ทีละรายการ</Typography>
                            <TextField
                                size="small" fullWidth label="Entity" sx={{ mb: 1, '& .MuiInputBase-input': { fontSize: '0.75rem', py: 1 } }}
                                value={newRow.entity} onChange={(e) => setNewRow({ ...newRow, entity: e.target.value })}
                            />
                            <TextField
                                size="small" fullWidth label="Entity GL" sx={{ mb: 1, '& .MuiInputBase-input': { fontSize: '0.75rem', py: 1 } }}
                                value={newRow.entity_gl} onChange={(e) => setNewRow({ ...newRow, entity_gl: e.target.value })}
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
                                เพิ่มข้อมูล
                            </Button>
                        </Paper>

                        <Paper sx={{ p: 2, borderRadius: '15px', textAlign: 'center' }}>
                            <Typography variant="caption" sx={{ fontWeight: 'bold', mb: 0.5, display: 'block' }}>อัปโหลดหลายรายการ (.xlsx)</Typography>
                            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block', mb: 1.5, fontSize: '0.65rem', lineHeight: 1.2 }}>
                                ต้องมี 4 คอลัมน์: Entity, Entity GL, Conso GL, Account Name
                            </Typography>
                            <input type="file" ref={fileInputRef} hidden onChange={handleFileSelect} accept=".xlsx" />
                            <Button
                                variant="outlined" fullWidth startIcon={uploading ? <CircularProgress size={18} /> : <CloudUploadIcon sx={{ fontSize: 18 }} />}
                                sx={{ borderRadius: '10px', textTransform: 'none', py: 0.8, fontSize: '0.8rem' }}
                                onClick={() => fileInputRef.current.click()} disabled={uploading}
                            >
                                {uploading ? 'กำลังอัปโหลด...' : 'อัปโหลดไฟล์ (.xlsx)'}
                            </Button>
                        </Paper>
                    </Grid>

                    {/* Right: Data Table */}
                    <Grid item sx={{ flexGrow: 1, minWidth: 0 }}>
                        <Paper sx={{ p: 3, borderRadius: '15px', height: '100%', display: 'flex', flexDirection: 'column' }}>
                            <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>
                                รายการทั้งหมด ({mappings.length})
                            </Typography>

                            <MappingTable
                                mappings={mappings}
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

export default GLMappingModal;
