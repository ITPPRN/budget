import { lazy } from "react";
import { Navigate, Outlet } from "react-router-dom";
import Loadable from "../utils/loadable";
import Layout from "../layouts";

// --- Import Pages (ใช้ Lazy Load) ---
const Login = Loadable(lazy(() => import("../pages/login")));
const HomePage = Loadable(lazy(() => import("../pages/HomePage")));
const DetailPage = Loadable(lazy(() => import("../pages/Detail")));
const UserManagePage = Loadable(lazy(() => import("../pages/UserManage")));
const DataManagePage = Loadable(lazy(() => import("../pages/DataManage")));

// --- Owner Pages ---
const OwnerDashboard = Loadable(lazy(() => import("../pages/Owner/OwnerDashboard")));
const OwnerDetailReport = Loadable(lazy(() => import("../pages/Owner/OwnerDetailReport")));
const OwnerUserManage = Loadable(lazy(() => import("../pages/Owner/OwnerUserManage")));

// --- Define Routes ---
// รับค่า isLoggedIn และ user เข้ามา เพื่อตัดสินใจว่าจะพาไปไหน
const Routes = (isLoggedIn, user) => [
  {
    path: "/",
    // ถ้า Login แล้ว -> ให้ใช้ Layout (มี Sidebar)
    // ถ้ายัง -> ดีดไปหน้า Login
    element: isLoggedIn ? <Layout /> : <Navigate to="/login" />,
    children: [
      {
        path: "/",
        element: (user?.roles?.some(r => r.toUpperCase().includes('OWNER')) || !!user?.department_code || !!user?.department)
          ? <Navigate to="/owner/dashboard" />
          : <Navigate to="/home" />
      },
      { path: "home", element: <HomePage /> },
      { path: "detail", element: <DetailPage /> },
      { path: "user", element: <UserManagePage /> },
      { path: "data", element: <DataManagePage /> },

      // ... ใส่หน้าอื่นๆ เพิ่มตรงนี้ ...

      //Owner Routes
      { path: "owner/dashboard", element: <OwnerDashboard /> },
      { path: "owner/detail", element: <OwnerDetailReport /> },
      { path: "owner/user", element: <OwnerUserManage /> },
    ],
  },
  {
    path: "/",
    // ถ้ายังไม่ Login -> ให้แสดงหน้า Login
    // ถ้า Login แล้ว -> ดีดกลับไป Home (กันคนกด Back มาหน้า Login)
    element: !isLoggedIn ? <Outlet /> : <Navigate to="/" />,
    children: [
      { path: "login", element: <Login /> },
      // { path: "register", element: <Register /> },
    ],
  },
  {
    path: "*",
    element: <h1>404 Not Found</h1>, // หรือใส่ Component <NotFound />
  },
];

export default Routes;