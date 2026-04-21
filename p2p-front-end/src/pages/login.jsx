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
      const user = await login(credentials.username, credentials.password);

      console.log("Login Success. User Data:", user); // 🔍 Debug Log
      console.log("Roles:", user?.role);
      console.log("Dept Code:", user?.department_code);
      console.log("Dept Name:", user?.department);

      // 2. เช็ค Role เพื่อ Redirect
      const roles = user?.roles || [];

      const isAdmin = roles.some(r => r.toUpperCase() === 'ADMIN' || r.toUpperCase() === 'SUPER_ADMIN');
      const isOwner = roles.some(r => r.toUpperCase() === 'OWNER');
      const isDelegate = roles.some(r => r.toUpperCase() === 'DELEGATE' || r.toUpperCase() === 'BRANCH_DELEGATE');

      if (isAdmin) {
        console.log("Role Admin detected. Redirecting to Home...");
        navigate("/home");
      } else if (isOwner) {
        console.log("Role Owner detected. Redirecting to Owner Dashboard...");
        navigate("/owner/dashboard");
      } else if (isDelegate) {
        // Delegate usually goes to Dashboard too, or specific page? Assuming Home/Owner Dashboard for now.
        // User didn't specify Delegate Dashboard, assumig same as Owner or Admin view but limited.
        // Let's send Delegate to /home or /owner/dashboard depending on system design.
        // For now, let's assume Delegate goes to /home (Admin Dashboard view) but limited, OR owner dashboard.
        // Actually, usually Delegate works for Owner.
        // But "Delegate looks at User List".
        // Let's send to /owner/dashboard for now if they are delegate of owner.
        // OR if system is single dashboard -> /home.
        // Let's default to /home for now to be safe, as they need to access User Management.
        navigate("/home");
      } else {
        // Rule 1: Default User = No Access
        console.warn("Unauthorized User:", user?.id || "Unknown");
        toast.error("คุณไม่มีสิทธิ์เข้าใช้งานระบบ (No Access)");

        try {
          // Logout immediately to clear state, but keep them on login page
          await logout(false);
        } catch (logoutErr) {
          console.error("Logout error in Login:", logoutErr);
        }
        return; // Stop navigation
      }
      toast.success("เข้าสู่ระบบสำเร็จ!");

    } catch (error) {
      // 3. ถ้า Error (เช่น รหัสผิด)
      console.error("Login Failed:", error);

      // Check Keycloak/Backend 401 for Invalid Credentials
      if (error.response && error.response.status === 401) {
        toast.error("ชื่อผู้ใช้ หรือ รหัสผ่าน ไม่ถูกต้อง!");
      } else {
        // Other errors (Network, 500, or Client Logic)
        // Only show if it's NOT a "No Access" flow (which shouldn't end up here anyway)
        toast.error("เข้าสู่ระบบไม่สำเร็จ: " + (error.message || "Unknown Error"));
      }
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
                  Budget Service
                </Typography>
                <Typography variant="body2" >
                  ระบบบริการงบประมาณ
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