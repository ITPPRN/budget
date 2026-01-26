import React from 'react';
import { Box, Grid, Paper, Typography, Button } from '@mui/material';
import { useAuth } from '../hooks/useAuth'; // เรียกใช้ Hook เพื่อดูว่าใครล็อกอินอยู่

const UserManagePage = () => {
  // ดึงข้อมูล User ที่ล็อกอินอยู่มาใช้งาน
  const { user } = useAuth();

  return (
    <Box>
      {/* 1. ส่วนหัวข้อ (Header) */}
      <Box sx={{ mb: 4 }}>
        <Typography variant="h4" sx={{ fontWeight: 'bold', color: 'primary.main' }}>
          จัดการผู้ใช้
        </Typography>
      </Box>
    </Box>
  );
};

export default UserManagePage;