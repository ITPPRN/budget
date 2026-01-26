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

// --- Define Routes ---
// รับค่า isLoggedIn เข้ามา เพื่อตัดสินใจว่าจะพาไปไหน
const Routes = (isLoggedIn) => [
  {
    path: "/",
    // ถ้า Login แล้ว -> ให้ใช้ Layout (มี Sidebar)
    // ถ้ายัง -> ดีดไปหน้า Login
    element: isLoggedIn ? <Layout /> : <Navigate to="/login" />,
    children: [
      { path: "/", element: <Navigate to="/home" /> },
      { path: "home", element: <HomePage /> },
      { path: "detail", element: <DetailPage /> },
      { path: "user", element: <UserManagePage /> },
      { path: "data", element: <DataManagePage /> },
      
      // ... ใส่หน้าอื่นๆ เพิ่มตรงนี้ ...
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