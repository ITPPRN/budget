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
    return items.map((item) => (
      <React.Fragment key={item.title}>
        <ListItem 
          button 
          onClick={() => item.children ? handleSubMenu(item.title) : navigate(item.path)}
          selected={item.path ? isSelected(item.path) : false} // Highlight เมนูที่เลือก
          sx={{
            // Style เพิ่มเติมเมื่อเมนูถูกเลือก
            '&.Mui-selected': {
              backgroundColor: 'rgba(25, 118, 210, 0.08)',
              borderLeft: '4px solid #1976d2',
              '&:hover': { backgroundColor: 'rgba(25, 118, 210, 0.12)' }
            }
          }}
        >
          {item.icon && <ListItemIcon sx={{ color: item.path && isSelected(item.path) ? '#1976d2' : 'inherit' }}>{item.icon}</ListItemIcon>}
          <ListItemText primary={item.title} />
          {item.children && (openSub[item.title] ? <ExpandLess /> : <ExpandMore />)}
        </ListItem>
        
        {/* ส่วนแสดงเมนูย่อย */}
        {item.children && (
          <Collapse in={openSub[item.title]} timeout="auto" unmountOnExit>
            <List component="div" disablePadding sx={{ pl: 4 }}>
              {renderMenu(item.children)}
            </List>
          </Collapse>
        )}
      </React.Fragment>
    ));
  };

  return (
    <Drawer
      variant="permanent"
      open={isOpen}
      sx={{
        width: isOpen ? drawerWidth : 0, // ถ้าปิด Sidebar ให้ความกว้างเป็น 0 (หรือจะทำเป็น Mini drawer ก็ได้)
        flexShrink: 0,
        whiteSpace: 'nowrap',
        boxSizing: 'border-box',
        transition: 'width 0.3s', // ใส่ Animation ให้ลื่นๆ
        [`& .MuiDrawer-paper`]: { 
          width: isOpen ? drawerWidth : 0, 
          boxSizing: 'border-box',
          transition: 'width 0.3s',
          overflowX: 'hidden'
        },
      }}
    >
      <Toolbar /> {/* ดันเมนูลงมาให้พ้น Navbar ด้านบน */}
      <List>
        {renderMenu(menuItems)}
      </List>
    </Drawer>
  );
};

export default Sidebar;