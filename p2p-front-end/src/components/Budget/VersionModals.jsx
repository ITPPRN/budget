import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import {
    Dialog, DialogTitle, DialogContent, Typography, IconButton,
    Box, Grid, TextField, Button, List, ListItem, ListItemAvatar,
    Avatar, Divider, CircularProgress, Paper
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import SaveIcon from '@mui/icons-material/Save';
import CancelIcon from '@mui/icons-material/Cancel';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import api from '../../utils/api/axiosInstance';
import { toast } from 'react-toastify';

// Shared UI List Component (Used by all modals to maintain look & feel)
const VersionList = React.memo(({ items, editingId, newName, onNewNameChange, onSave, onCancel, onStartRename, onDelete, activeId }) => (
    <Paper sx={{ maxHeight: 350, overflowY: 'auto', bgcolor: 'white', borderRadius: '15px', border: '1px solid #e0e0e0' }}>
        <List dense>
            {items.map((item) => (
                <React.Fragment key={item.id}>
                    <ListItem sx={{ py: 1.5, px: 2 }}>
                        <ListItemAvatar>
                            <Avatar sx={{ bgcolor: '#e8eaf6', color: '#1a237e' }}>
                                <InsertDriveFileIcon style={{ fontSize: 18 }} />
                            </Avatar>
                        </ListItemAvatar>

                        <Box sx={{ flexGrow: 1, minWidth: 0, mr: 2 }}>
                            {editingId === item.id ? (
                                <TextField
                                    size="small"
                                    value={newName}
                                    onChange={(e) => onNewNameChange(e.target.value)}
                                    autoFocus
                                    fullWidth
                                    variant="standard"
                                    sx={{ '& .MuiInput-underline:before': { borderBottomColor: '#1a237e' } }}
                                />
                            ) : (
                                <>
                                    <Typography variant="body2" sx={{ fontWeight: 600, color: '#333' }} noWrap title={item.file_name}>
                                        {item.file_name}
                                    </Typography>
                                    <Typography variant="caption" display="block" color="text.secondary" noWrap>
                                        อัปโหลดเมื่อ: {new Date(item.upload_at).toLocaleString()}
                                    </Typography>
                                </>
                            )}
                        </Box>

                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                            {editingId === item.id ? (
                                <>
                                    <IconButton size="small" onClick={() => onSave(item.id)} sx={{ color: 'success.main' }}>
                                        <SaveIcon fontSize="small" />
                                    </IconButton>
                                    <IconButton size="small" onClick={onCancel} color="default">
                                        <CancelIcon fontSize="small" />
                                    </IconButton>
                                </>
                            ) : (
                                <>
                                    <IconButton size="small" onClick={() => onStartRename(item.id, item.file_name)} sx={{ color: '#1a237e' }}>
                                        <EditIcon fontSize="small" />
                                    </IconButton>
                                    <IconButton
                                        size="small"
                                        onClick={() => onDelete(item.id)}
                                        sx={{ color: item.id === activeId ? '#ccc' : 'error.main' }}
                                        disabled={item.id === activeId}
                                        title={item.id === activeId ? "ไม่สามารถลบไฟล์ที่ใช้งานอยู่ได้" : "ลบไฟล์"}
                                    >
                                        <DeleteIcon fontSize="small" />
                                    </IconButton>
                                </>
                            )}
                        </Box>
                    </ListItem>
                    <Divider component="li" />
                </React.Fragment>
            ))}
            {items.length === 0 && (
                <Box sx={{ p: 4, textAlign: 'center', color: '#999' }}>ไม่พบข้อมูล Version</Box>
            )}
        </List>
    </Paper>
));

// ==========================================
// 1. BUDGET VERSION MODAL
// ==========================================
export const BudgetVersionModal = ({ open, onClose, items, onRefresh, activeId }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [fileToUpload, setFileToUpload] = useState(null);
    const [versionName, setVersionName] = useState('');
    const [isUploading, setIsUploading] = useState(false);
    const [editingId, setEditingId] = useState(null);
    const [newName, setNewName] = useState('');
    const [isDragging, setIsDragging] = useState(false);
    const fileInputRef = useRef(null);

    useEffect(() => { if (!open) { setSearchTerm(''); setFileToUpload(null); setVersionName(''); setEditingId(null); setIsDragging(false); } }, [open]);

    const activeList = useMemo(() => {
        const filtered = (items || []).filter(item => item.file_name?.toLowerCase().includes(searchTerm.toLowerCase()));
        return [...filtered].sort((a, b) => new Date(a.upload_at) - new Date(b.upload_at));
    }, [items, searchTerm]);

    const handleFileChange = (file) => {
        if (file) {
            setFileToUpload(file);
            setVersionName(file.name);
        }
    };

    const handleDragOver = (e) => {
        e.preventDefault();
        setIsDragging(true);
    };

    const handleDragLeave = () => {
        setIsDragging(false);
    };

    const handleDrop = (e) => {
        e.preventDefault();
        setIsDragging(false);
        const file = e.dataTransfer.files[0];
        handleFileChange(file);
    };

    const handleSaveUpload = async () => {
        if (!fileToUpload) return;
        setIsUploading(true);
        const formData = new FormData();
        formData.append('file', fileToUpload);
        formData.append('version_name', versionName);
        try {
            await api.post('/budgets/import-budget', formData, { headers: { 'Content-Type': 'multipart/form-data' } });
            toast.success("อัปโหลดไฟล์ Budget สำเร็จ!");
            setFileToUpload(null);
            onRefresh();
        } catch (error) {
            const errorMsg = error.response?.data?.error || "อัปโหลดล้มเหลว คอลัมน์หรือไฟล์ไม่ถูกต้อง";
            toast.error(errorMsg);
        } finally { setIsUploading(false); }
    };

    const handleDelete = async (id) => {
        if (!window.confirm("ยืนยันการลบไฟล์งบประมาณ?")) return;
        try {
            await api.delete(`/budgets/files-budget/${id}`);
            toast.success("ลบไฟล์สำเร็จ");
            onRefresh();
        } catch (error) { toast.error("ลบไฟล์ไม่สำเร็จ"); }
    };

    const handleSaveRename = async (id) => {
        try {
            await api.patch(`/budgets/files-budget/${id}`, { new_name: newName });
            toast.success("เปลี่ยนชื่อสำเร็จ");
            setEditingId(null);
            onRefresh();
        } catch (error) { toast.error("เปลี่ยนชื่อไม่สำเร็จ"); }
    };

    return (
        <Dialog
            open={open}
            onClose={onClose}
            maxWidth="md"
            fullWidth
        >
            <DialogTitle sx={{ bgcolor: '#1a237e', color: 'white', display: 'flex', justifyContent: 'space-between', alignItems: 'center', py: 1 }}>
                <Typography sx={{ fontWeight: 'bold', color: 'white' }}>MANAGE BUDGET</Typography>
                <IconButton onClick={onClose} sx={{ color: 'white' }}><CloseIcon /></IconButton>
            </DialogTitle>
            <DialogContent sx={{ p: 3, bgcolor: '#f8f9fc' }}>
                <Grid container spacing={3}>
                    <Grid item sx={{ width: '250px', flexShrink: 0 }}>
                        <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>UPLOAD NEW BUDGET</Typography>
                        {!fileToUpload ? (
                            <Box
                                sx={{
                                    border: '2px dashed #ccc',
                                    borderRadius: '15px',
                                    p: 4,
                                    textAlign: 'center',
                                    bgcolor: isDragging ? '#f0f4ff' : 'white',
                                    borderColor: isDragging ? '#1a237e' : '#ccc',
                                    mb: 3,
                                    cursor: 'pointer',
                                    transition: 'all 0.2s',
                                    '&:hover': { borderColor: '#1a237e', bgcolor: '#f0f4ff' }
                                }}
                                onClick={() => fileInputRef.current.click()}
                                onDragOver={handleDragOver}
                                onDragLeave={handleDragLeave}
                                onDrop={handleDrop}
                            >
                                <input type="file" hidden ref={fileInputRef} onChange={(e) => handleFileChange(e.target.files[0])} accept=".xlsx" />
                                <CloudUploadIcon sx={{ fontSize: 40, color: isDragging ? '#1a237e' : '#ccc', mb: 1 }} />
                                <Typography variant="body2" color="text.secondary">ลากไฟล์มาวางที่นี่ หรือคลิกเพื่อค้นหา</Typography>
                                <Button variant="contained" sx={{ mt: 2, borderRadius: '20px', bgcolor: '#1a237e' }}>เลือกไฟล์ .xlsx</Button>
                            </Box>
                        ) : (
                            <Box sx={{ p: 2, bgcolor: 'white', borderRadius: '15px', border: '1px solid #e3e6f0' }}>
                                <Typography variant="caption" color="text.secondary">SELECTED:</Typography>
                                <Typography variant="body2" sx={{ fontWeight: 'bold', mb: 2 }}>{fileToUpload.name}</Typography>
                                <TextField fullWidth size="small" label="ชื่อ Version" value={versionName} onChange={(e) => setVersionName(e.target.value)} sx={{ mb: 2 }} />
                                <Box sx={{ display: 'flex', gap: 1 }}>
                                    <Button variant="contained" fullWidth onClick={handleSaveUpload} disabled={isUploading} sx={{ bgcolor: '#1a237e' }}>{isUploading ? <CircularProgress size={20} /> : 'บันทึก'}</Button>
                                    <Button variant="outlined" onClick={() => setFileToUpload(null)}>ยกเลิก</Button>
                                </Box>
                            </Box>
                        )}
                    </Grid>
                    <Grid item sx={{ flexGrow: 1, minWidth: 0 }}>
                        <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>AVAILABLE BUDGET VERSIONS</Typography>
                        <TextField fullWidth size="small" placeholder="ค้นหาไฟล์งบประมาณ..." value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)} sx={{ mb: 2, bgcolor: 'white' }} />
                        <VersionList items={activeList} editingId={editingId} newName={newName} onNewNameChange={setNewName} onSave={handleSaveRename} onCancel={() => setEditingId(null)} onStartRename={(id, name) => { setEditingId(id); setNewName(name); }} onDelete={handleDelete} activeId={activeId} />
                    </Grid>
                </Grid>
            </DialogContent>
        </Dialog>
    );
};

