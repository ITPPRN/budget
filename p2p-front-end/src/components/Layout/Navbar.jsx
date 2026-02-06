import React from 'react';
import { AppBar, Toolbar, Typography, Button, IconButton, Box, Avatar } from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import LogoutIcon from '@mui/icons-material/Logout';

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
        <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1, fontWeight: 'bold', color: 'white' }}>
          Budget Service System
        </Typography>

        {/* โซนขวา: Theme + User + Logout */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>

          {/* ถ้ามี User ให้โชว์ชื่อ */}
          {user && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <Avatar sx={{ bgcolor: 'secondary.main', width: 32, height: 32, }}>
                {/* JSON from Backend: { userName, name, ... } */}
                {(user.name || user.userName || 'U').charAt(0).toUpperCase()}
              </Avatar>
              <Typography variant="subtitle2" sx={{ display: { xs: 'none', sm: 'block', color: 'white' } }}>
                {user.name || user.userName || 'ผู้ใช้งาน'}
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