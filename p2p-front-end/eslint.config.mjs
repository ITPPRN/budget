import js from "@eslint/js";
import globals from "globals";
import pluginReact from "eslint-plugin-react";
import pluginReactHooks from "eslint-plugin-react-hooks";
import pluginReactRefresh from "eslint-plugin-react-refresh";
import json from "@eslint/json";
import { defineConfig } from "eslint/config";

export default defineConfig([
  {
    // Ignores: รวมของ Yarn และไฟล์ Build (ใส่ไว้บนสุด)
    ignores: [
        'dist', 
        '.eslintrc.cjs', 
        '.pnp.*', 
        '.yarn/**'
    ], 
  },
  
  // -----------------------------------------------------
  // 1. Config สำหรับ JavaScript / React (แก้แล้ว ✅)
  // -----------------------------------------------------
  {
    // ✅ ระบุชัดเจนว่ากฎพวกนี้ใช้กับไฟล์นามสกุลเหล่านี้เท่านั้น (JSON จะไม่โดนหางเลข)
    files: ["**/*.{js,jsx,mjs,cjs}"],
    
    languageOptions: { 
        globals: globals.browser,
        parserOptions: {
            ecmaFeatures: { jsx: true }
        }
    },
    
    plugins: { 
        js, 
        react: pluginReact,
        'react-hooks': pluginReactHooks, 
        'react-refresh': pluginReactRefresh 
    },
    
    rules: {
      // ✅ ย้าย js.configs.recommended มาแตกใส่ตรงนี้แทน
      ...js.configs.recommended.rules, 
      
      ...pluginReact.configs.flat.recommended.rules,
      ...pluginReactHooks.configs.recommended.rules, 
      
      "react/react-in-jsx-scope": "off",
      "react/prop-types": "off",
      
      "react-refresh/only-export-components": [
        "warn",
        { allowConstantExport: true },
      ],

      // กฎอนุญาตตัวแปรที่ไม่ได้ใช้ ถ้าขึ้นต้นด้วยตัวใหญ่ หรือ _
      "no-unused-vars": [
        "warn", 
        { 
            "varsIgnorePattern": "^[A-Z_]", 
            "argsIgnorePattern": "^[A-Z_]" 
        }
      ]
    },
    settings: {
      react: {
        version: "detect",
      },
    },
  },
  
  // -----------------------------------------------------
  // 2. Config สำหรับ JSON (แยกออกมาต่างหาก)
  // -----------------------------------------------------
  {
    files: ["**/*.json"],
    plugins: { json },
    language: "json/json",
    extends: ["json/recommended"],
    // แถม: ปิดกฎ JS ที่อาจจะหลุดเข้ามา
    rules: {
      "no-irregular-whitespace": "off" 
    }
  },
  {
    files: ["**/*.jsonc"],
    plugins: { json },
    language: "json/jsonc",
    extends: ["json/recommended"],
  },
]);