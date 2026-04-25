// src/hooks/useAuth.jsx
import React, { createContext, useContext, useState, useEffect } from 'react';
import api from '../utils/api/axiosInstance';

// 1. สร้าง Context (ห้องโถงกลาง)
const AuthContext = createContext(null);

// 2. สร้าง Provider (ตัวกระจายข้อมูล)
export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  // ฟังก์ชันเช็ค User (ใช้ตอนเปิดเว็บ หรือหลัง Login)
  const checkUser = async () => {
    // ... (unchanged)
    console.log("AuthProvider: checkUser called");
    try {
      const response = await api.get('/auth/profile');
      console.log("AuthProvider: User found", response.data);
      const userData = response.data.data;

      // ✅ Strict Role Check: Only set user state if they have a privileged role
      const roles = userData.roles || [];
      const hasAccess = roles.some(r => ['ADMIN', 'SUPER_ADMIN', 'OWNER', 'DELEGATE', 'BRANCH_DELEGATE'].includes(r.toUpperCase()));

      if (hasAccess) {
        setUser(userData);
      } else {
        console.warn("AuthProvider: User has no privileged roles. Redirecting to gateway login.");
        setUser(null);
        window.location.href = "/budget-dash/v1/login";
      }

      return userData; // Always return data so Login.jsx can check it
    } catch {
      console.log("AuthProvider: No user found");
      setUser(null);
      return null;
    } finally {
      setIsLoading(false);
    }
  };

  // เช็ค User ครั้งแรกตอนเปิดเว็บ
  useEffect(() => {
    console.log("AuthProvider: MOUNTED");
    checkUser();
    // return () => console.log("AuthProvider: UNMOUNTED");
  }, []);

  // ฟังก์ชัน Login
  // const login = async (username, password) => {
  //   await api.post('/auth/login', { username, password });
  //   return await checkUser();
  // };

  const login = () => {
    setIsLoading(true);
    // await api.post('/auth/login', { username, password });
    // await checkUser(); // 🔥 สำคัญ: โหลดข้อมูลใหม่ทันทีหลัง Login ผ่าน
    window.location.href = "/budget-dash/v1/login";
  };

  // ฟังก์ชัน Logout
  // const logout = async (redirect = true) => {
  //   try {
  //     await api.post('/auth/logout');
  //     setUser(null);
  //     if (redirect) {
  //       navigate("/login");
  //     }
  //   } catch (error) {
  //     console.error("Logout failed", error);
  //     // Ensure state is cleared even if network fails
  //     setUser(null);
  //     if (redirect) {
  //       navigate("/login");
  //     }
  //   }
  // };

  const logout = () => {
    setIsLoading(true);
    
    // 1. ล้างฝั่ง React
    setUser(null);
    localStorage.clear();
  
    // 2. ให้ Browser วิ่งไปหา Gateway เพื่อสั่ง Keycloak เคลียร์ Session 
    // (เดี๋ยวมันจะเด้งกลับมาหน้าเว็บเราเองตาม post_logout_redirect_uri)
    window.location.href = "/budget-dash/v1/logout"; 
  };

  return (
    <AuthContext.Provider value={{ user, login, logout, isLoading }}>
      {children}
    </AuthContext.Provider>
  );
};

// 3. สร้าง Hook สำหรับเรียกใช้ (หน้าอื่นเรียกใช้ตัวนี้เหมือนเดิม)
// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = () => {
  return useContext(AuthContext);
};