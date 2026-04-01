import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: true,
    proxy: {
      // เมื่อไหร่ที่ Frontend ยิงไปที่ /api
      "/v1": {
        target: "https://dzv6zh3b-8000.asse.devtunnels.ms/", // 👉 ส่งต่อไปหา Go Backend ที่พอร์ต 8000
        changeOrigin: true,
        secure: false,
      },
    },
  },
});
