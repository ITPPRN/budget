import React, { useState, useEffect, useMemo } from 'react';
import { 
    Dialog, DialogTitle, DialogContent, DialogActions, 
    Box, Typography, TextField, InputAdornment, CircularProgress, 
    Paper, List, ListItem, ListItemText, IconButton, Divider, 
    Chip, TableContainer, Table, TableHead, TableRow, TableCell, TableBody, Button,
    Alert 
} from '@mui/material';
import ReportProblemIcon from '@mui/icons-material/ReportProblem';
import SearchIcon from '@mui/icons-material/Search';
import AddCircleIcon from '@mui/icons-material/AddCircle';
import DeleteIcon from '@mui/icons-material/Delete';
import InfoIcon from '@mui/icons-material/Info';
import CloseIcon from '@mui/icons-material/Close'; 
import api from '../../utils/api/axiosInstance';

// 🌟 เอา initialItems ออกจาก Props
const AuditReportModal = ({ open, onClose, filters, onSubmit, loading: submitting, basketItems = [] }) => {
    const [searchQuery, setSearchQuery] = useState('');
    const [searchResults, setSearchResults] = useState([]); 
    const [searching, setSearching] = useState(false);
    const [addedItems, setAddedItems] = useState([]); // ของที่จะรอใส่ตะกร้า (เฉพาะรอบนี้)
    const [errorMsg, setErrorMsg] = useState(null); 

    // 🌟 Reset ทุกครั้งที่เปิด Modal
    useEffect(() => {
        if (open) {
            setSearchQuery('');
            setSearchResults([]);
            setAddedItems([]); // เคลียร์ของเก่าทิ้ง เริ่มต้นใหม่ทุกครั้งที่กดเปิดค้นหา
            setErrorMsg(null);
        }
    }, [open]);

    // Search Logic
    useEffect(() => {
        const m = filters?.month || (filters?.months && filters.months[0]);
        if (searchQuery.length < 2) {
            setSearchResults([]);
            setErrorMsg(null);
            return;
        }

        const controller = new AbortController(); 

        const delayDebounceFn = setTimeout(async () => {
            setSearching(true);
            setErrorMsg(null);
            try {
                const searchList = searchQuery.includes(',')
                    ? searchQuery.split(',').map(s => s.trim()).filter(s => s.length > 0)
                    : [];

                const res = await api.post(`/budgets/audit/reportable`, {
                    department: filters.department || "",
                    year: filters.year,
                    month: m,
                    conso_gls: filters.conso_gls || [],
                    search: searchQuery, 
                    search_list: searchList
                }, {
                    signal: controller.signal 
                });
                
                setSearchResults(res.data || []);
            } catch (err) {
                if (err.name !== 'CanceledError') {
                    console.error("Search Error:", err);
                    setErrorMsg("เกิดข้อผิดพลาดในการดึงข้อมูล กรุณาลองใหม่อีกครั้ง");
                }
            } finally {
                setSearching(false);
            }
        }, 500);

        return () => {
            clearTimeout(delayDebounceFn);
            controller.abort(); 
        };
    }, [searchQuery, filters]); 

    const handleAddItem = (item) => {
        setAddedItems(prev => [...prev, item]);
    };

    const handleRemoveItem = (id) => {
        setAddedItems(prev => prev.filter(item => item.id !== id));
    };

    const handleSubmit = () => {
        if (addedItems.length === 0) {
            alert("กรุณาเลือกอย่างน้อย 1 รายการเพื่อแจ้งแก้ไข");
            return;
        }
        // ส่งของทั้งหมดที่เลือกรอบนี้ไปให้หน้าหลัก (ActualTable) จับยัดลง DB
        onSubmit(addedItems); 
    };

    const basketIds = useMemo(
        () => new Set((basketItems || []).map(item => item.id)),
        [basketItems]
    );

    const displayResults = useMemo(() => {
        const addedIds = new Set(addedItems.map(item => item.id));
        // ของที่เลือกในรอบนี้แล้ว → ซ่อน (ไปโผล่ในตารางล่างอยู่แล้ว)
        // ของที่อยู่ในตะกร้าอยู่แล้ว → แสดง พร้อม chip "เพิ่มเข้าตะกร้าแล้ว"
        return searchResults
            .filter(result => !addedIds.has(result.id))
            .slice(0, 50);
    }, [searchResults, addedItems]);

    return (
        <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
            <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1, bgcolor: '#d32f2f', color: 'white' }}>
                <ReportProblemIcon />
                <Typography variant="h6">ค้นหาเพื่อแจ้งปฏิเสธรายการ (เฉพาะรอบปัจจุบัน)</Typography>
            </DialogTitle>
            <DialogContent sx={{ p: 2, bgcolor: '#fcfcfc' }}>
                {/* Search Section */}
                <Box sx={{ mb: 3 }}>
                    <Typography variant="subtitle2" gutterBottom fontWeight="bold" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <SearchIcon fontSize="small" color="primary" /> ค้นหาเลขที่เอกสาร (Search Document)
                    </Typography>
                    <TextField 
                        autoFocus 
                        fullWidth
                        placeholder="พิมพ์เลขที่เอกสาร (คั่นด้วยลูกน้ำได้ เช่น 001, 002) ..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        variant="outlined"
                        size="medium"
                        InputProps={{
                            startAdornment: (
                                <InputAdornment position="start">
                                    {searching ? <CircularProgress size={20} /> : <SearchIcon color="action" />}
                                </InputAdornment>
                            ),
                            endAdornment: searchQuery ? ( 
                                <InputAdornment position="end">
                                    <IconButton size="small" onClick={() => setSearchQuery('')}>
                                        <CloseIcon fontSize="small" />
                                    </IconButton>
                                </InputAdornment>
                            ) : null,
                            sx: { bgcolor: 'white', borderRadius: 2 }
                        }}
                    />

                    {errorMsg && (
                        <Alert severity="error" sx={{ mt: 1, borderRadius: 2 }}>{errorMsg}</Alert>
                    )}

                    {/* Search Results */}
                    {displayResults.length > 0 && (
                        <Paper elevation={3} sx={{ mt: 1, maxHeight: 300, overflow: 'auto', borderRadius: 2 }}>
                            <List dense>
                                {displayResults.map((item) => {
                                    const inBasket = basketIds.has(item.id);
                                    return (
                                        <React.Fragment key={item.id}>
                                            <ListItem
                                                button
                                                disabled={inBasket}
                                                onClick={() => !inBasket && handleAddItem(item)}
                                                secondaryAction={
                                                    inBasket ? (
                                                        <Chip
                                                            label="เพิ่มเข้าตะกร้าแล้ว"
                                                            size="small"
                                                            color="success"
                                                            variant="outlined"
                                                        />
                                                    ) : (
                                                        <IconButton edge="end" color="primary">
                                                            <AddCircleIcon />
                                                        </IconButton>
                                                    )
                                                }
                                                sx={inBasket ? { opacity: 0.6, bgcolor: '#f5f5f5' } : undefined}
                                            >
                                                <ListItemText
                                                    primary={<Typography variant="subtitle2"><b>Doc No: {item.doc_no || item.document_no}</b> | {item.gl_account_name}</Typography>}
                                                    secondary={`Amount: ${parseFloat(item.amount).toLocaleString()} | Dept: ${item.department} | ${item.description}`}
                                                />
                                            </ListItem>
                                            <Divider />
                                        </React.Fragment>
                                    );
                                })}
                            </List>
                        </Paper>
                    )}
                    
                    {searchQuery.length >= 2 && !searching && displayResults.length === 0 && !errorMsg && (
                        <Typography variant="caption" color="textSecondary" sx={{ mt: 1, display: 'block' }}>
                            ไม่พบรายการที่ตรงกับ "{searchQuery}" หรือรายการถูกเลือกในรอบนี้แล้ว
                        </Typography>
                    )}
                </Box>

                <Divider sx={{ mb: 3 }} />

                {/* Selected Items Section (โชว์เฉพาะที่เพิ่งกดเพิ่มมาในรอบนี้) */}
                <Box sx={{ position: 'relative' }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                        <Typography variant="h6" sx={{ fontWeight: 'bold' }}>
                            รายการที่จะนำเข้าตะกร้า ({addedItems.length})
                        </Typography>
                        {addedItems.length > 0 && (
                            <Chip label="Clear All" size="small" variant="outlined" onDelete={() => setAddedItems([])} color="error" />
                        )}
                    </Box>

                    {addedItems.length === 0 ? (
                        <Paper variant="outlined" sx={{ p: 5, textAlign: 'center', bgcolor: '#f9f9f9', borderStyle: 'dashed', borderRadius: 3 }}>
                            <InfoIcon sx={{ fontSize: 48, color: '#ccc', mb: 1 }} />
                            <Typography color="textSecondary">ยังไม่ได้เลือกรายการในรอบนี้ | ค้นหาและกด + ด้านบนเพื่อเลือก</Typography>
                        </Paper>
                    ) : (
                        <TableContainer component={Paper} elevation={0} sx={{ border: '1px solid #eee', borderRadius: 2, maxHeight: '40vh' }}>
                            <Table stickyHeader size="small">
                                <TableHead>
                                    <TableRow>
                                        <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Doc No.</TableCell>
                                        <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>GL Account</TableCell>
                                        <TableCell align="right" sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Amount</TableCell>
                                        <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Description</TableCell>
                                        <TableCell align="center" sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Action</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {addedItems.map((item) => (
                                        <TableRow key={item.id} hover>
                                            <TableCell sx={{ fontWeight: 'bold' }}>{item.doc_no || item.document_no}</TableCell>
                                            <TableCell sx={{ fontSize: '0.75rem' }}>{item.conso_gl} - {item.gl_account_name}</TableCell>
                                            <TableCell align="right" sx={{ fontWeight: 'bold', color: item.amount < 0 ? 'error.main' : 'success.main' }}>
                                                {parseFloat(item.amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                                            </TableCell>
                                            <TableCell sx={{ fontSize: '0.75rem' }}>{item.description}</TableCell>
                                            <TableCell align="center">
                                                <IconButton size="small" color="error" onClick={() => handleRemoveItem(item.id)}>
                                                    <DeleteIcon fontSize="small" />
                                                </IconButton>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </TableContainer>
                    )}
                </Box>
            </DialogContent>
            <DialogActions sx={{ p: 2, bgcolor: '#f9f9f9', borderTop: '1px solid #eee' }}>
                <Button onClick={onClose} disabled={submitting} sx={{ textTransform: 'none' }}>ยกเลิก</Button>
                <Box sx={{ flex: 1 }} />
                <Button 
                    onClick={handleSubmit} 
                    variant="contained" 
                    color="error" 
                    size="large"
                    startIcon={submitting ? <CircularProgress size={20} color="inherit" /> : <AddCircleIcon />}
                    disabled={submitting || addedItems.length === 0}
                    sx={{ px: 4, borderRadius: 2, textTransform: 'none' }}
                >
                    {submitting ? "กำลังบันทึก..." : "ส่งเข้าตะกร้าปฏิเสธ"}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default AuditReportModal;