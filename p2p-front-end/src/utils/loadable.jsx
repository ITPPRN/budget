import React, { Suspense } from 'react';
import { Box, CircularProgress } from '@mui/material';

// ✅ เพิ่มบรรทัดนี้: ปิดการตรวจชื่อ Display Name สำหรับไฟล์นี้
// eslint-disable-next-line react/display-name
const Loadable = (Component) => (props) => ( // 👈 เช็คตรงนี้! ต้องมีคำว่า Component ในวงเล็บ
  <Suspense 
    fallback={
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress />
      </Box>
    }
  >
    {/* ถ้าข้างบนมี (Component) ตรงนี้จะทำงานได้ถูกต้อง */}
    <Component {...props} />
  </Suspense>
);

export default Loadable;