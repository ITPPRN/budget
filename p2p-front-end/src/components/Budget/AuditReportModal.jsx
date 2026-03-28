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
    TextField,
    InputAdornment,
    CircularProgress,
    IconButton,
    Chip,
    Paper,
    List,
    ListItem,
    ListItemText,
    Divider
} from '@mui/material';
import ReportProblemIcon from '@mui/icons-material/ReportProblem';
import SearchIcon from '@mui/icons-material/Search';
import AddCircleIcon from '@mui/icons-material/AddCircle';
import DeleteIcon from '@mui/icons-material/Delete';
import InfoIcon from '@mui/icons-material/Info';
import api from '../../utils/api/axiosInstance';

const AuditReportModal = ({ open, onClose, filters, onSubmit, loading: submitting }) => {
    const [searchQuery, setSearchQuery] = useState('');
    const [searchResults, setSearchResults] = useState([]);
    const [searching, setSearching] = useState(false);
    const [addedItems, setAddedItems] = useState([]);

    // Reset when modal opens
    useEffect(() => {
        if (open) {
            setSearchQuery('');
            setSearchResults([]);
            setAddedItems([]);
        }
    }, [open]);

    // Search Logic (Debounced or on change - let's do local search for performance if pre-loaded, 
    // but the senior said "Search to add", implying searching the WHOLE month dataset)
    useEffect(() => {
        const m = filters?.month || (filters?.months && filters.months[0]);
        if (searchQuery.length < 2) {
            setSearchResults([]);
            return;
        }

        const delayDebounceFn = setTimeout(async () => {
            setSearching(true);
            try {
                const res = await api.post(`/budgets/audit/reportable`, {
                    department: filters.department || "",
                    year: filters.year,
                    month: m,
                    conso_gls: filters.conso_gls || [],
                    search: searchQuery // Backend support needed
                });
                // Filter out items already added
                const results = (res.data || []).filter(r => !addedItems.some(a => a.id === r.id));
                setSearchResults(results);
            } catch (err) {
                console.error("Search Error:", err);
            } finally {
                setSearching(false);
            }
        }, 500);

        return () => clearTimeout(delayDebounceFn);
    }, [searchQuery, filters, addedItems]);

    const handleAddItem = (item) => {
        setAddedItems(prev => [...prev, item]);
        setSearchResults(prev => prev.filter(r => r.id !== item.id));
        setSearchQuery(''); // Clear search after adding
    };

    const handleRemoveItem = (id) => {
        setAddedItems(prev => prev.filter(item => item.id !== id));
    };

    const handleSubmit = () => {
        if (addedItems.length === 0) {
            alert("กรุณาเลือกอย่างน้อย 1 รายการเพื่อแจ้งแก้ไข");
            return;
        }
        const ids = addedItems.map(item => item.id);
        onSubmit(ids);
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
            <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1, bgcolor: '#d32f2f', color: 'white' }}>
                <ReportProblemIcon />
                <Typography variant="h6">แจ้งแก้ไขรายการ (Audit Correction Report)</Typography>
            </DialogTitle>
            <DialogContent sx={{ p: 2, bgcolor: '#fcfcfc' }}>
                {/* Search Section */}
                <Box sx={{ mb: 3 }}>
                    <Typography variant="subtitle2" gutterBottom fontWeight="bold" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <SearchIcon fontSize="small" color="primary" /> ค้นหาเลขที่เอกสาร (Search Document)
                    </Typography>
                    <TextField 
                        fullWidth
                        placeholder="พิมพ์เลขที่เอกสาร (Doc No. / Description) ..."
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
                            sx: { bgcolor: 'white', borderRadius: 2 }
                        }}
                    />

                    {/* Search Results Dropdown-like List */}
                    {searchResults.length > 0 && (
                        <Paper elevation={3} sx={{ mt: 1, maxHeight: 300, overflow: 'auto', borderRadius: 2 }}>
                            <List dense>
                                {searchResults.map((item) => (
                                    <React.Fragment key={item.id}>
                                        <ListItem 
                                            button 
                                            onClick={() => handleAddItem(item)}
                                            secondaryAction={
                                                <IconButton edge="end" color="primary">
                                                    <AddCircleIcon />
                                                </IconButton>
                                            }
                                        >
                                            <ListItemText 
                                                primary={<Typography variant="subtitle2"><b>Doc No: {item.doc_no || item.document_no}</b> | {item.gl_account_name}</Typography>}
                                                secondary={`Amount: ${parseFloat(item.amount).toLocaleString()} | Dept: ${item.department} | ${item.description}`}
                                            />
                                        </ListItem>
                                        <Divider />
                                    </React.Fragment>
                                ))}
                            </List>
                        </Paper>
                    )}
                    {searchQuery.length >= 2 && !searching && searchResults.length === 0 && (
                        <Typography variant="caption" color="textSecondary" sx={{ mt: 1, display: 'block' }}>
                            ไม่พบรายการที่ตรงกับ "{searchQuery}" หรือรายการถูกเพิ่มไปแล้ว
                        </Typography>
                    )}
                </Box>

                <Divider sx={{ mb: 3 }} />

                {/* Selected Items Section */}
                <Box sx={{ position: 'relative' }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                        <Typography variant="h6" sx={{ fontWeight: 'bold' }}>
                            รายการที่เลือกแจ้งแก้ไข ({addedItems.length})
                        </Typography>
                        {addedItems.length > 0 && (
                            <Chip label="Clear All" size="small" variant="outlined" onDelete={() => setAddedItems([])} color="error" />
                        )}
                    </Box>

                    {addedItems.length === 0 ? (
                        <Paper variant="outlined" sx={{ p: 5, textAlign: 'center', bgcolor: '#f9f9f9', borderStyle: 'dashed', borderRadius: 3 }}>
                            <InfoIcon sx={{ fontSize: 48, color: '#ccc', mb: 1 }} />
                            <Typography color="textSecondary">ยังไม่มีรายการที่เพิ่ม | ค้นหาและกด + เพื่อเพิ่มชื่อเอกสารลงในใบรายงาน</Typography>
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
                    startIcon={submitting ? <CircularProgress size={20} color="inherit" /> : <ReportProblemIcon />}
                    disabled={submitting || addedItems.length === 0}
                    sx={{ px: 4, borderRadius: 2, textTransform: 'none' }}
                >
                    {submitting ? "กำลังส่ง..." : "แจ้งแก้ไขยอดเงิน (Submit Report)"}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default AuditReportModal;
