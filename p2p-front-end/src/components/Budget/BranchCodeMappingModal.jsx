import React, { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import {
    Dialog, DialogTitle, DialogContent, Typography, IconButton,
    Box, Grid, TextField, Button, Table, TableBody, TableCell,
    TableContainer, TableHead, TableRow, Paper, CircularProgress,
    MenuItem, Autocomplete, Chip
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import DeleteIcon from '@mui/icons-material/Delete';
import AddIcon from '@mui/icons-material/Add';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import DownloadIcon from '@mui/icons-material/Download';
import api from '../../utils/api/axiosInstance';
import { toast } from 'react-toastify';

// Memoized table — each row = one (company, branch_code) pair.
// A company with multiple codes appears multiple times.
const MappingTable = React.memo(({ rows, companies, loading, onDelete, onAddCodesForCompany }) => (
    <TableContainer sx={{ flexGrow: 1, maxHeight: 500, overflow: 'auto' }}>
        <Table size="small" stickyHeader>
            <TableHead>
                <TableRow>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>บริษัท</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>สาขา</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>Branch No.</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', fontSize: '0.75rem' }}>Branch Code</TableCell>
                    <TableCell sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', width: 50 }}>ลบ</TableCell>
                </TableRow>
            </TableHead>
            <TableBody>
                {loading ? (
                    <TableRow><TableCell colSpan={5} align="center"><CircularProgress size={30} sx={{ my: 4 }} /></TableCell></TableRow>
                ) : rows.map((m) => {
                    const co = companies.find(c => c.id === m.company_id) || m.company;
                    return (
                        <TableRow
                            key={m.id} hover
                            sx={{ cursor: 'pointer' }}
                            onClick={() => onAddCodesForCompany(m.company_id)}
                            title="คลิกเพื่อเพิ่ม code อื่นให้บริษัทนี้"
                        >
                            <TableCell sx={{ fontSize: '0.7rem' }}>{co?.name || '—'}</TableCell>
                            <TableCell sx={{ fontSize: '0.7rem' }}>{co?.branch_name || co?.branch_name_en || '—'}</TableCell>
                            <TableCell sx={{ fontSize: '0.7rem' }}>{co?.branch_no || '—'}</TableCell>
                            <TableCell sx={{ fontSize: '0.7rem' }}>
                                <code style={{ background: '#eef2ff', padding: '2px 6px', borderRadius: 4 }}>{m.branch_code}</code>
                            </TableCell>
                            <TableCell onClick={(e) => e.stopPropagation()}>
                                <IconButton size="small" color="error" onClick={() => onDelete(m)}>
                                    <DeleteIcon fontSize="inherit" />
                                </IconButton>
                            </TableCell>
                        </TableRow>
                    );
                })}
                {!loading && rows.length === 0 && (
                    <TableRow><TableCell colSpan={5} align="center" sx={{ py: 4, color: '#999' }}>ยังไม่มี mapping — เพิ่มจากฟอร์มด้านซ้าย</TableCell></TableRow>
                )}
            </TableBody>
        </Table>
    </TableContainer>
));

const BranchCodeMappingModal = ({ open, onClose }) => {
    const [mappings, setMappings] = useState([]);
    const [companies, setCompanies] = useState([]);
    const [availableCodes, setAvailableCodes] = useState([]);
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [uploading, setUploading] = useState(false);
    const [downloading, setDownloading] = useState(false);
    // Form: 1 company → many codes (multi-select Autocomplete)
    const [form, setForm] = useState({ company_id: '', branch_codes: [] });
    const fileInputRef = useRef(null);

    const loadAll = useCallback(async () => {
        setLoading(true);
        try {
            const [mapRes, coRes, codeRes] = await Promise.all([
                api.get('/auth/manage/branch-code-mappings'),
                api.get('/auth/manage/companies'),
                api.get('/auth/manage/branch-codes'),
            ]);
            setMappings(mapRes.data?.data || mapRes.data || []);
            setCompanies(coRes.data?.data || coRes.data || []);
            setAvailableCodes(codeRes.data?.data || codeRes.data || []);
        } catch (e) {
            toast.error('โหลดข้อมูลล้มเหลว: ' + (e.response?.data?.error || e.message));
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        if (open) {
            loadAll();
        } else {
            setForm({ company_id: '', branch_codes: [] });
        }
    }, [open, loadAll]);

    // Codes already mapped to the currently-selected company — shown for context
    const existingCodesForSelected = useMemo(() => {
        if (!form.company_id) return [];
        return mappings.filter(m => m.company_id === form.company_id).map(m => m.branch_code);
    }, [form.company_id, mappings]);

    const handleSave = async () => {
        if (!form.company_id) {
            toast.warning('กรุณาเลือกบริษัท/สาขา');
            return;
        }
        const codes = form.branch_codes.map(c => (c || '').trim()).filter(Boolean);
        if (codes.length === 0) {
            toast.warning('กรุณาเลือกหรือกรอก Branch Code อย่างน้อย 1 ค่า');
            return;
        }

        setSaving(true);
        try {
            // Upsert is idempotent on (company_id, branch_code) — already-mapped codes are no-op
            await Promise.all(codes.map(code =>
                api.put('/auth/manage/branch-code-mappings', {
                    company_id: form.company_id,
                    branch_code: code,
                })
            ));
            toast.success(`บันทึก ${codes.length} mapping สำเร็จ`);
            setForm({ company_id: '', branch_codes: [] });
            loadAll();
        } catch (e) {
            toast.error('บันทึกล้มเหลว: ' + (e.response?.data?.error || e.message));
        } finally {
            setSaving(false);
        }
    };

    const handleDelete = useCallback(async (m) => {
        if (!window.confirm(`ลบ code "${m.branch_code}" ออกจากบริษัทนี้?`)) return;
        try {
            await api.delete(`/auth/manage/branch-code-mappings/${m.id}`);
            toast.success('ลบ mapping สำเร็จ');
            loadAll();
        } catch (e) {
            toast.error('ลบล้มเหลว: ' + (e.response?.data?.error || e.message));
        }
    }, [loadAll]);

    // Click row → pre-select that company so admin can quickly add more codes
    const handleAddCodesForCompany = useCallback((companyID) => {
        setForm({ company_id: companyID, branch_codes: [] });
    }, []);

    const handleDownloadTemplate = async () => {
        setDownloading(true);
        try {
            const res = await api.get('/auth/manage/branch-code-mappings/template', {
                responseType: 'blob',
            });
            const blob = new Blob([res.data], {
                type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
            });
            const url = window.URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.href = url;
            link.download = 'branch_code_mapping_template.xlsx';
            document.body.appendChild(link);
            link.click();
            link.remove();
            window.URL.revokeObjectURL(url);
        } catch (e) {
            toast.error('โหลด template ล้มเหลว: ' + (e.response?.data?.error || e.message));
        } finally {
            setDownloading(false);
        }
    };

    const handleFileSelect = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        const formData = new FormData();
        formData.append('file', file);

        setUploading(true);
        try {
            const res = await api.post('/auth/manage/branch-code-mappings/import', formData, {
                headers: { 'Content-Type': 'multipart/form-data' },
            });
            const r = res.data?.data || res.data || {};
            const summary = `เพิ่ม ${r.imported || 0} • อัปเดต ${r.updated || 0} • ข้าม ${r.skipped || 0}`;
            if (r.errors && r.errors.length > 0) {
                toast.warning(`อัปโหลดเสร็จ — ${summary}\n${r.errors.slice(0, 3).join('\n')}${r.errors.length > 3 ? `\n(+${r.errors.length - 3} อื่น)` : ''}`);
            } else {
                toast.success(`อัปโหลดเสร็จ — ${summary}`);
            }
            loadAll();
        } catch (err) {
            const msg = err.response?.data?.message || err.response?.data?.error || 'อัปโหลดไฟล์ล้มเหลว';
            toast.error(msg);
        } finally {
            setUploading(false);
            e.target.value = null;
        }
    };

    if (!open) return null;

    return (
        <Dialog open={open} onClose={onClose} maxWidth="xl" fullWidth>
            <DialogTitle sx={{ bgcolor: '#043478', color: 'white', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography sx={{ fontWeight: 'bold', color: 'white' }}>MANAGE BRANCH CODE MAPPING</Typography>
                <IconButton onClick={onClose} sx={{ color: 'white' }}><CloseIcon /></IconButton>
            </DialogTitle>

            <DialogContent sx={{ p: 4, bgcolor: '#f8f9fc' }}>
                <Grid container spacing={3} wrap="nowrap" alignItems="flex-start">

                    {/* Left: Form + Upload */}
                    <Grid item sx={{ width: '340px', flexShrink: 0 }}>
                        <Paper sx={{ p: 2, borderRadius: '15px', mb: 2 }}>
                            <Typography variant="caption" sx={{ fontWeight: 'bold', mb: 1, display: 'block' }}>
                                เพิ่ม Branch Code ให้บริษัท
                            </Typography>
                            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.2 }}>
                                <TextField
                                    select size="small" fullWidth label="บริษัท / สาขา" variant="filled"
                                    value={form.company_id}
                                    onChange={(e) => setForm({ company_id: e.target.value, branch_codes: [] })}
                                >
                                    <MenuItem value="" disabled>-- เลือก --</MenuItem>
                                    {companies.map(c => (
                                        <MenuItem key={c.id} value={c.id}>
                                            {formatCompanyShort(c)}
                                        </MenuItem>
                                    ))}
                                </TextField>

                                {existingCodesForSelected.length > 0 && (
                                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, mb: 0.5 }}>
                                        <Typography variant="caption" sx={{ width: '100%', color: 'text.secondary', fontSize: '0.65rem' }}>
                                            มีอยู่แล้ว:
                                        </Typography>
                                        {existingCodesForSelected.map(c => (
                                            <Chip key={c} label={c} size="small" sx={{ bgcolor: '#e0e7ff', fontSize: '0.7rem', height: 20 }} />
                                        ))}
                                    </Box>
                                )}

                                <Autocomplete
                                    multiple
                                    freeSolo
                                    size="small"
                                    options={availableCodes}
                                    value={form.branch_codes}
                                    onChange={(_, v) => setForm({ ...form, branch_codes: v })}
                                    renderTags={(value, getTagProps) =>
                                        value.map((option, index) => (
                                            <Chip variant="outlined" label={option} size="small" {...getTagProps({ index })} />
                                        ))
                                    }
                                    renderInput={(params) => (
                                        <TextField
                                            {...params}
                                            label="Branch Codes (เลือกได้หลายค่า)"
                                            variant="filled"
                                            placeholder="เช่น HOF, Branch00"
                                            helperText={
                                                availableCodes.length > 0
                                                    ? `เลือกจาก ${availableCodes.length} ค่าที่พบใน actual หรือพิมพ์เพิ่มได้`
                                                    : 'พิมพ์ค่าใหม่ได้เลย (ยังไม่มีข้อมูล actual)'
                                            }
                                        />
                                    )}
                                />

                                <Button
                                    variant="contained" fullWidth startIcon={<AddIcon />}
                                    sx={{ borderRadius: '10px', bgcolor: '#043478', mt: 1, fontWeight: 'bold' }}
                                    onClick={handleSave} disabled={saving}
                                >
                                    {saving ? 'กำลังบันทึก...' : 'บันทึก'}
                                </Button>

                                <Button
                                    size="small" fullWidth onClick={() => setForm({ company_id: '', branch_codes: [] })}
                                    disabled={saving}
                                >
                                    ล้างฟอร์ม
                                </Button>
                            </Box>
                        </Paper>

                        <Paper sx={{ p: 2, borderRadius: '15px', textAlign: 'center', mb: 2 }}>
                            <Typography variant="caption" sx={{ fontWeight: 'bold', mb: 0.5, display: 'block' }}>อัปโหลด .xlsx (4 คอลัมน์)</Typography>
                            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block', mb: 1.5, fontSize: '0.6rem', lineHeight: 1.3 }}>
                                Company Name, Branch Name, Branch No, Branch Code<br />
                                (1 row = 1 code, company ซ้ำได้ถ้ามีหลาย code)
                            </Typography>
                            <input type="file" ref={fileInputRef} hidden onChange={handleFileSelect} accept=".xlsx" />
                            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                                <Button
                                    variant="text" fullWidth size="small"
                                    startIcon={downloading ? <CircularProgress size={14} /> : <DownloadIcon />}
                                    sx={{ fontWeight: 'bold', fontSize: '0.7rem' }}
                                    onClick={handleDownloadTemplate} disabled={downloading}
                                >
                                    {downloading ? 'กำลังโหลด...' : 'ดาวน์โหลด Template'}
                                </Button>
                                <Button
                                    variant="outlined" fullWidth startIcon={uploading ? <CircularProgress size={18} /> : <CloudUploadIcon />}
                                    sx={{ borderRadius: '10px', fontWeight: 'bold' }}
                                    onClick={() => fileInputRef.current?.click()} disabled={uploading}
                                >
                                    {uploading ? 'กำลังอัปโหลด...' : 'อัปโหลดไฟล์'}
                                </Button>
                            </Box>
                        </Paper>

                        <Paper sx={{ p: 2, borderRadius: '15px' }}>
                            <Typography variant="caption" sx={{ fontWeight: 'bold', display: 'block', mb: 0.5 }}>
                                คำแนะนำ
                            </Typography>
                            <Typography variant="caption" sx={{ color: 'text.secondary', display: 'block', fontSize: '0.7rem', lineHeight: 1.6 }}>
                                • <b>1 บริษัท map ได้หลาย code</b> (เช่น CLIK HQ → HOF + Branch00)<br />
                                • คลิกแถวในตาราง → เลือก company นั้นใน form เพื่อเพิ่ม code อื่น<br />
                                • ปุ่ม "ลบ" ลบทีละ code (code อื่นของบริษัทเดียวกันยังอยู่)<br />
                                • Mapping นี้ใช้สำหรับ role <b>BRANCH_DELEGATE</b>
                            </Typography>
                        </Paper>
                    </Grid>

                    {/* Right: Table */}
                    <Grid item sx={{ flexGrow: 1, minWidth: 0 }}>
                        <Paper sx={{ p: 3, borderRadius: '15px', height: '100%', display: 'flex', flexDirection: 'column' }}>
                            <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>
                                รายการทั้งหมด ({mappings.length})
                            </Typography>

                            <MappingTable
                                rows={mappings}
                                companies={companies}
                                loading={loading}
                                onDelete={handleDelete}
                                onAddCodesForCompany={handleAddCodesForCompany}
                            />
                        </Paper>
                    </Grid>

                </Grid>
            </DialogContent>
        </Dialog>
    );
};

function formatCompanyShort(c) {
    if (!c) return '—';
    const branch = c.branch_name || c.branch_name_en;
    return branch ? `${c.name} • ${branch}` : c.name;
}

export default BranchCodeMappingModal;
