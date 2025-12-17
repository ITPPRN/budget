// import axios from 'axios';
// import { toast } from 'react-toastify';

// const api = axios.create({
//   baseURL: '/v1', 
//   withCredentials: true, // สำคัญ! เพื่อส่ง Cookie
//   headers: { 'Content-Type': 'application/json' },
// });

// // --- เพิ่ม Interceptor เพื่อดักจับ Error ---
// api.interceptors.response.use(
//   (response) => {
//     return response;
//   },
//   (error) => {
//     // เช็คว่า Backend ตอบกลับมาว่า 401 (หมดสิทธิ์/Token หมดอายุ) หรือไม่
//     if (error.response && error.response.status === 401) {
      
//       // ถ้าไม่ใช่หน้า Login อยู่แล้ว ให้ดีดออก
//       if (window.location.pathname !== '/login') {
//         toast.error("Session หมดอายุ กรุณาเข้าสู่ระบบใหม่");
//         window.location.href = '/login';
//       }
//     }
//     return Promise.reject(error);
//   }
// );

// export default api;
import axios from 'axios';
import { toast } from 'react-toastify';

// 1. ฟังก์ชันช่วยแกะ Cookie (Helper Function)
// ทำหน้าที่ค้นหา Cookie ที่ชื่อตามที่ส่งเข้ามา
function getCookie(name) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(';').shift();
}

const api = axios.create({
  baseURL: '/v1', 
  withCredentials: true, // สำคัญ! เพื่อส่ง Cookie (Auth & CSRF)
  headers: { 'Content-Type': 'application/json' },
});

// --- [เพิ่มใหม่] Interceptor ฝั่ง Request ---
// ทำงานก่อนที่ Request จะถูกส่งออกไป
api.interceptors.request.use(
  (config) => {
    // พยายามอ่านค่า Cookie ที่ชื่อ "csrf_" (ต้องตรงกับ Backend: CookieName)
    const csrfToken = getCookie('csrf_');
    
    // ถ้าเจอ Cookie ให้แนบไปใน Header "X-CSRF-Token" (ต้องตรงกับ Backend: KeyLookup)
    // เฉพาะเมธอดที่เปลี่ยนแปลงข้อมูล (POST, PUT, DELETE, PATCH)
    if (csrfToken && ['post', 'put', 'delete', 'patch'].includes(config.method)) {
      config.headers['X-CSRF-Token'] = csrfToken;
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// --- Interceptor ฝั่ง Response (ของเดิมของคุณ) ---
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // เช็คว่า Backend ตอบกลับมาว่า 401 (หมดสิทธิ์/Token หมดอายุ) หรือไม่
    if (error.response && error.response.status === 401) {
      // ถ้าไม่ใช่หน้า Login อยู่แล้ว ให้ดีดออก
      if (window.location.pathname !== '/login') {
        toast.error("Session หมดอายุ กรุณาเข้าสู่ระบบใหม่");
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

export default api;