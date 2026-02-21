import HomeIcon from '@mui/icons-material/Home';
import SettingsIcon from '@mui/icons-material/Settings';
import PersonIcon from '@mui/icons-material/Person';
import StepIcon from '@mui/material/StepIcon';
import BusinessCenterIcon from '@mui/icons-material/BusinessCenter';
import WorkIcon from '@mui/icons-material/Work'

// กำหนดรายการเมนูที่นี่ที่เดียว
export const MENU_ITEMS = [
  {
    title: 'Dashboard (ADMIN)',
    path: '/home',
    icon: <HomeIcon />,
  },
  {
    title: 'Dashboard (OWNER)',
    path: '/owner/dashboard',
    icon: <HomeIcon />,
  },
  {
    title: 'Detail Report',
    path: '/detail',
    icon: <BusinessCenterIcon />,
  },
  {
    title: 'User Management',
    path: '/user',
    icon: <PersonIcon />
  },
  {
    title: 'Data Management',
    path: '/data',
    icon: <SettingsIcon />
  }
];

export const OWNER_MENU_ITEMS = [
  {
    title: 'Dashboard',
    path: '/owner/dashboard',
    icon: <HomeIcon />,
  },
  {
    title: 'Detail Report',
    path: '/owner/detail',
    icon: <BusinessCenterIcon />,
  },
  {
    title: 'User Management',
    path: '/owner/user',
    icon: <PersonIcon />
  }
];