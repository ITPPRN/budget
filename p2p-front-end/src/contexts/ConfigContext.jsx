import React, { createContext, useContext, useState, useMemo, useEffect } from 'react';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { getTheme } from '../utils/theme'; // เรียกฟังก์ชันจากข้อ 1

// สร้าง Context
const ConfigContext = createContext();

// eslint-disable-next-line react-refresh/only-export-components
export const useConfig = () => useContext(ConfigContext);

export const ConfigProvider = ({ children }) => {
  // 1. อ่านค่าจาก LocalStorage ก่อน (ถ้าไม่มีให้เป็น 'light')
  const [mode, setMode] = useState(localStorage.getItem('themeMode') || 'light');

  // 2. ฟังก์ชันสลับโหมด
  const toggleColorMode = () => {
    setMode((prevMode) => {
      const newMode = prevMode === 'light' ? 'dark' : 'light';
      localStorage.setItem('themeMode', newMode); // บันทึกลงเครื่อง
      return newMode;
    });
  };

  // 3. สร้าง Theme Object จริงๆ จากโหมดปัจจุบัน
  // useMemo ช่วยไม่ให้สร้าง Theme ใหม่พร่ำเพรื่อถ้า mode ไม่เปลี่ยน
  const theme = useMemo(() => createTheme(getTheme(mode)), [mode]);

  return (
    <ConfigContext.Provider value={{ mode, toggleColorMode }}>
      {/* ส่ง ThemeProvider ให้ลูกหลานใช้ตรงนี้เลยก็ได้ */}
      <ThemeProvider theme={theme}>
        {children}
      </ThemeProvider>
    </ConfigContext.Provider>
  );
};