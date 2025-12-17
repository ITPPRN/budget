import { createTheme } from "@mui/material/styles";

// เปลี่ยนจาก const Theme = ... เป็น function ที่รับค่า mode
export const getTheme = (mode) => ({
  palette: {
    mode, // บอก MUI ว่าตอนนี้เป็น light หรือ dark
    ...(mode === "light"
      ? {
          // ☀️ โหมดสว่าง (ใช้สีเดิมของคุณเป๊ะๆ)
          primary: {
            main: "#043478",
            light: "#10254a",
            contrastText: "#ffffff", // ปกติสีเข้ม ตัวหนังสือควรขาว
          },
          text: {
            primary: "#343A40",
            secondary: "#6c757d",
          },
          background: {
            default: "#f4f6f8",
            paper: "#ffffff",
          },
          // Custom Colors ของคุณ (Light)
          primaryCustom: {
            main: "#23c6c8",
            dark: "#1FB2B4",
            contrastText: "#fff",
          },
          info: {
            main: "#1C84C6",
            dark: "#416393",
            contrastText: "#fff",
          },
          success: {
            main: "#6fbf73",
            dark: "#3e8e46",
            contrastText: "#fff",
          },
          warning: {
            main: "#f8AC59",
            dark: "#ffa000",
            contrastText: "#fff",
          },
          danger: {
            main: "#ED5565",
            dark: "#D54C5A",
            contrastText: "#fff",
          },
        }
      : {
          // 🌙 โหมดมืด (ปรับสีให้อ่านง่ายบนพื้นดำ)
          primary: {
            main: "#90caf9", // สีฟ้าอ่อนลงเพื่อให้เด่นบนพื้นดำ
            light: "#043478",
            contrastText: "#000000",
          },
          text: {
            primary: "#ffffff",
            secondary: "#aaaaaa",
          },
          background: {
            default: "#121212", // สีพื้นหลังดำมาตรฐาน MUI
            paper: "#1e1e1e",   // สีการ์ดเทาเข้ม
          },
          // Custom Colors (Dark) - ใช้สีเดิมแต่ปรับให้สว่างขึ้นนิดหน่อยได้ถ้าต้องการ
          primaryCustom: {
            main: "#23c6c8", // ใช้สีเดิมได้ หรือจะปรับให้อ่อนลงก็ได้
            dark: "#1FB2B4",
            contrastText: "#000",
          },
          info: {
            main: "#1C84C6",
            dark: "#416393",
            contrastText: "#fff",
          },
          success: {
            main: "#6fbf73",
            dark: "#3e8e46",
            contrastText: "#fff",
          },
          warning: {
            main: "#f8AC59",
            dark: "#ffa000",
            contrastText: "#fff",
          },
          danger: {
            main: "#ED5565",
            dark: "#D54C5A",
            contrastText: "#fff",
          },
        }),
  },
  typography: {
    fontFamily: "Kanit, sans-serif",
    allVariants: {
      // ⚠️ แก้ไขสำคัญ: ต้องเช็ค mode ไม่งั้นสีดำจะไปจมบนพื้นดำ
      color: mode === "light" ? "#343A40" : "#ffffff",
    },
  },
  // ถ้ามี Components Override (เช่น สไตล์ปุ่ม) ก็ใส่ต่อตรงนี้ได้เลย
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          textTransform: 'none', // ตัวอย่าง: ไม่ให้ปุ่มเป็นตัวใหญ่หมด
        },
      },
    },
  },
});