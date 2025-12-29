import axios from 'axios';
import { toast } from 'react-toastify';

// 1. ฟังก์ชันช่วยแกะ Cookie
function getCookie(name) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(';').shift();
}

const api = axios.create({
  baseURL: '/v1', // ⚠️ เช็ค baseURL ให้ตรงกับที่ใช้ 
  withCredentials: true, // สำคัญมาก! เพื่อส่ง Cookie (Auth & CSRF)
  headers: { 'Content-Type': 'application/json' },
});

// -------------------------------------------------------------------
// ✅ 1. Request Interceptor: จัดการ CSRF 
// -------------------------------------------------------------------
api.interceptors.request.use(
  (config) => {
    // อ่าน CSRF Token จาก Cookie
    const csrfToken = getCookie('csrf_');
    
    // ถ้ามี Token และเป็น Method ที่มีการแก้ไขข้อมูล ให้แนบ Header ไปด้วย
    if (csrfToken && ['post', 'put', 'delete', 'patch'].includes(config.method)) {
      config.headers['X-CSRF-Token'] = csrfToken;
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// -------------------------------------------------------------------
// ✅ 2. Response Interceptor: จัดการ Refresh Token 
// -------------------------------------------------------------------
api.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // เงื่อนไข: เจอ 401 AND ยังไม่เคยลอง Refresh มาก่อน
    if (
      error.response?.status === 401 && 
      !originalRequest._retry &&
      !originalRequest.url.includes('refresh-token')
    ) {
      originalRequest._retry = true; // แปะป้ายว่ากำลังกู้ชีพ

      try {
        // 1. แอบยิงไปขอต่ออายุ Token (Backend จะอ่าน Refresh Token ใน Cookie เอง)
        // ⚠️ เช็ค URL ตรงนี้ให้ชัวร์ว่า Backend คุณใช้ /v1/auth/refresh-token หรือ path ไหน
        await api.post('/auth/refresh-token');

        // 2. ถ้าผ่าน ให้ยิง Request เดิมซ้ำอีกรอบ
        return api(originalRequest);

      } catch (refreshError) {
        // 3. ถ้าไม่ผ่าน (เช่น Refresh Token หมดอายุด้วย) -> ดีดไปหน้า Login
        if (window.location.pathname !== '/login') {
            toast.error("Session หมดอายุ กรุณาเข้าสู่ระบบใหม่");
            window.location.href = '/login';
        }
        return Promise.reject(refreshError);
      }
    }

    // ถ้าไม่ใช่ 401 หรือแก้ไม่ได้ ก็ส่ง Error ต่อไปตามปกติ
    return Promise.reject(error);
  }
);

export default api;