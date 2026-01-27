import React, { createContext, useContext, useState, useMemo, useEffect } from 'react';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { getTheme } from '../utils/theme'; // เรียกฟังก์ชันจากข้อ 1

// สร้าง Context
const ConfigContext = createContext();

// eslint-disable-next-line react-refresh/only-export-components
export const useConfig = () => useContext(ConfigContext);

export const ConfigProvider = ({ children }) => {
  // 1. Force 'light' mode
  const mode = 'light';

  // 2. Disable toggle function
  const toggleColorMode = () => {
    // No-op
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