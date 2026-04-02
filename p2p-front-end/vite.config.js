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
        target: "http://localhost:8000", // Go Backend
        changeOrigin: true,
        secure: false,
      },
    },
  },
});
