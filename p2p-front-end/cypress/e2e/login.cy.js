describe('Login Flow', () => {
    // ก่อนเริ่มทุก Test ให้ไปหน้า Login ก่อน
    beforeEach(() => {
      cy.visit('/login'); // ไม่ต้องใส่ localhost:5173 เพราะตั้งใน config แล้ว
    });
  
    it('Login สำเร็จ และเปลี่ยนไปหน้า Home', () => {
      // 1. เช็คว่าเจอช่องกรอกไหม
      cy.get('[data-testid="username-input"] input').should('be.visible');
  
      // 2. พิมพ์ User / Pass (ใส่ user จริงที่มีใน DB หรือ Mock เอา)
      cy.get('[data-testid="username-input"] input').type('tes11t'); 
      cy.get('[data-testid="password-input"] input').type('tes11t');
  
      // 3. กดปุ่ม Login
      cy.get('[data-testid="login-button"]').click();
  
      // 4. คาดหวังผลลัพธ์ (Expectation)
      // 4.1 URL ต้องเปลี่ยนไปหน้า /home
      cy.url().should('include', '/home');
      
      // 4.2 ต้องเจอข้อความต้อนรับ หรือ Toast Success
      cy.contains('เข้าสู่ระบบสำเร็จ').should('be.visible');
    });
  
    it('Login ผิด ต้องแจ้งเตือน Error', () => {
      cy.get('[data-testid="username-input"] input').type('wronguser');
      cy.get('[data-testid="password-input"] input').type('wrongpass');
      cy.get('[data-testid="login-button"]').click();
  
      // ต้องเจอข้อความ Error (จาก Toastify)
      cy.contains('ชื่อผู้ใช้ หรือ รหัสผ่าน ไม่ถูกต้อง').should('be.visible');
      
      // URL ต้องอยู่ที่เดิม
      cy.url().should('include', '/login');
    });
  });