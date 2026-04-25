import React, { useState } from 'react';
import { Box, Toolbar } from '@mui/material';
import { Outlet } from 'react-router-dom';
import MonitorHeartIcon from '@mui/icons-material/MonitorHeart';
import Navbar from '../components/Layout/Navbar'; // แยกไปสร้างเหมือน Sidebar
import Sidebar from '../components/Layout/sidebar';
import { useAuth } from '../hooks/useAuth';
import { MENU_ITEMS, OWNER_MENU_ITEMS } from '../config/menuConfig'; // ดึง Config มาใช้

export default function MainLayout() {
  const { user, logout } = useAuth(); // เรียกใช้ Logic สั้นๆ
  const [isSidebarOpen, setSidebarOpen] = useState(true);

  // เลือกเมนูตาม Role
  const isAdmin = user?.roles && user.roles.some(r => r.toUpperCase().includes('ADMIN'));
  const isOwner = user?.roles && user.roles.some(r => ['OWNER', 'DELEGATE'].some(role => r.toUpperCase().includes(role)));

  // Logic: ถ้าเป็น Admin ให้เห็นเมนู Admin (แม้จะเป็น Owner ด้วย)
  // ถ้าเป็น Owner (และไม่ใช่ Admin) หรือมี department -> เห็นเมนู Owner
  const showOwnerMenu = !isAdmin && (isOwner || !!user?.department_code || !!user?.department);

  const baseMenu = showOwnerMenu ? OWNER_MENU_ITEMS : MENU_ITEMS;

  // เมนู Sync Monitor — แสดงเฉพาะ username = "admin" เท่านั้น
  const isSyncAdmin = user?.username === 'admin';
  const menuItems = isSyncAdmin
    ? [...baseMenu, { title: 'Sync Monitor', path: '/sync-monitor', icon: <MonitorHeartIcon /> }]
    : baseMenu;

  return (
    <Box sx={{ display: 'flex', width: '100vw', minHeight: '100vh', overflow: 'hidden' }}>
      <Navbar
        user={user}
        onLogout={logout}
        onToggle={() => setSidebarOpen(!isSidebarOpen)}
      />

      <Sidebar
        isOpen={isSidebarOpen}
        menuItems={menuItems} // ส่งรายการเมนูที่เลือกแล้วเข้าไป
      />

      <Box
        component="main"
        sx={{
          flex: 1,
          minWidth: 0,
          display: 'flex',
          flexDirection: 'column',
          width: '100% !important',
          maxWidth: 'none !important',
          bgcolor: '#f8f9fc',
          overflowY: 'auto',
          overflowX: 'hidden'
        }}
      >
        <Toolbar /> {/* ดัน Content ลงมา */}
        <Box sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', width: '100% !important', maxWidth: 'none !important' }}>
          <Outlet />  {/* เนื้อหาเปลี่ยนไปตาม Route */}
        </Box>
      </Box>
    </Box>
  );
}