// ==========================================
// 2. CAPEX PLAN MODAL
// ==========================================
export const CapexPlanModal = ({ open, onClose, items, onRefresh, activeId }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [fileToUpload, setFileToUpload] = useState(null);
    const [versionName, setVersionName] = useState('');
    const [isUploading, setIsUploading] = useState(false);
    const [editingId, setEditingId] = useState(null);
    const [newName, setNewName] = useState('');
    const [isDragging, setIsDragging] = useState(false);
    const fileInputRef = useRef(null);

    useEffect(() => { if (!open) { setSearchTerm(''); setFileToUpload(null); setVersionName(''); setEditingId(null); setIsDragging(false); } }, [open]);

    const activeList = useMemo(() => {
        const filtered = (items || []).filter(item => item.file_name?.toLowerCase().includes(searchTerm.toLowerCase()));
        return [...filtered].sort((a, b) => new Date(a.upload_at) - new Date(b.upload_at));
    }, [items, searchTerm]);

    const handleFileChange = (file) => {
        if (file) {
            setFileToUpload(file);
            setVersionName(file.name);
        }
    };

    const handleDragOver = (e) => {
        e.preventDefault();
        setIsDragging(true);
    };

    const handleDragLeave = () => {
        setIsDragging(false);
    };

    const handleDrop = (e) => {
        e.preventDefault();
        setIsDragging(false);
        const file = e.dataTransfer.files[0];
        handleFileChange(file);
    };

    const handleSaveUpload = async () => {
        if (!fileToUpload) return;
        setIsUploading(true);
        const formData = new FormData();
        formData.append('file', fileToUpload);
        formData.append('version_name', versionName);
        try {
            await api.post('/budgets/import-capex-budget', formData, { headers: { 'Content-Type': 'multipart/form-data' } });
            toast.success("อัปโหลด CAPEX Plan สำเร็จ!");
            setFileToUpload(null);
            onRefresh();
        } catch (error) {
            const errorMsg = error.response?.data?.error || "อัปโหลดล้มเหลว คอลัมน์หรือไฟล์ไม่ถูกต้อง";
            toast.error(errorMsg);
        } finally { setIsUploading(false); }
    };

    const handleDelete = async (id) => {
        if (!window.confirm("ยืนยันการลบไฟล์ CAPEX Plan?")) return;
        try {
            await api.delete(`/budgets/files-capex-budget/${id}`);
            toast.success("ลบไฟล์สำเร็จ");
            onRefresh();
        } catch (error) { toast.error("ลบไฟล์ไม่สำเร็จ"); }
    };

    const handleSaveRename = async (id) => {
        try {
            await api.patch(`/budgets/files-capex-budget/${id}`, { new_name: newName });
            toast.success("เปลี่ยนชื่อสำเร็จ");
            setEditingId(null);
            onRefresh();
        } catch (error) { toast.error("เปลี่ยนชื่อไม่สำเร็จ"); }
    };

    return (
        <Dialog
            open={open}
            onClose={onClose}
            maxWidth="md"
            fullWidth
        >
            <DialogTitle sx={{ bgcolor: '#1a237e', color: 'white', display: 'flex', justifyContent: 'space-between', alignItems: 'center', py: 1 }}>
                <Typography sx={{ fontWeight: 'bold', color: 'white' }}>MANAGE CAPEX PLAN</Typography>
                <IconButton onClick={onClose} sx={{ color: 'white' }}><CloseIcon /></IconButton>
            </DialogTitle>
            <DialogContent sx={{ p: 3, bgcolor: '#f8f9fc' }}>
                <Grid container spacing={3}>
                    <Grid item sx={{ width: '250px', flexShrink: 0 }}>
                        <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>UPLOAD CAPEX PLAN</Typography>
                        {!fileToUpload ? (
                            <Box
                                sx={{
                                    border: '2px dashed #ccc',
                                    borderRadius: '15px',
                                    p: 4,
                                    textAlign: 'center',
                                    bgcolor: isDragging ? '#f0f4ff' : 'white',
                                    borderColor: isDragging ? '#1a237e' : '#ccc',
                                    mb: 3,
                                    cursor: 'pointer',
                                    transition: 'all 0.2s',
                                    '&:hover': { borderColor: '#1a237e', bgcolor: '#f0f4ff' }
                                }}
                                onClick={() => fileInputRef.current.click()}
                                onDragOver={handleDragOver}
                                onDragLeave={handleDragLeave}
                                onDrop={handleDrop}
                            >
                                <input type="file" hidden ref={fileInputRef} onChange={(e) => handleFileChange(e.target.files[0])} accept=".xlsx" />
                                <CloudUploadIcon sx={{ fontSize: 40, color: isDragging ? '#1a237e' : '#ccc', mb: 1 }} />
                                <Typography variant="body2" color="text.secondary">ลากไฟล์มาวางที่นี่ หรือคลิกเพื่อค้นหา</Typography>
                                <Button variant="contained" sx={{ mt: 2, borderRadius: '20px', bgcolor: '#1a237e' }}>เลือกไฟล์ .xlsx</Button>
                            </Box>
                        ) : (
                            <Box sx={{ p: 2, bgcolor: 'white', borderRadius: '15px', border: '1px solid #e3e6f0' }}>
                                <Typography variant="caption" color="text.secondary">SELECTED:</Typography>
                                <Typography variant="body2" sx={{ fontWeight: 'bold', mb: 2 }}>{fileToUpload.name}</Typography>
                                <TextField fullWidth size="small" label="ชื่อ Version" value={versionName} onChange={(e) => setVersionName(e.target.value)} sx={{ mb: 2 }} />
                                <Box sx={{ display: 'flex', gap: 1 }}>
                                    <Button variant="contained" fullWidth onClick={handleSaveUpload} disabled={isUploading} sx={{ bgcolor: '#1a237e' }}> บันทึก </Button>
                                    <Button variant="outlined" onClick={() => setFileToUpload(null)}>ยกเลิก</Button>
                                </Box>
                            </Box>
                        )}
                    </Grid>
                    <Grid item sx={{ flexGrow: 1, minWidth: 0 }}>
                        <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>AVAILABLE CAPEX PLANS</Typography>
                        <TextField fullWidth size="small" placeholder="ค้นหาแผน CAPEX..." value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)} sx={{ mb: 2, bgcolor: 'white' }} />
                        <VersionList items={activeList} editingId={editingId} newName={newName} onNewNameChange={setNewName} onSave={handleSaveRename} onCancel={() => setEditingId(null)} onStartRename={(id, name) => { setEditingId(id); setNewName(name); }} onDelete={handleDelete} activeId={activeId} />
                    </Grid>
                </Grid>
            </DialogContent>
        </Dialog>
    );
};

