import React, { useState } from 'react';
import { Box, Toolbar } from '@mui/material';
import { Outlet } from 'react-router-dom';
import Navbar from '../components/Layout/Navbar'; // แยกไปสร้างเหมือน Sidebar
import Sidebar from '../components/Layout/sidebar';
import { useAuth } from '../hooks/useAuth';
import { MENU_ITEMS } from '../config/menuConfig'; // ดึง Config มาใช้

export default function MainLayout() {
  const { user, logout } = useAuth(); // เรียกใช้ Logic สั้นๆ
  const [isSidebarOpen, setSidebarOpen] = useState(true);

  return (
    <Box sx={{ display: 'flex' }}>
      {/* ส่ง Props เข้าไป ไม่ต้องเขียน Logic รกๆ ตรงนี้ */}
      <Navbar 
        user={user} 
        onLogout={logout} 
        onToggle={() => setSidebarOpen(!isSidebarOpen)} 
      />
      
      <Sidebar 
        isOpen={isSidebarOpen} 
        menuItems={MENU_ITEMS} // ส่งรายการเมนูเข้าไป
      />

      <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
        <Toolbar /> {/* ดัน Content ลงมา */}
        <Outlet />  {/* เนื้อหาเปลี่ยนไปตาม Route */}
      </Box>
    </Box>
  );
}