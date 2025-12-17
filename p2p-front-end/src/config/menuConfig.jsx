import HomeIcon from '@mui/icons-material/Home';
import SettingsIcon from '@mui/icons-material/Settings';
import PersonIcon from '@mui/icons-material/Person';

// กำหนดรายการเมนูที่นี่ที่เดียว
export const MENU_ITEMS = [
  {
    title: 'หน้าหลัก',
    path: '/home',
    icon: <HomeIcon />,
  },
  {
    title: 'กลุ่มลูกค้าคาดหวัง',
    path: '/prospect',
    icon: <PersonIcon />,
  },
  {
    title: 'จัดการข้อมูล',
    icon: <SettingsIcon />,
    children: [ // เมนูย่อย
      { title: 'ใบเสนอราคา', path: '/manage/quotation' },
      { title: 'ข้อมูลลูกค้า', path: '/manage/customer' },
    ]
  }
];