import React from 'react';
import { Box, Grid, Paper, Typography, Button } from '@mui/material';
import { useAuth } from '../hooks/useAuth'; // เรียกใช้ Hook เพื่อดูว่าใครล็อกอินอยู่

const HomePage = () => {
  // ดึงข้อมูล User ที่ล็อกอินอยู่มาใช้งาน
  const { user } = useAuth();

  return (
    <Box>
      {/* 1. ส่วนหัวข้อ (Header) */}
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" sx={{ fontWeight: 'bold', color: 'primary.main' }}>
          ยินดีต้อนรับ, {user?.username || 'ผู้ใช้งาน'} 👋
        </Typography>
        <Typography variant="body1" color="text.secondary">
          นี่คือภาพรวมระบบ P2P Service ของคุณในวันนี้
        </Typography>
      </Box>

      {/* 2. ส่วนแสดงข้อมูลสรุป (Stats Cards) */}
      <Grid container spacing={3}>
        {/* Card 1: ยอดขาย */}
        <Grid item xs={12} sm={6} md={3}>
          <Paper elevation={3} sx={{ p: 3, borderRadius: 2, height: '100%', borderLeft: '5px solid #1976d2' }}>
            <Typography variant="subtitle2" color="text.secondary">ยอดขายรวม</Typography>
            <Typography variant="h4" sx={{ my: 1, fontWeight: 'bold' }}>฿ 1,250,000</Typography>
            <Typography variant="caption" color="success.main">+15% จากเดือนที่แล้ว</Typography>
          </Paper>
        </Grid>

        {/* Card 2: ลูกค้าใหม่ */}
        <Grid item xs={12} sm={6} md={3}>
          <Paper elevation={3} sx={{ p: 3, borderRadius: 2, height: '100%', borderLeft: '5px solid #2e7d32' }}>
            <Typography variant="subtitle2" color="text.secondary">ลูกค้าใหม่</Typography>
            <Typography variant="h4" sx={{ my: 1, fontWeight: 'bold' }}>34 ราย</Typography>
            <Typography variant="caption" color="text.secondary">อัปเดตล่าสุดเมื่อสักครู่</Typography>
          </Paper>
        </Grid>

        {/* Card 3: รออนุมัติ */}
        <Grid item xs={12} sm={6} md={3}>
          <Paper elevation={3} sx={{ p: 3, borderRadius: 2, height: '100%', borderLeft: '5px solid #ed6c02' }}>
            <Typography variant="subtitle2" color="text.secondary">รอการอนุมัติ</Typography>
            <Typography variant="h4" sx={{ my: 1, fontWeight: 'bold' }}>12 รายการ</Typography>
            <Button size="small" sx={{ mt: 1 }}>ดูรายการทั้งหมด</Button>
          </Paper>
        </Grid>

        {/* Card 4: งานที่ต้องทำ */}
        <Grid item xs={12} sm={6} md={3}>
          <Paper elevation={3} sx={{ p: 3, borderRadius: 2, height: '100%', borderLeft: '5px solid #d32f2f' }}>
            <Typography variant="subtitle2" color="text.secondary">งานคงค้าง</Typography>
            <Typography variant="h4" sx={{ my: 1, fontWeight: 'bold' }}>5 งาน</Typography>
            <Typography variant="caption" color="error.main">ครบกำหนดภายใน 3 วัน</Typography>
          </Paper>
        </Grid>
      </Grid>

      {/* 3. ส่วนเนื้อหาเพิ่มเติม (เช่น กราฟ หรือ ตารางล่าสุด) */}
      <Box sx={{ mt: 4 }}>
        <Paper elevation={2} sx={{ p: 3, borderRadius: 2, minHeight: '300px' }}>
          <Typography variant="h6" sx={{ mb: 2 }}>
            📊 กราฟแสดงผลการดำเนินงาน (พื้นที่สำหรับใส่ Chart)
          </Typography>
          
          <Box sx={{ 
            height: '200px', 
            bgcolor: '#f5f5f5', 
            borderRadius: 1, 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center',
            color: '#999' 
          }}>
            [ พื้นที่สำหรับใส่ ApexCharts หรือ Chart.js ]
          </Box>
        </Paper>
      </Box>
    </Box>
  );
};

export default HomePage;