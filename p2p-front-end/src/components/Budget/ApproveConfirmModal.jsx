import React from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Typography,
    Box,
    Button
} from '@mui/material';

const ApproveConfirmModal = ({ 
    open, 
    onClose, 
    onConfirm, 
    loading, 
    filters, 
    basketCount 
}) => {

    // ตัวช่วยแปลงเลขเดือนเป็นชื่อเดือนภาษาไทย
    const getDisplayMonthYear = () => {
        const y = filters?.year || '';
        let m = filters?.month || (filters?.months && filters.months[0]) || '';
        
        if (!y || !m) return 'ไม่ระบุเดือน/ปี';
    
        // แปลงให้เป็นตัวพิมพ์ใหญ่ เพื่อรองรับ "JAN", "Jan", "jan"
        const monthKey = String(m).toUpperCase();
    
        const monthNames = {
            // รองรับแบบตัวเลข (เผื่อไว้)
            "01": "JAN", "02": "FEB", "03": "MAR", "04": "APR",
            "05": "MAY", "06": "JUN", "07": "JUL", "08": "AUG",
            "09": "SEP", "10": "OCT", "11": "NOV", "12": "DEC",
            
            // 🌟 เพิ่มการรองรับแบบชื่อเดือนภาษาอังกฤษ (ตามที่ Debug เจอ)
            "JAN": "JAN", "FEB": "FEB", "MAR": "MAR", "APR": "APR",
            "MAY": "MAY", "JUN": "JUN", "JUL": "JUL", "AUG": "AUG",
            "SEP": "SEP", "OCT": "OCT", "NOV": "NOV", "DEC": "DEC"
        };
    
        const monthName = monthNames[monthKey] || monthKey;
    
        return `${monthName} ${y}`;
    };

    return (
        <Dialog open={open} onClose={!loading ? onClose : undefined} maxWidth="sm" fullWidth>
            <DialogTitle sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold' }}>
                ยืนยันการทำรายการ
            </DialogTitle>
            
            <DialogContent sx={{ mt: 3 }}>
                <Box sx={{ textAlign: 'center', mb: 3, p: 2, bgcolor: '#f5f5f5', borderRadius: 2 }}>
                    <Typography variant="body2" color="text.secondary">
                        กำลังดำเนินการตรวจสอบข้อมูลประจำเดือน
                    </Typography>
                    <Typography variant="h5" color="primary" fontWeight="bold" sx={{ mt: 1 }}>
                        {getDisplayMonthYear()}
                    </Typography>
                </Box>
                
                <Typography variant="body1" textAlign="center" sx={{ mb: 2 }}>
                    {basketCount > 0 
                        ? <>คุณมี <b>{basketCount} รายการ</b> ในตะกร้าที่จะถูกปฏิเสธกลับไป<br/>และรายการที่เหลือทั้งหมดจะถูก <b>ยืนยันความถูกต้องอัตโนมัติ</b></>
                        : <>คุณกำลัง <b>ยืนยันความถูกต้องของข้อมูลทั้งหมด (Approve All)</b><br/>สำหรับเดือนนี้โดยไม่มีรายการปฏิเสธ</>
                    }
                </Typography>

                <Typography variant="body2" color="error" textAlign="center" fontWeight="bold">
                    *โปรดตรวจสอบความถูกต้องก่อนกดยืนยัน*
                </Typography>
            </DialogContent>

            <DialogActions sx={{ p: 2, justifyContent: 'center', gap: 2 }}>
                <Button 
                    onClick={onClose} 
                    color="inherit" 
                    variant="outlined" 
                    disabled={loading}
                    sx={{ minWidth: 120 }}
                >
                    ยกเลิก
                </Button>
                <Button 
                    onClick={onConfirm} 
                    color="success" 
                    variant="contained" 
                    disabled={loading} 
                    sx={{ minWidth: 120 }}
                >
                    {loading ? "กำลังประมวลผล..." : "ยืนยันดำเนินการ"}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default ApproveConfirmModal;