import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  Box,
  Button,
  FormControl,
  IconButton,
  InputAdornment,
  InputLabel,
  OutlinedInput,
  Stack,
  TextField,
  Typography,
  CircularProgress
} from "@mui/material";
import VisibilityIcon from "@mui/icons-material/Visibility";
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff";
import { toast } from "react-toastify";

// ✅ เรียกใช้ Hook ที่ทำไว้
import { useAuth } from "../hooks/useAuth"; 
import { useConfig } from "../contexts/ConfigContext";

function Login() {

  const { mode } = useConfig();

  const [credentials, setCredentials] = useState({
    username: "",
    password: "",
  });
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false); // เพิ่มสถานะ Loading ของปุ่ม

  const { login } = useAuth(); // ดึงฟังก์ชัน login มาใช้
  const navigate = useNavigate();

  const handleChange = (prop) => (event) => {
    setCredentials({ ...credentials, [prop]: event.target.value });
  };

  const handleClickShowPassword = () => {
    setShowPassword(!showPassword);
  };

  const handleMouseDownPassword = (event) => {
    event.preventDefault();
  };

  // --- ฟังก์ชัน Login ---
  const onSubmitLogin = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      // 1. เรียกฟังก์ชัน login จาก useAuth
      // (ข้างในมันจะยิง API -> ได้ HttpOnly Cookie -> update user state)
      await login(credentials.username, credentials.password);
      
      // 2. ถ้าผ่าน มันจะไปต่อที่ routes/index.jsx ซึ่งจะดีดไปหน้า Home ให้เอง
      // แต่ใส่ navigate กันเหนียวไว้ก็ได้
      navigate("/home"); 
      toast.success("เข้าสู่ระบบสำเร็จ!");

    } catch (error) {
      // 3. ถ้า Error (เช่น รหัสผิด)
      console.error("Login Failed:", error);
      toast.error("ชื่อผู้ใช้ หรือ รหัสผ่าน ไม่ถูกต้อง!");
    } finally {
      setIsSubmitting(false); 
    }
  };

  
  const styles = {
    paperContainer: {
        background: "linear-gradient(135deg, #043478 0%, #10254a 100%)", 
      backgroundImage: `url(${"/image/home-ir.jpeg"})`, // ⚠️ เช็คว่ามีไฟล์รูปนี้จริงไหม
      backgroundPosition: "center",
      backgroundRepeat: "no-repeat",
      backgroundSize: "cover",
    },
  };

  return (
    <Box style={styles.paperContainer}>
      <Box
        sx={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          minHeight: "100vh",
        }}
      >
        <Box
          sx={{
            p: 4,
            width: "100%",
            maxWidth: 400, // จำกัดความกว้างไม่ให้ยืดเกิน
            display: "flex",
            flexDirection: "column",
            backgroundColor: "background.paper",
            borderRadius: "8px",
            boxShadow: "0px 10px 40px rgba(0,0,0,0.3)",
          }}
        >
          <form onSubmit={onSubmitLogin}>
            <Stack spacing={3} alignItems="center">
              
              {/* Logo Section */}
              <Box sx={{ mb: 1, textAlign: 'center' }}>
                {/* ⚠️ เช็คว่ามีไฟล์ Logo นี้จริงไหม */}
                <img
                  src={mode === 'dark' ? "/ac_white.png" : "/logo-acg.png"}
                  alt="logo"
                  style={{ height: "80px", objectFit: 'contain' }}
                  onError={(e) => { e.target.style.display = 'none' }} // ถ้าไม่มีรูปให้ซ่อน
                />
                <Typography variant="h5" sx={{ mt: 2, fontWeight: 'bold' }} >
                  P2P Service
                </Typography>
                <Typography variant="body2" >
                  ระบบจัดซื้อและการจัดการ
                </Typography>
              </Box>

              {/* Input Section */}
              <Stack spacing={2} width="100%">
                <TextField
                  fullWidth
                  size="small"
                  required
                  label="ชื่อผู้ใช้ (Username)"
                  value={credentials.username}
                  onChange={handleChange("username")}
                  variant="outlined"
                  autoFocus
                  data-testid="username-input"
                />

                <FormControl fullWidth required variant="outlined" size="small">
                  <InputLabel htmlFor="outlined-adornment-password">
                    รหัสผ่าน
                  </InputLabel>
                  <OutlinedInput
                    id="outlined-adornment-password"
                    data-testid="password-input"
                    type={showPassword ? "text" : "password"}
                    value={credentials.password}
                    onChange={handleChange("password")}
                    endAdornment={
                      <InputAdornment position="end">
                        <IconButton
                          onClick={handleClickShowPassword}
                          onMouseDown={handleMouseDownPassword}
                          edge="end"
                        >
                          {showPassword ? <VisibilityIcon /> : <VisibilityOffIcon />}
                        </IconButton>
                      </InputAdornment>
                    }
                    label="รหัสผ่าน"
                  />
                </FormControl>
              </Stack>

              {/* Button Section */}
              <Button
                variant="contained"
                size="large"
                color="primary"
                fullWidth
                type="submit"
                data-testid="login-button"
                disabled={isSubmitting} // ปิดปุ่มตอนกำลังโหลด
                sx={{ py: 1.2, fontWeight: 'bold' }}
              >
                {isSubmitting ? <CircularProgress size={24} color="inherit" /> : "เข้าสู่ระบบ"}
              </Button>
            </Stack>
          </form>
        </Box>
      </Box>
    </Box>
  );
}

export default Login;