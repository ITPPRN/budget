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
          {user && (() => {
            const roles = user.roles || [];
            let displayRole = 'User';
            if (roles.some(r => r.toUpperCase().includes('ADMIN'))) displayRole = 'Admin';
            else if (roles.some(r => r.toUpperCase().includes('OWNER'))) displayRole = 'Owner';
            else if (roles.some(r => r.toUpperCase().includes('DELEGATE'))) displayRole = 'Delegate';

            return (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                <Box sx={{ display: { xs: 'none', sm: 'flex' }, flexDirection: 'column', alignItems: 'flex-end' }}>
                  <Typography variant="subtitle2" sx={{ fontWeight: 700, color: 'white', lineHeight: 1.2 }}>
                    {user.name || user.userName || 'ผู้ใช้งาน'} | {displayRole}
                  </Typography>
                  <Typography variant="caption" sx={{ color: 'rgba(255,255,255,0.85)', fontWeight: 600, fontSize: '0.7rem', letterSpacing: 0.5 }}>
                    {user.department_code || '-'}
                  </Typography>
                </Box>
                <Avatar sx={{ bgcolor: 'secondary.main', width: 34, height: 34, fontSize: '0.9rem', fontWeight: 'bold', border: '1px solid rgba(255,255,255,0.2)' }}>
                  {(user.name || user.userName || 'U').charAt(0).toUpperCase()}
                </Avatar>
              </Box>
            );
          })()}

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