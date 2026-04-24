import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import mkcert from "vite-plugin-mkcert";

export default defineConfig({
  // base: "/front-budget/",
  plugins: [react(), mkcert()],
  server: {
    port: 3000,
    host: true,
    strictPort: true,
    https: true,
    hmr: {
      clientPort: 3000, // ✅ แก้ปัญหา WebSocket Error
    },
    proxy: {
      "/budget-dash/v1": {
        // แนะนำให้ใช้ IP ตรงๆ หรือ 127.0.0.1 แทน localhost
        target: "https://api.yourdomain.com:443", 
        changeOrigin: true,
        secure: false,
        // เพิ่ม rewrite เพื่อเช็กว่า path ตรงกับที่ Backend ต้องการไหม
        // rewrite: (path) => path.replace(/^\/v1/, '/v1') 
      },
    },
  },
});