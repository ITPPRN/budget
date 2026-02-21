import React, { useState } from 'react';
import { Box, Toolbar } from '@mui/material';
import { Outlet } from 'react-router-dom';
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

  const menuItems = showOwnerMenu ? OWNER_MENU_ITEMS : MENU_ITEMS;

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
        menuItems={menuItems} // ส่งรายการเมนูที่เลือกแล้วเข้าไป
      />

      <Box component="main" sx={{ flexGrow: 1, minWidth: 0, width: '100%', overflowX: 'hidden' }}>
        <Toolbar /> {/* ดัน Content ลงมา */}
        <Outlet />  {/* เนื้อหาเปลี่ยนไปตาม Route */}
      </Box>
    </Box>
  );
}