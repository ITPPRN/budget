const { defineConfig } = require("cypress");

module.exports = defineConfig({
  // ✅ 1. ตั้งค่า Reporter
  reporter: 'cypress-mochawesome-reporter',
  reporterOptions: {
    charts: true,             // มีกราฟวงกลมสรุปผล
    reportPageTitle: 'P2P Test Report', // ชื่อหัวข้อรายงาน
    embeddedScreenshots: true, // ฝังรูปตอน Error ลงในไฟล์เลย
    inlineAssets: true,        // รวมทุกอย่างในไฟล์ HTML เดียว (ส่งต่อง่าย)
    saveAllAttempts: false,
  },

  e2e: {
    baseUrl: 'http://localhost:3000',
    
    // ✅ 2. บังคับให้อัดวิดีโอ (ปกติ Cypress รุ่นใหม่จะปิดไว้ถ้าไม่สั่ง)
    video: true,
    
    // ตั้งค่าความละเอียดวิดีโอ (Optional)
    viewportWidth: 1280,
    viewportHeight: 720,

    setupNodeEvents(on, config) {
      // ✅ 3. เรียกใช้งาน Plugin Reporter ตรงนี้
      require('cypress-mochawesome-reporter/plugin')(on);
      return config;
    },
  },
});