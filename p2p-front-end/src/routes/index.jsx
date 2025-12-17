import { useRoutes } from 'react-router-dom';
import { Box, CircularProgress } from '@mui/material';
import Routes from './routes';
import { useAuth } from '../hooks/useAuth';

export default function ThemeRoutes() {
    const { user, isLoading } = useAuth();
    const isLoggedIn = !!user;

    // ✅ แก้ไข: เรียก Hook ให้เสร็จก่อนเสมอ (ห้ามมี if/return มาคั่นก่อนบรรทัดนี้)
    const routing = useRoutes(Routes(isLoggedIn));

    // ⏳ ค่อยมาเช็ค Loading ตรงนี้
    if (isLoading) {
        return (
            <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
                <CircularProgress />
            </Box>
        );
    }

    // ส่ง Routing ที่เตรียมไว้กลับไป
    return routing;
}