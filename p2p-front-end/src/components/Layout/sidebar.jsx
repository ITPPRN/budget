import React from 'react';
import { Drawer, List, ListItem, ListItemIcon, ListItemText, Collapse, Toolbar } from '@mui/material';
import { useNavigate, useLocation } from 'react-router-dom';
import ExpandLess from '@mui/icons-material/ExpandLess';
import ExpandMore from '@mui/icons-material/ExpandMore';

// กำหนดความกว้างของ Sidebar
const drawerWidth = 240;

const Sidebar = ({ isOpen, menuItems }) => {
  const navigate = useNavigate();
  const location = useLocation(); // เอาไว้เช็คว่าอยู่หน้าไหน จะได้ Highlight เมนูถูก
  const [openSub, setOpenSub] = React.useState({});

  const handleSubMenu = (title) => {
    setOpenSub({ ...openSub, [title]: !openSub[title] });
  };

  // ฟังก์ชันเช็คว่าเมนูนี้ Active อยู่หรือไม่ (เพื่อเปลี่ยนสีพื้นหลัง)
  const isSelected = (path) => location.pathname === path;

  // ฟังก์ชันสร้างเมนูแบบ Recursive
  const renderMenu = (items) => {
    return items.map((item) => {
      const active = item.path ? isSelected(item.path) : false;
      const hasChildren = item.children && item.children.length > 0;
      const expanded = openSub[item.title];

      return (
        <React.Fragment key={item.title}>
          <ListItem
            button
            onClick={() => hasChildren ? handleSubMenu(item.title) : navigate(item.path)}
            selected={active}
            sx={{
              mx: 1, // Add horizontal margin for "pill" look
              my: 0.5, // Add vertical spacing
              borderRadius: 2, // Rounded corners
              transition: 'all 0.2s',
              // Active State
              '&.Mui-selected': {
                backgroundColor: 'primary.main',
                color: 'primary.contrastText',
                boxShadow: '0 4px 12px rgba(25, 118, 210, 0.2)', // Add depth
                '&:hover': {
                  backgroundColor: 'primary.dark',
                },
                // Icon Color in Active State
                '& .MuiListItemIcon-root': {
                  color: 'inherit', // Follow text color (white)
                }
              },
              // Hover State (Inactive)
              '&:hover': {
                backgroundColor: 'action.hover',
                transform: 'translateX(4px)', // Subtle movement
              }
            }}
          >
            {item.icon && (
              <ListItemIcon sx={{
                color: active ? 'inherit' : 'text.secondary',
                minWidth: 40 // Adjust icon spacing
              }}>
                {item.icon}
              </ListItemIcon>
            )}
            <ListItemText
              primary={item.title}
              primaryTypographyProps={{
                fontWeight: active ? 'bold' : 'medium',
                variant: 'body2',
                fontSize: '0.95rem'
              }}
            />
            {hasChildren && (
              expanded ? <ExpandLess sx={{ color: active ? 'inherit' : 'action.active' }} /> : <ExpandMore sx={{ color: active ? 'inherit' : 'action.active' }} />
            )}
          </ListItem>

          {/* ส่วนแสดงเมนูย่อย */}
          {hasChildren && (
            <Collapse in={expanded} timeout="auto" unmountOnExit>
              <List component="div" disablePadding sx={{ py: 0.5 }}>
                {/* Indent children */}
                {renderMenu(item.children)}
              </List>
            </Collapse>
          )}
        </React.Fragment>
      );
    });
  };

  return (
    <Drawer
      variant="permanent"
      open={isOpen}
      sx={{
        width: isOpen ? drawerWidth : 0,
        flexShrink: 0,
        whiteSpace: 'nowrap',
        boxSizing: 'border-box',
        transition: 'width 0.3s cubic-bezier(0.4, 0, 0.2, 1)', // Smoother bezier
        [`& .MuiDrawer-paper`]: {
          width: isOpen ? drawerWidth : 0,
          boxSizing: 'border-box',
          transition: 'width 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          overflowX: 'hidden',
          borderRight: 'none', // Remove default border
          boxShadow: '4px 0 24px rgba(0,0,0,0.05)', // Soft shadow instead
          backgroundColor: '#ffffff', // Clean white background
          backgroundImage: 'linear-gradient(180deg, #ffffff 0%, #f8f9fa 100%)' // Subtle gradient
        },
      }}
    >
      <Toolbar /> {/* Spacer taking Navbar height */}
      <List sx={{ pt: 2 }}>
        {renderMenu(menuItems)}
      </List>
    </Drawer>
  );
};

export default Sidebar;