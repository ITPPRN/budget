import { lazy, useEffect } from "react";
import { Navigate } from "react-router-dom";
import Loadable from "../utils/loadable";
import Layout from "../layouts";

// --- Redirect unauthenticated users straight to APISIX OIDC login ---
function RedirectToGatewayLogin() {
  useEffect(() => {
    const returnTo = window.location.pathname + window.location.search;
    sessionStorage.setItem('oidc_return_to', returnTo);
    window.location.href = "/budget-dash/v1/login";
  }, []);
  return null;
}

// --- Import Pages (ใช้ Lazy Load) ---
const HomePage = Loadable(lazy(() => import("../pages/HomePage")));
const DetailPage = Loadable(lazy(() => import("../pages/Detail")));
const UserManagePage = Loadable(lazy(() => import("../pages/UserManage")));
const DataManagePage = Loadable(lazy(() => import("../pages/DataManage")));
const LogPage = Loadable(lazy(() => import("../pages/LogPage")));
const SyncMonitorPage = Loadable(lazy(() => import("../pages/SyncMonitor")));

// --- Owner Pages ---
const OwnerDashboard = Loadable(lazy(() => import("../pages/Owner/OwnerDashboard")));
const OwnerDetailReport = Loadable(lazy(() => import("../pages/Owner/OwnerDetailReport")));
const OwnerUserManage = Loadable(lazy(() => import("../pages/Owner/OwnerUserManage")));

// --- Define Routes ---
// รับค่า isLoggedIn และ user เข้ามา เพื่อตัดสินใจว่าจะพาไปไหน
const Routes = (isLoggedIn, user) => [
  {
    path: "/",
    // ถ้า Login แล้ว -> ใช้ Layout (มี Sidebar)
    // ถ้ายัง -> เด้งออกไปที่ OIDC login ของ APISIX gateway
    element: isLoggedIn ? <Layout /> : <RedirectToGatewayLogin />,
    children: [
      {
        path: "/",
        element: (() => {
          const roles = user?.roles?.map(r => r.toUpperCase()) || [];
          const isAdmin = roles.some(r => r.includes('ADMIN'));
          const isOwner = roles.some(r => r.includes('OWNER')) || !!user?.department_code || !!user?.department;

          if (isAdmin) return <Navigate to="/home" />;
          if (isOwner) return <Navigate to="/owner/dashboard" />;
          return <Navigate to="/home" />;
        })()
      },
      { path: "home", element: <HomePage /> },
      { path: "detail", element: <DetailPage /> },
      { path: "user", element: <UserManagePage /> },
      { path: "data", element: <DataManagePage /> },
      { path: "logs", element: <LogPage /> },
      { path: "sync-monitor", element: <SyncMonitorPage /> },

      // ... ใส่หน้าอื่นๆ เพิ่มตรงนี้ ...

      //Owner Routes
      { path: "owner/dashboard", element: <OwnerDashboard /> },
      { path: "owner/detail", element: <OwnerDetailReport /> },
      { path: "owner/user", element: <OwnerUserManage /> },
    ],
  },
  {
    path: "*",
    element: <h1>404 Not Found</h1>, // หรือใส่ Component <NotFound />
  },
];

export default Routes;