// ==========================================
// 3. CAPEX ACTUAL MODAL
// ==========================================
export const CapexActualModal = ({ open, onClose, items, onRefresh, activeId }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [fileToUpload, setFileToUpload] = useState(null);
    const [versionName, setVersionName] = useState('');
    const [isUploading, setIsUploading] = useState(false);
    const [editingId, setEditingId] = useState(null);
    const [newName, setNewName] = useState('');
    const [isDragging, setIsDragging] = useState(false);
    const fileInputRef = useRef(null);

    useEffect(() => { if (!open) { setSearchTerm(''); setFileToUpload(null); setVersionName(''); setEditingId(null); setIsDragging(false); } }, [open]);

    const activeList = useMemo(() => {
        const filtered = (items || []).filter(item => item.file_name?.toLowerCase().includes(searchTerm.toLowerCase()));
        return [...filtered].sort((a, b) => new Date(a.upload_at) - new Date(b.upload_at));
    }, [items, searchTerm]);

    const handleFileChange = (file) => {
        if (file) {
            setFileToUpload(file);
            setVersionName(file.name);
        }
    };

    const handleDragOver = (e) => {
        e.preventDefault();
        setIsDragging(true);
    };

    const handleDragLeave = () => {
        setIsDragging(false);
    };

    const handleDrop = (e) => {
        e.preventDefault();
        setIsDragging(false);
        const file = e.dataTransfer.files[0];
        handleFileChange(file);
    };

    const handleSaveUpload = async () => {
        if (!fileToUpload) return;
        setIsUploading(true);
        const formData = new FormData();
        formData.append('file', fileToUpload);
        formData.append('version_name', versionName);
        try {
            await api.post('/budgets/import-capex-actual', formData, { headers: { 'Content-Type': 'multipart/form-data' } });
            toast.success("อัปโหลด CAPEX Actual สำเร็จ!");
            setFileToUpload(null);
            onRefresh();
        } catch (error) {
            const errorMsg = error.response?.data?.error || "อัปโหลดล้มเหลว คอลัมน์หรือไฟล์ไม่ถูกต้อง";
            toast.error(errorMsg);
        } finally { setIsUploading(false); }
    };

    const handleDelete = async (id) => {
        if (!window.confirm("ยืนยันการลบไฟล์ CAPEX Actual?")) return;
        try {
            await api.delete(`/budgets/files-capex-actual/${id}`);
            toast.success("ลบไฟล์สำเร็จ");
            onRefresh();
        } catch (error) { toast.error("ลบไฟล์ไม่สำเร็จ"); }
    };

    const handleSaveRename = async (id) => {
        try {
            await api.patch(`/budgets/files-capex-actual/${id}`, { new_name: newName });
            toast.success("เปลี่ยนชื่อสำเร็จ");
            setEditingId(null);
            onRefresh();
        } catch (error) { toast.error("เปลี่ยนชื่อไม่สำเร็จ"); }
    };

    return (
        <Dialog
            open={open}
            onClose={onClose}
            maxWidth="md"
            fullWidth
        >
            <DialogTitle sx={{ bgcolor: '#1a237e', color: 'white', display: 'flex', justifyContent: 'space-between', alignItems: 'center', py: 1 }}>
                <Typography sx={{ fontWeight: 'bold', color: 'white' }}>MANAGE CAPEX ACTUAL</Typography>
                <IconButton onClick={onClose} sx={{ color: 'white' }}><CloseIcon /></IconButton>
            </DialogTitle>
            <DialogContent sx={{ p: 3, bgcolor: '#f8f9fc' }}>
                <Grid container spacing={3}>
                    <Grid item sx={{ width: '250px', flexShrink: 0 }}>
                        <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>UPLOAD CAPEX ACTUAL</Typography>
                        {!fileToUpload ? (
                            <Box
                                sx={{
                                    border: '2px dashed #ccc',
                                    borderRadius: '15px',
                                    p: 4,
                                    textAlign: 'center',
                                    bgcolor: isDragging ? '#f0f4ff' : 'white',
                                    borderColor: isDragging ? '#1a237e' : '#ccc',
                                    mb: 3,
                                    cursor: 'pointer',
                                    transition: 'all 0.2s',
                                    '&:hover': { borderColor: '#1a237e', bgcolor: '#f0f4ff' }
                                }}
                                onClick={() => fileInputRef.current.click()}
                                onDragOver={handleDragOver}
                                onDragLeave={handleDragLeave}
                                onDrop={handleDrop}
                            >
                                <input type="file" hidden ref={fileInputRef} onChange={(e) => handleFileChange(e.target.files[0])} accept=".xlsx" />
                                <CloudUploadIcon sx={{ fontSize: 40, color: isDragging ? '#1a237e' : '#ccc', mb: 1 }} />
                                <Typography variant="body2" color="text.secondary">ลากไฟล์มาวางที่นี่ หรือคลิกเพื่อค้นหา</Typography>
                                <Button variant="contained" sx={{ mt: 2, borderRadius: '20px', bgcolor: '#1a237e' }}>เลือกไฟล์ .xlsx</Button>
                            </Box>
                        ) : (
                            <Box sx={{ p: 2, bgcolor: 'white', borderRadius: '15px', border: '1px solid #e3e6f0' }}>
                                <Typography variant="caption" color="text.secondary">SELECTED:</Typography>
                                <Typography variant="body2" sx={{ fontWeight: 'bold', mb: 2 }}>{fileToUpload.name}</Typography>
                                <TextField fullWidth size="small" label="ชื่อ Version" value={versionName} onChange={(e) => setVersionName(e.target.value)} sx={{ mb: 2 }} />
                                <Box sx={{ display: 'flex', gap: 1 }}>
                                    <Button variant="contained" fullWidth onClick={handleSaveUpload} disabled={isUploading} sx={{ bgcolor: '#1a237e' }}> บันทึก </Button>
                                    <Button variant="outlined" onClick={() => setFileToUpload(null)}>ยกเลิก</Button>
                                </Box>
                            </Box>
                        )}
                    </Grid>
                    <Grid item sx={{ flexGrow: 1, minWidth: 0 }}>
                        <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>AVAILABLE CAPEX ACTUALS</Typography>
                        <TextField fullWidth size="small" placeholder="ค้นหาข้อมูลจริง CAPEX..." value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)} sx={{ mb: 2, bgcolor: 'white' }} />
                        <VersionList items={activeList} editingId={editingId} newName={newName} onNewNameChange={setNewName} onSave={handleSaveRename} onCancel={() => setEditingId(null)} onStartRename={(id, name) => { setEditingId(id); setNewName(name); }} onDelete={handleDelete} activeId={activeId} />
                    </Grid>
                </Grid>
            </DialogContent>
        </Dialog>
    );
};
