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
    console.log("AuthProvider: checkUser called");
    try {
      // ⚠️ เช็ค URL ให้ตรงกับ Backend ของคุณ (เช่น /v1/auth/profile)
      const response = await api.get('/auth/profile');
      // Backend Returns { data: UserInfo, status: "OK", ... }
      // Axios returns { data: { data: UserInfo, ... } }
      console.log("AuthProvider: User found", response.data);
      setUser(response.data.data);
    } catch {
      console.log("AuthProvider: No user found");
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  };

  // เช็ค User ครั้งแรกตอนเปิดเว็บ
  useEffect(() => {
    console.log("AuthProvider: MOUNTED");
    checkUser();
    return () => console.log("AuthProvider: UNMOUNTED");
  }, []);

  // ฟังก์ชัน Login
  const login = async (username, password) => {
    await api.post('/auth/login', { username, password });
    await checkUser(); // 🔥 สำคัญ: โหลดข้อมูลใหม่ทันทีหลัง Login ผ่าน
  };

  // ฟังก์ชัน Logout
  const logout = async () => {
    try {
      await api.post('/auth/logout');
      setUser(null);
      window.location.href = "/login";
    } catch (error) {
      console.error("Logout failed", error);
    }
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