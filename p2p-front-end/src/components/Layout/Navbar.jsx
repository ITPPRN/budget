import React from 'react';
import { AppBar, Toolbar, Typography, Button, IconButton, Box, Avatar } from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import LogoutIcon from '@mui/icons-material/Logout';

// ✅ 1. อย่าลืม Import Component ปุ่มที่เราเพิ่งสร้าง
// (เช็ค path ให้ถูกต้องนะครับ ว่าไฟล์ Navbar อยู่ห่างจาก components แค่ไหน)
import ThemeToggle from '../../components/ThemeToggle'; 

const Navbar = ({ user, onLogout, onToggle }) => {
  return (
    <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
      <Toolbar>
        {/* ปุ่ม Hamburger Menu */}
        <IconButton
          color="inherit"
          aria-label="open drawer"
          edge="start"
          onClick={onToggle}
          sx={{ mr: 2 }}
        >
          <MenuIcon />
        </IconButton>

        {/* ชื่อระบบ */}
        <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1, fontWeight: 'bold' }}>
          P2P Service System
        </Typography>

        {/* โซนขวา: Theme + User + Logout */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            
            {/* ✅ 2. วางปุ่มเปลี่ยนธีมไว้ตรงนี้ครับ */}
            <ThemeToggle />

            {/* ถ้ามี User ให้โชว์ชื่อ */}
            {user && (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Avatar sx={{ bgcolor: 'secondary.main', width: 32, height: 32 }}>
                        {user.username ? user.username.charAt(0).toUpperCase() : 'U'}
                    </Avatar>
                    <Typography variant="subtitle2" sx={{ display: { xs: 'none', sm: 'block' } }}>
                        {user.username || 'ผู้ใช้งาน'}
                    </Typography>
                </Box>
            )}

            <Button 
                color="inherit" 
                onClick={onLogout} 
                endIcon={<LogoutIcon />}
                size="small"
            >
              ออก
            </Button>
        </Box>
      </Toolbar>
    </AppBar>
  );
};

export default Navbar;