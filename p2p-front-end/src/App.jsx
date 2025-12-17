import React from "react";
import { BrowserRouter } from "react-router-dom";
import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { StyledEngineProvider, CssBaseline } from "@mui/material"; // เพิ่ม CssBaseline
import { ConfigProvider as AntConfigProvider } from "antd";

// import Theme from "./utils/theme"; ❌ ไม่ใช้แล้ว เพราะไปอยู่ใน ConfigContext
import { ConfigProvider } from "./contexts/ConfigContext"; // ✅ เรียก Context ที่เราเพิ่งทำ
import { AuthProvider } from "./hooks/useAuth"; 
import ThemeRoutes from "./routes";

function App() {
  return (
    <StyledEngineProvider injectFirst>
      <AntConfigProvider theme={{ token: { fontFamily: '"Kanit", sans-serif' } }}>
        
        {/* ✅ 1. ใช้ ConfigProvider เป็นตัวจัดการ Theme */}
        <ConfigProvider>
          {/* ✅ CssBaseline ช่วยรีเซ็ตสีพื้นหลังให้เป็น Dark/Light ตาม Theme */}
          <CssBaseline /> 
          
          <AuthProvider>
            <BrowserRouter>
              <ThemeRoutes />
            </BrowserRouter>
          </AuthProvider>
          
          <ToastContainer position="top-right" autoClose={3000} />
        </ConfigProvider>

      </AntConfigProvider>
    </StyledEngineProvider>
  );
}

export default App;