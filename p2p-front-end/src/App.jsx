import React, { useEffect } from "react";
import { BrowserRouter, useNavigate } from "react-router-dom";
import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";
import { StyledEngineProvider, CssBaseline } from "@mui/material"; // เพิ่ม CssBaseline
import { ConfigProvider as AntConfigProvider } from "antd";

// import Theme from "./utils/theme"; ❌ ไม่ใช้แล้ว เพราะไปอยู่ใน ConfigContext
import { ConfigProvider } from "./contexts/ConfigContext"; // ✅ เรียก Context ที่เราเพิ่งทำ
import { AuthProvider } from "./hooks/useAuth";
import ThemeRoutes from "./routes";
import ErrorBoundary from "./components/ErrorBoundary";

// หลัง re-login เสร็จ ถ้ามี return_to ใน sessionStorage → navigate ไปหน้านั้น
function ReturnToHandler() {
  const navigate = useNavigate();
  useEffect(() => {
    const returnTo = sessionStorage.getItem('oidc_return_to');
    if (returnTo && returnTo !== window.location.pathname + window.location.search) {
      sessionStorage.removeItem('oidc_return_to');
      navigate(returnTo, { replace: true });
    } else if (returnTo) {
      sessionStorage.removeItem('oidc_return_to');
    }
  }, [navigate]);
  return null;
}

function App() {
  return (
    <StyledEngineProvider injectFirst>
      <AntConfigProvider theme={{ token: { fontFamily: '"Kanit", sans-serif' } }}>

        {/* ✅ 1. ใช้ ConfigProvider เป็นตัวจัดการ Theme */}
        <ConfigProvider>
          {/* ✅ CssBaseline ช่วยรีเซ็ตสีพื้นหลังให้เป็น Dark/Light ตาม Theme */}
          <CssBaseline />

          <BrowserRouter>
            <ErrorBoundary>
              <AuthProvider>
                <ReturnToHandler />
                <ThemeRoutes />
              </AuthProvider>
            </ErrorBoundary>
          </BrowserRouter>

          <ToastContainer position="top-right" autoClose={3000} />
        </ConfigProvider>

      </AntConfigProvider>
    </StyledEngineProvider>
  );
}

export default App;