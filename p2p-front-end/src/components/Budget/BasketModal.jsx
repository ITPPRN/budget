import React from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    Box,
    Typography,
    IconButton,
    TableContainer,
    Table,
    TableHead,
    TableRow,
    TableCell,
    TableBody,
    Paper,
    Button
} from '@mui/material';
import ShoppingBasketIcon from '@mui/icons-material/ShoppingBasket';
import CloseIcon from '@mui/icons-material/Close';
import DeleteIcon from '@mui/icons-material/Delete';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';

const BasketModal = ({ 
    open, 
    onClose, 
    basketItems = [], 
    onRemove, 
    onApprove, 
    loading 
}) => {
    return (
        <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
            <DialogTitle sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', bgcolor: '#f5f5f5' }}>
                <Typography variant="h6" fontWeight="bold" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <ShoppingBasketIcon color="error" /> ตะกร้ารายการปฏิเสธ ({basketItems.length})
                </Typography>
                <IconButton onClick={onClose} size="small" disabled={loading}>
                    <CloseIcon />
                </IconButton>
            </DialogTitle>
            
            <DialogContent sx={{ p: 2 }}>
                {basketItems.length === 0 ? (
                    <Box sx={{ p: 4, textAlign: 'center' }}>
                        <Typography color="textSecondary">ตะกร้าว่างเปล่า</Typography>
                    </Box>
                ) : (
                    <TableContainer component={Paper} elevation={0} sx={{ border: '1px solid #eee' }}>
                        <Table size="small">
                            <TableHead>
                                <TableRow sx={{ bgcolor: '#fafafa' }}>
                                    <TableCell sx={{ fontWeight: 'bold' }}>Doc No.</TableCell>
                                    <TableCell sx={{ fontWeight: 'bold' }}>GL Account</TableCell>
                                    <TableCell align="right" sx={{ fontWeight: 'bold' }}>Amount</TableCell>
                                    <TableCell align="center" sx={{ fontWeight: 'bold' }}>Action</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {basketItems.map((item) => (
                                    <TableRow key={item.id} hover>
                                        <TableCell>{item.doc_no || item.document_no}</TableCell>
                                        <TableCell>{item.conso_gl} - {item.gl_account_name}</TableCell>
                                        <TableCell align="right" sx={{ color: item.amount < 0 ? 'error.main' : 'success.main', fontWeight: 'bold' }}>
                                            {parseFloat(item.amount).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                                        </TableCell>
                                        <TableCell align="center">
                                            <IconButton 
                                                size="small" 
                                                color="error" 
                                                onClick={() => onRemove(item.id)}
                                                title="ลบออกจากตะกร้า"
                                                disabled={loading}
                                            >
                                                <DeleteIcon fontSize="small" />
                                            </IconButton>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </TableContainer>
                )}
            </DialogContent>
            
            {basketItems.length > 0 && (
                <Box sx={{ p: 2, borderTop: '1px solid #eee', bgcolor: '#fafafa', textAlign: 'right' }}>
                    <Button 
                        variant="contained" 
                        color="success" 
                        startIcon={<CheckCircleIcon />}
                        onClick={onApprove}
                        disabled={loading}
                        sx={{ textTransform: 'none' }}
                    >
                        {loading ? "กำลังประมวลผล..." : "ยืนยันและทำรายการ"}
                    </Button>
                </Box>
            )}
        </Dialog>
    );
};

export default BasketModal;