// import React, { useState } from 'react';
// import {
//     Paper,
//     Typography,
//     Box,
//     TableContainer,
//     Table,
//     TableHead,
//     TableRow,
//     TableCell,
//     TableBody,
//     Skeleton,
//     TablePagination,
//     IconButton,
//     Dialog,
//     DialogContent,
//     Popover,
//     TextField,
//     Button,
//     Tooltip
// } from '@mui/material';
// import FullscreenIcon from '@mui/icons-material/Fullscreen';
// import CloseIcon from '@mui/icons-material/Close';
// import FilterListIcon from '@mui/icons-material/FilterList';
// import DownloadIcon from '@mui/icons-material/Download';
// import CheckCircleIcon from '@mui/icons-material/CheckCircle';
// import AnnouncementIcon from '@mui/icons-material/Announcement';
// import { useAuth } from '../../hooks/useAuth';
// import api from '../../utils/api/axiosInstance';
// import { toast } from 'react-toastify';
// import AuditReportModal from './AuditReportModal';

// const ActualTable = React.memo(({
//     loading,
//     data = [],
//     dateFilter,
//     onDateFilterChange,
//     page,
//     rowsPerPage,
//     totalCount,
//     onPageChange,
//     onRowsPerPageChange,
//     onDownload,
//     filters = {} // Added filters for Audit API
// }) => {
//     const { user } = useAuth();
//     const isOwner = user?.roles?.some(r => r.toUpperCase() === 'OWNER');

//     const [openFullScreen, setOpenFullScreen] = useState(false);
//     const [auditLoading, setAuditLoading] = useState(false);
//     const [reportModalOpen, setReportModalOpen] = useState(false);
//     const [rejectedBasket, setRejectedBasket] = useState([]); // 🛡️ Local Basket for Rejections

//     // Filter Popover State
//     const [anchorEl, setAnchorEl] = useState(null);
//     const [tempDateFilter, setTempDateFilter] = useState({ startDate: '', endDate: '' });

//     const handleFilterClick = (event) => {
//         setTempDateFilter(dateFilter || { startDate: '', endDate: '' });
//         setAnchorEl(event.currentTarget);
//     };

//     const handleFilterClose = () => {
//         setAnchorEl(null);
//     };

//     const handleApplyFilter = () => {
//         onDateFilterChange(tempDateFilter);
//         setAnchorEl(null);
//     };

//     const handleClearFilter = () => {
//         setTempDateFilter({ startDate: '', endDate: '' });
//         onDateFilterChange({ startDate: '', endDate: '' });
//         setAnchorEl(null);
//     };

//     const handleChangePage = (event, newPage) => {
//         onPageChange(newPage);
//     };

//     const handleChangeRowsPerPage = (event) => {
//         onRowsPerPageChange(parseInt(event.target.value, 10));
//     };

//     // const handleApprove = async () => {
//     //     const basketCount = rejectedBasket.length;
//     //     const msg = basketCount > 0
//     //         ? `คุณมี ${basketCount} รายการในตะกร้าที่จะถูกปฏิเสธ และรายการที่เหลือทั้งหมดในเดือนนี้จะถูกยืนยันอัตโนมัติ. คุณแน่ใจหรือไม่?`
//     //         : "คุณแน่ใจหรือไม่ว่าต้องการยืนยันความถูกต้องของข้อมูลทั้งหมด (Approve All) สำหรับเดือนที่เลือก?";

//     //     if (!window.confirm(msg)) return;

//     //     setAuditLoading(true);
//     //     try {
//     //         await api.post('/budgets/audit/approve', {
//     //             entity: filters.entity,
//     //             branch: filters.branch,
//     //             department: filters.department,
//     //             year: filters.year,
//     //             month: filters.month || (filters.months && filters.months[0]),
//     //             rejected_item_ids: rejectedBasket // Send the basket!
//     //         });
//     //         toast.success("ยืนยันข้อมูลและส่งรายการปฏิเสธเรียบร้อยแล้ว");
//     //         setRejectedBasket([]); // Clear basket after success
//     //     } catch (err) {
//     //         toast.error("เกิดข้อผิดพลาดในการยืนยันข้อมูล: " + (err.response?.data?.error || err.message));
//     //     } finally {
//     //         setAuditLoading(false);
//     //     }
//     // };

//     const handleApprove = async () => {
//         const basketCount = rejectedBasket.length;
//         const msg = basketCount > 0
//             ? `คุณมี ${basketCount} รายการในตะกร้าที่จะถูกปฏิเสธ และรายการที่เหลือทั้งหมดในเดือนนี้จะถูกยืนยันอัตโนมัติ. คุณแน่ใจหรือไม่?`
//             : "คุณแน่ใจหรือไม่ว่าต้องการยืนยันความถูกต้องของข้อมูลทั้งหมด (Approve All) สำหรับเดือนที่เลือก?";

//         if (!window.confirm(msg)) return;

//         setAuditLoading(true);
//         try {
//             await api.post('/budgets/audit/approve', {
//                 entity: filters.entity,
//                 branch: filters.branch,
//                 department: filters.department,
//                 year: filters.year,
//                 month: filters.month || (filters.months && filters.months[0]),
//                 // 👇 แก้ตรงนี้: สกัดเอาเฉพาะ ID ของตะกร้าส่งไปให้ API
//                 rejected_item_ids: rejectedBasket.map(item => item.id)
//             });
//             toast.success("ยืนยันข้อมูลและส่งรายการปฏิเสธเรียบร้อยแล้ว");
//             setRejectedBasket([]); // ล้างตะกร้าหน้าจอ
//         } catch (err) {
//             toast.error("เกิดข้อผิดพลาดในการยืนยันข้อมูล: " + (err.response?.data?.error || err.message));
//         } finally {
//             setAuditLoading(false);
//         }
//     };

//     // const handleReportSubmit = (selectedIds) => {
//     //     // 🛠️ Just add to basket, don't call API yet!
//     //     setRejectedBasket(selectedIds);
//     //     console.log(selectedIds)
//     //     setReportModalOpen(false);
//     //     toast.info(`เก็บ ${selectedIds.length} รายการไว้ในตะกร้าปฏิเสธแล้ว`);
//     // };

//     const handleReportSubmit = async (selectedItems) => {
//         // selectedItems ที่ส่งมาจาก Modal จะเป็น Array ของ Object
//         setAuditLoading(true);
//         try {
//             // 1. สกัดเอาเฉพาะ ID เป็น Array คลีนๆ เพื่อส่งให้ Go
//             // หน้าตาจะเป็น ["uuid-1", "uuid-2", ...]
//             const payload = selectedItems.map(item => item.id);

//             // 2. ยิง API ไปหา Controller addBasket (ปรับ URL ให้ตรงกับของคุณ)
//             await api.post('/budgets/audit/basket/add', payload);

//             // 3. ถ้า API สำเร็จ ค่อยอัปเดตตะกร้าหน้าจอ (เพื่อให้ UI เปลี่ยน)
//             setRejectedBasket(selectedItems);
//             setReportModalOpen(false);
//             toast.success(`เพิ่ม ${selectedItems.length} รายการลงตะกร้าในฐานข้อมูลเรียบร้อย`);

//         } catch (err) {
//             toast.error("เกิดข้อผิดพลาดในการบันทึกลงตะกร้า: " + (err.response?.data?.error || err.message));
//         } finally {
//             setAuditLoading(false);
//         }
//     };

//     // Data is now paginated from server
//     const paginatedData = data;

//     const renderTable = (isMaximized = false) => (
//         <React.Fragment>
//             <TableContainer sx={{ flexGrow: 1, width: '100%', overflow: 'auto', border: '1px solid #e0e0e0', borderRadius: '4px' }}>
//                 <Table stickyHeader size={isMaximized ? "medium" : "small"} sx={{ minWidth: '100%' }}>
//                     <TableHead>
//                         <TableRow>
//                             <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>GL Code</TableCell>
//                             <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Account Name</TableCell>
//                             <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Doc No.</TableCell>
//                             <TableCell align="right" sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Amount</TableCell>
//                             <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Vendor</TableCell>
//                             <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Description</TableCell>
//                             <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Date</TableCell>
//                             <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold' }}>Status</TableCell>
//                         </TableRow>
//                     </TableHead>
//                     <TableBody>
//                         {loading ? (
//                             Array.from(new Array(10)).map((_, index) => (
//                                 <TableRow key={index}>
//                                     <TableCell><Skeleton variant="text" /></TableCell>
//                                     <TableCell><Skeleton variant="text" /></TableCell>
//                                     <TableCell><Skeleton variant="text" /></TableCell>
//                                     <TableCell><Skeleton variant="text" /></TableCell>
//                                     <TableCell><Skeleton variant="text" /></TableCell>
//                                     <TableCell><Skeleton variant="text" /></TableCell>
//                                     <TableCell><Skeleton variant="text" /></TableCell>
//                                     <TableCell><Skeleton variant="text" /></TableCell>
//                                 </TableRow>
//                             ))
//                         ) : paginatedData.length > 0 ? (
//                             paginatedData.map((row, index) => (
//                                 <TableRow key={index} hover>
//                                     <TableCell sx={{ borderRight: '1px solid #e0e0e0', fontSize: '0.85rem' }}>{row.conso_gl}</TableCell>
//                                     <TableCell sx={{ borderRight: '1px solid #e0e0e0', fontSize: '0.85rem' }}>{row.gl_account_name}</TableCell>
//                                     <TableCell sx={{ borderRight: '1px solid #e0e0e0' }}>{row.document_no}</TableCell>
//                                     <TableCell align="right" sx={{ fontWeight: 'bold', color: parseFloat(row.amount) < 0 ? 'red' : 'green', borderRight: '1px solid #e0e0e0' }}>
//                                         {parseFloat(row.amount || 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
//                                     </TableCell>
//                                     <TableCell sx={{ borderRight: '1px solid #e0e0e0' }}>{row.vendor || "-"}</TableCell>
//                                     <TableCell sx={{ borderRight: '1px solid #e0e0e0' }}>{row.description}</TableCell>
//                                     <TableCell sx={{ borderRight: '1px solid #e0e0e0', fontSize: '0.85rem' }}>{row.posting_date}</TableCell>
//                                     <TableCell sx={{ fontWeight: 'bold', fontSize: '0.75rem', color:
//                                         row.status === 'REPORTED' ? '#1cc88a' :
//                                         row.status === 'DRAFT' ? '#f6c23e' :
//                                         row.status === 'COMPLETE' ? '#4e73df' :
//                                         'inherit'
//                                     }}>
//                                         {row.status || row.audit_status || "-"}
//                                     </TableCell>
//                                 </TableRow>
//                             ))
//                         ) : (
//                             <TableRow>
//                                 <TableCell colSpan={8} align="center" sx={{ py: 5, color: 'text.secondary' }}>
//                                     No actual data found for selected filters
//                                 </TableCell>
//                             </TableRow>
//                         )}
//                     </TableBody>
//                 </Table>
//             </TableContainer>

//             <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderTop: '1px solid #e0e0e0' }}>
//                 {/* Audit Action Buttons (Bottom Left) */}
//                 <Box sx={{ pl: 1, display: 'flex', gap: 1 }}>
//                     {isOwner && data.length > 0 && (
//                         <>
//                             <Button
//                                 variant="contained"
//                                 color="success"
//                                 size="small"
//                                 startIcon={<CheckCircleIcon />}
//                                 onClick={handleApprove}
//                                 disabled={auditLoading}
//                                 sx={{ textTransform: 'none', borderRadius: '4px', fontSize: '0.75rem' }}
//                             >
//                                 ยืนยันรายการ
//                             </Button>
//                             <Button
//                                 variant="contained"
//                                 color="error"
//                                 size="small"
//                                 startIcon={<AnnouncementIcon />}
//                                 onClick={() => setReportModalOpen(true)}
//                                 disabled={auditLoading}
//                                 sx={{ textTransform: 'none', borderRadius: '4px', fontSize: '0.75rem' }}
//                             >
//                                 {rejectedBasket.length > 0 ? `ปฏิเสธรายการ (${rejectedBasket.length})` : "ปฏิเสธรายการ"}
//                             </Button>
//                         </>
//                     )}
//                 </Box>

//                 {/* Pagination Control (Right) */}
//                 {!loading && data.length > 0 && (
//                     <TablePagination
//                         rowsPerPageOptions={[10, 25, 50, 100]}
//                         component="div"
//                         count={totalCount}
//                         rowsPerPage={rowsPerPage}
//                         page={page}
//                         onPageChange={handleChangePage}
//                         onRowsPerPageChange={handleChangeRowsPerPage}
//                         sx={{
//                             '.MuiTablePagination-toolbar': { minHeight: '36px', px: 1 },
//                             '.MuiTablePagination-selectLabel, .MuiTablePagination-input, .MuiTablePagination-displayedRows': { fontSize: '0.75rem' }
//                         }}
//                     />
//                 )}
//             </Box>
//         </React.Fragment>
//     );

//     const openFilter = Boolean(anchorEl);

//     return (
//         <Paper sx={{
//             p: 1.5,
//             pt: 1,
//             display: 'flex',
//             flexDirection: 'column',
//             borderRadius: 2,
//             mb: 2,
//             flex: '1 1 0', // Equal height split
//             minHeight: { xs: '400px', md: 0 },
//             overflow: 'hidden'
//         }}>
//             <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2, px: 1 }}>
//                 <Box sx={{ bgcolor: '#043478', px: 2, py: 0.5, borderRadius: '8px' }}>
//                     <Typography variant="h6" sx={{ color: 'white', fontWeight: 'bold' }}>
//                         Actual Detail
//                     </Typography>
//                 </Box>
//                 <Box>
//                     <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
//                     <IconButton onClick={handleFilterClick} size="small" sx={{ color: '#424242'}}>
//                         <FilterListIcon />
//                     </IconButton>
//                     <IconButton onClick={() => setOpenFullScreen(true)} size="small" sx={{ color: '#424242' }}>
//                         <FullscreenIcon />
//                     </IconButton>
//                     <IconButton size="small" onClick={onDownload} sx={{ color: '#424242' }} disabled={!onDownload || loading || data.length === 0}>
//                         <DownloadIcon sx={{ fontSize: 18 }} />
//                     </IconButton>
//                 </Box>
//                 </Box>
//             </Box>

//             {/* Filter Popover */}
//             <Popover
//                 open={openFilter}
//                 anchorEl={anchorEl}
//                 onClose={handleFilterClose}
//                 anchorOrigin={{
//                     vertical: 'bottom',
//                     horizontal: 'right',
//                 }}
//                 transformOrigin={{
//                     vertical: 'top',
//                     horizontal: 'right',
//                 }}
//             >
//                 <Box sx={{ p: 2, display: 'flex', flexDirection: 'column', gap: 2, minWidth: 300 }}>
//                     <Typography variant="subtitle2" fontWeight="bold">Filter Date</Typography>
//                     <TextField
//                         label="Start Date"
//                         type="date"
//                         size="small"
//                         InputLabelProps={{ shrink: true }}
//                         value={tempDateFilter.startDate}
//                         onChange={(e) => setTempDateFilter({ ...tempDateFilter, startDate: e.target.value })}
//                         fullWidth
//                     />
//                     <TextField
//                         label="End Date"
//                         type="date"
//                         size="small"
//                         InputLabelProps={{ shrink: true }}
//                         value={tempDateFilter.endDate}
//                         onChange={(e) => setTempDateFilter({ ...tempDateFilter, endDate: e.target.value })}
//                         fullWidth
//                     />
//                     <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 1 }}>
//                         <Button onClick={handleClearFilter} size="small" color="inherit">
//                             Clear
//                         </Button>
//                         <Button onClick={handleApplyFilter} size="small" variant="contained" color="primary">
//                             Apply
//                         </Button>
//                     </Box>
//                 </Box>
//             </Popover>

//             {renderTable(false)}

//             {/* Fullscreen Dialog */}
//             <Dialog
//                 open={openFullScreen}
//                 onClose={() => setOpenFullScreen(false)}
//                 fullScreen
//             >
//                 <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', bgcolor: '#043478', p: 2 }}>
//                     <Typography variant="h6" sx={{ color: 'white', fontWeight: 'bold' }}>
//                         Actual Detail (Full Screen)
//                     </Typography>
//                     <Box>
//                         {/* We can show filter here too if needed, but per request it's next to expand button on main view?
//                              Let's assume users might want to filter inside full screen too. I'll clone the button here. */}
//                         <IconButton onClick={handleFilterClick} sx={{ color: 'white', mr: 2 }}>
//                             <FilterListIcon />
//                         </IconButton>
//                         <IconButton onClick={() => setOpenFullScreen(false)} sx={{ color: 'white' }}>
//                             <CloseIcon />
//                         </IconButton>
//                     </Box>
//                 </Box>
//                 <DialogContent sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
//                     {renderTable(true)}
//                 </DialogContent>
//             </Dialog>

//             <AuditReportModal
//                 open={reportModalOpen}
//                 onClose={() => setReportModalOpen(false)}
//                 filters={filters}
//                 // initialItems={data.filter(d => rejectedBasket.includes(d.id))} // Pass actual objects
//                 initialItems={rejectedBasket}
//                 onSubmit={handleReportSubmit}
//                 loading={auditLoading}
//             />
//         </Paper>
//     );
// });

// export default ActualTable;
import React, { useState, useEffect } from "react";
import {
  Paper,
  Typography,
  Box,
  TableContainer,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  Skeleton,
  TablePagination,
  IconButton,
  Dialog,
  DialogContent,
  DialogTitle,
  DialogActions,
  Popover,
  TextField,
  Button,
  Tooltip,
  Badge,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from "@mui/material";
import FullscreenIcon from "@mui/icons-material/Fullscreen";
import CloseIcon from "@mui/icons-material/Close";
import FilterListIcon from "@mui/icons-material/FilterList";
import DownloadIcon from "@mui/icons-material/Download";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import AnnouncementIcon from "@mui/icons-material/Announcement";
import ShoppingBasketIcon from "@mui/icons-material/ShoppingBasket";
import { useAuth } from "../../hooks/useAuth";
import api from "../../utils/api/axiosInstance";
import { toast } from "react-toastify";
import AuditReportModal from "./AuditReportModal";
import BasketModal from "./BasketModal";
import ApproveConfirmModal from "./ApproveConfirmModal";

const ActualTable = React.memo(
  ({
    loading,
    data = [],
    dateFilter,
    onDateFilterChange,
    page,
    rowsPerPage,
    totalCount,
    onPageChange,
    onRowsPerPageChange,
    onDownload,
    filters = {},
  }) => {
    const { user } = useAuth();
    const roles = (user?.roles || []).map((r) => r.toUpperCase());
    const isOwner = roles.includes("OWNER");
    const isDelegate =
      roles.includes("DELEGATE") || roles.includes("BRANCH_DELEGATE");
    const canAudit = isOwner || isDelegate; // OWNER + delegates: เพิ่มเข้าตะกร้าได้

    const [openFullScreen, setOpenFullScreen] = useState(false);
    const [auditLoading, setAuditLoading] = useState(false);
    const [reportModalOpen, setReportModalOpen] = useState(false);
    const [basketModalOpen, setBasketModalOpen] = useState(false);
    const [rejectedBasket, setRejectedBasket] = useState([]); // 🛡️ State หลักเก็บข้อมูลตะกร้า

    const [approveDialogOpen, setApproveDialogOpen] = useState(false);
    const [auditComplete, setAuditComplete] = useState(false);
    const [auditStats, setAuditStats] = useState({
      total: 0,
      reviewed: 0,
      remaining: 0,
    });

    // 🌟 1. ฟังก์ชันดึงข้อมูลตะกร้า (แยกออกมาเพื่อให้เรียกใช้ซ้ำได้)
    const fetchBasket = async () => {
      try {
        const res = await api.get("/budgets/audit/basket/list");
        setRejectedBasket(res.data || []); // ได้ข้อมูลมาปุ๊บ อัปเดต State ปั๊บ (Badge จะเปลี่ยนเลขทันที)
      } catch (err) {
        console.error("Failed to load basket", err);
      }
    };

    // 🌟 2. ดึงข้อมูล "ทันที" ที่โหลดหน้าเว็บเสร็จ
    useEffect(() => {
      if (canAudit) {
        fetchBasket();
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isOwner]);

    // Filter Popover State
    const [anchorEl, setAnchorEl] = useState(null);
    const [tempDateFilter, setTempDateFilter] = useState({
      startDate: "",
      endDate: "",
    });

    const getApproveTarget = () => {
      // 1. User กด filter ตรงหัวตาราง Actual แล้ว Apply → ใช้ตามนั้น
      if (dateFilter?.startDate) {
        const dateParts = dateFilter.startDate.split("-"); // "2026-04-15" -> ["2026", "04", "15"]
        return {
          year: dateParts[0],
          month: dateParts[1],
        };
      }

      // 2. ยังไม่ได้เลือก → ใช้เดือนปัจจุบัน
      const now = new Date();
      return {
        year: String(now.getFullYear()),
        month: String(now.getMonth() + 1).padStart(2, "0"),
      };
    };

    // 🌟 เช็คว่า audit log ครบทุก department ของ user แล้วหรือยัง
    const checkAuditComplete = async (year, month) => {
      try {
        const res = await api.get(`/budgets/audit/check-complete?year=${year}&month=${month}`);
        const d = res.data || {};
        setAuditComplete(d.is_complete || false);
        setAuditStats({
          total: d.total_count || 0,
          reviewed: d.reviewed_count || 0,
          remaining: d.pending_count || 0,
        });
      } catch (err) {
        console.error("Failed to check audit complete", err);
        setAuditComplete(false);
        setAuditStats({ total: 0, reviewed: 0, remaining: 0 });
      }
    };

    useEffect(() => {
      if (canAudit) {
        const target = getApproveTarget();
        checkAuditComplete(target.year, target.month);
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [canAudit, dateFilter?.startDate]);

    const handleFilterClick = (event) => {
      setTempDateFilter(dateFilter || { startDate: "", endDate: "" });
      setAnchorEl(event.currentTarget);
    };

    const handleFilterClose = () => {
      setAnchorEl(null);
    };

    const handleApplyFilter = () => {
      onDateFilterChange(tempDateFilter);
      setAnchorEl(null);
    };

    const handleClearFilter = () => {
      setTempDateFilter({ startDate: "", endDate: "" });
      onDateFilterChange({ startDate: "", endDate: "" });
      setAnchorEl(null);
    };

    const handleChangePage = (event, newPage) => {
      onPageChange(newPage);
    };
    const handleChangeRowsPerPage = (event) => {
      onRowsPerPageChange(parseInt(event.target.value, 10));
    };

    // const handleApprove = async () => {
    //     const basketCount = rejectedBasket.length;
    //     const msg = basketCount > 0
    //         ? `คุณมี ${basketCount} รายการในตะกร้าที่จะถูกปฏิเสธ และรายการที่เหลือทั้งหมดในเดือนนี้จะถูกยืนยันอัตโนมัติ. คุณแน่ใจหรือไม่?`
    //         : "คุณแน่ใจหรือไม่ว่าต้องการยืนยันความถูกต้องของข้อมูลทั้งหมด (Approve All) สำหรับเดือนที่เลือก?";

    //     if (!window.confirm(msg)) return;

    //     setAuditLoading(true);
    //     try {
    //         await api.post('/budgets/audit/approve', {
    //             entity: filters.entity,
    //             branch: filters.branch,
    //             department: filters.department,
    //             year: filters.year,
    //             month: filters.month || (filters.months && filters.months[0]),
    //             // rejected_item_ids: rejectedBasket.map(item => item.id)
    //         });
    //         toast.success("ยืนยันข้อมูลและส่งรายการปฏิเสธเรียบร้อยแล้ว");

    //         // ล้างตะกร้าหลังกดยืนยันสำเร็จ
    //         setRejectedBasket([]);
    //         setBasketModalOpen(false);
    //     } catch (err) {
    //         toast.error("เกิดข้อผิดพลาดในการยืนยันข้อมูล: " + (err.response?.data?.error || err.message));
    //     } finally {
    //         setAuditLoading(false);
    //     }
    // };

    // 🌟 1. ฟังก์ชันสำหรับ "เรียกเปิดหน้าต่าง Modal"
    const handleOpenApproveDialog = () => {
      setBasketModalOpen(false); // ปิดหน้าต่างตะกร้า (ถ้าเปิดอยู่) เพื่อไม่ให้ Modal ซ้อนกัน
      setApproveDialogOpen(true); // เปิดหน้าต่างยืนยันอันใหม่
    };

    // 🌟 2. ฟังก์ชันสำหรับ "ยิง API จริง" เมื่อกดปุ่มใน Modal
    const executeApprove = async () => {
      const target = getApproveTarget(); // 👈 แกะเอาจากตัวเลือกในตารางนี้เองเลย

      setApproveDialogOpen(false);
      setAuditLoading(true);
      try {
        // console.log(filters.entity);
        // console.log(filters.branch);
        // console.log(filters.department);
        // console.log(target.year);
        // console.log(target.month);
        await api.post("/budgets/audit/approve", {
          entity: filters.entity,
          branch: filters.branch,
          department: filters.department,
          year: target.year, // 🌟 ใช้ปีที่แกะออกมา
          month: target.month, // 🌟 ใช้เดือนที่แกะออกมา
        });
        toast.success(
          `ยืนยันข้อมูลเดือน ${target.month}/${target.year} เรียบร้อยแล้ว`
        );
        setRejectedBasket([]);
        // เช็คสถานะ audit อีกครั้งหลัง approve สำเร็จ เพื่อ disable ปุ่มทันที
        checkAuditComplete(target.year, target.month);
      } catch (err) {
        toast.error(
          "ยืนยันไม่สำเร็จ: " + (err.response?.data?.error || err.message)
        );
      } finally {
        setAuditLoading(false);
      }
    };

    // 🌟 3. เมื่อกดเพิ่มของลงตะกร้าจาก AuditReportModal
    const handleReportSubmit = async (selectedItems) => {
      setAuditLoading(true);
      try {
        const payload = selectedItems.map((item) => ({
          transaction_id: item.id,
          note: item.note || "",
        }));
        await api.post("/budgets/audit/basket/add", payload); // ยิงหลังบ้านเซฟลง DB

        // 🌟 หัวใจความเรียลไทม์: เซฟเสร็จ สั่งดึงข้อมูลใหม่ทันที
        await fetchBasket();

        setReportModalOpen(false);
        toast.success(`เพิ่ม ${selectedItems.length} รายการลงตะกร้าเรียบร้อย`);
      } catch (err) {
        toast.error(
          "เกิดข้อผิดพลาดในการบันทึกลงตะกร้า: " +
            (err.response?.data?.error || err.message)
        );
      } finally {
        setAuditLoading(false);
      }
    };

    // 🌟 4. เมื่อกดถังขยะลบรายการจาก BasketModal
    const handleRemoveFromBasket = async (id) => {
      try {
        await api.delete(`/budgets/audit/basket/${id}`); // ยิงหลังบ้านลบจาก DB

        // 🌟 ลบออกจากหน้าจอทันที ไม่ต้องรอรีเฟรช (ตัวเลข Badge จะลดลงทันที)
        setRejectedBasket((prev) => prev.filter((item) => item.id !== id));
        toast.success("ลบรายการออกจากตะกร้าแล้ว");
      } catch (err) {
        toast.error(
          "ลบข้อมูลไม่สำเร็จ: " + (err.response?.data?.error || err.message)
        );
      }
    };

    // 🌟 5. แก้ note ของรายการในตะกร้า — optimistic update + sync DB
    const handleUpdateBasketNote = async (id, note) => {
      setRejectedBasket((prev) =>
        prev.map((item) => (item.id === id ? { ...item, note } : item))
      );
      try {
        await api.patch(`/budgets/audit/basket/${id}`, { note });
      } catch (err) {
        toast.error(
          "บันทึก Note ไม่สำเร็จ: " + (err.response?.data?.error || err.message)
        );
        // โหลดใหม่เพื่อย้อน state กลับหากบันทึกพลาด
        fetchBasket();
      }
    };

    const paginatedData = data;

    const renderTable = (isMaximized = false) => (
      <React.Fragment>
        <TableContainer
          sx={{
            flexGrow: 1,
            width: "100%",
            overflow: "auto",
            border: "1px solid #e0e0e0",
            borderRadius: "4px",
          }}
        >
          <Table
            stickyHeader
            size={isMaximized ? "medium" : "small"}
            sx={{ minWidth: "100%" }}
          >
            {/* ... ส่วนหัวตารางและข้อมูลตารางเหมือนเดิมไม่เปลี่ยนแปลง ... */}
            <TableHead>
              <TableRow>
                <TableCell
                  sx={{
                    bgcolor: "#043478",
                    color: "white",
                    fontWeight: "bold",
                    borderRight: "1px solid rgba(255,255,255,0.3)",
                  }}
                >
                  GL Code
                </TableCell>
                <TableCell
                  sx={{
                    bgcolor: "#043478",
                    color: "white",
                    fontWeight: "bold",
                    borderRight: "1px solid rgba(255,255,255,0.3)",
                  }}
                >
                  Account Name
                </TableCell>
                <TableCell
                  sx={{
                    bgcolor: "#043478",
                    color: "white",
                    fontWeight: "bold",
                    borderRight: "1px solid rgba(255,255,255,0.3)",
                  }}
                >
                  Doc No.
                </TableCell>
                <TableCell
                  align="right"
                  sx={{
                    bgcolor: "#043478",
                    color: "white",
                    fontWeight: "bold",
                    borderRight: "1px solid rgba(255,255,255,0.3)",
                  }}
                >
                  Amount
                </TableCell>
                <TableCell
                  sx={{
                    bgcolor: "#043478",
                    color: "white",
                    fontWeight: "bold",
                    borderRight: "1px solid rgba(255,255,255,0.3)",
                  }}
                >
                  Vendor
                </TableCell>
                <TableCell
                  sx={{
                    bgcolor: "#043478",
                    color: "white",
                    fontWeight: "bold",
                    borderRight: "1px solid rgba(255,255,255,0.3)",
                  }}
                >
                  Description
                </TableCell>
                <TableCell
                  sx={{
                    bgcolor: "#043478",
                    color: "white",
                    fontWeight: "bold",
                    borderRight: "1px solid rgba(255,255,255,0.3)",
                  }}
                >
                  Date
                </TableCell>
                <TableCell
                  sx={{
                    bgcolor: "#043478",
                    color: "white",
                    fontWeight: "bold",
                  }}
                >
                  Status
                </TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {loading ? (
                Array.from(new Array(10)).map((_, index) => (
                  <TableRow key={index}>
                    <TableCell>
                      <Skeleton variant="text" />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" />
                    </TableCell>
                    <TableCell>
                      <Skeleton variant="text" />
                    </TableCell>
                  </TableRow>
                ))
              ) : paginatedData.length > 0 ? (
                paginatedData.map((row, index) => (
                  <TableRow key={index} hover>
                    <TableCell
                      sx={{
                        borderRight: "1px solid #e0e0e0",
                        fontSize: "0.85rem",
                      }}
                    >
                      {row.conso_gl}
                    </TableCell>
                    <TableCell
                      sx={{
                        borderRight: "1px solid #e0e0e0",
                        fontSize: "0.85rem",
                      }}
                    >
                      {row.gl_account_name}
                    </TableCell>
                    <TableCell sx={{ borderRight: "1px solid #e0e0e0" }}>
                      {row.document_no}
                    </TableCell>
                    <TableCell
                      align="right"
                      sx={{
                        fontWeight: "bold",
                        color: parseFloat(row.amount) < 0 ? "red" : "green",
                        borderRight: "1px solid #e0e0e0",
                      }}
                    >
                      {parseFloat(row.amount || 0).toLocaleString(undefined, {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2,
                      })}
                    </TableCell>
                    <TableCell sx={{ borderRight: "1px solid #e0e0e0" }}>
                      {row.vendor || "-"}
                    </TableCell>
                    <TableCell sx={{ borderRight: "1px solid #e0e0e0" }}>
                      {row.description}
                    </TableCell>
                    <TableCell
                      sx={{
                        borderRight: "1px solid #e0e0e0",
                        fontSize: "0.85rem",
                      }}
                    >
                      {row.posting_date}
                    </TableCell>
                    <TableCell
                      sx={{
                        fontWeight: "bold",
                        fontSize: "0.75rem",
                        color:
                          row.status === "Request Change"
                            ? "#e74a3b"
                            : row.status === "Approved"
                            ? "#1cc88a"
                            : row.status === "In Basket"
                            ? "#fd7e14"
                            : row.status === "Draft"
                            ? "#f6c23e"
                            : row.status === "Pending"
                            ? "#4e73df"
                            : row.status === "None"
                            ? "#858796"
                            : "inherit",
                      }}
                    >
                      {row.status || row.audit_status || "-"}
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell
                    colSpan={8}
                    align="center"
                    sx={{ py: 5, color: "text.secondary" }}
                  >
                    No actual data found for selected filters
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>

        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            borderTop: "1px solid #e0e0e0",
            mt: 1,
          }}
        >
          <Box sx={{ pl: 1, display: "flex", gap: 1, alignItems: "center" }}>
            {canAudit && (
              <Box
                sx={{
                  display: "flex",
                  gap: 1.5,
                  alignItems: "center",
                  px: 1.5,
                  py: 0.5,
                  bgcolor: "#f5f5f5",
                  borderRadius: 1,
                  border: "1px solid #e0e0e0",
                  mr: 0.5,
                }}
              >
                {auditComplete ? (
                  <Typography
                    variant="caption"
                    sx={{ fontWeight: 600, color: "#2e7d32" }}
                  >
                    ✓ ตรวจสอบครบแล้ว
                  </Typography>
                ) : (
                  <>
                    <Typography
                      variant="caption"
                      sx={{ fontWeight: 600, color: "#555" }}
                    >
                      ทั้งหมด:{" "}
                      <span style={{ color: "#1976d2" }}>
                        {auditStats.total}
                      </span>
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{ fontWeight: 600, color: "#555" }}
                    >
                      ตรวจสอบแล้ว:{" "}
                      <span style={{ color: "#2e7d32" }}>
                        {auditStats.reviewed}
                      </span>
                    </Typography>
                    <Typography
                      variant="caption"
                      sx={{ fontWeight: 600, color: "#555" }}
                    >
                      คงเหลือ:{" "}
                      <span style={{ color: "#ed6c02" }}>
                        {auditStats.remaining}
                      </span>
                    </Typography>
                  </>
                )}
              </Box>
            )}
            {canAudit && data.length > 0 && (
              <>
                {isOwner && (
                  <Button
                    variant="contained"
                    color="success"
                    size="small"
                    startIcon={<CheckCircleIcon />}
                    onClick={() => {
                      if (rejectedBasket.length === 0) {
                        console.log(filters.month);
                        handleOpenApproveDialog(); // ถ้าตะกร้าว่าง ให้เด้ง Modal ยืนยันเลย
                      } else {
                        setBasketModalOpen(true); // ถ้าตะกร้ามีของ ให้เปิดดูตะกร้าก่อน
                      }
                    }}
                    disabled={auditLoading || auditComplete}
                    sx={{
                      textTransform: "none",
                      borderRadius: "4px",
                      fontSize: "0.75rem",
                    }}
                  >
                    ยืนยันรายการ
                  </Button>
                )}

                <Button
                  variant="contained"
                  color="error"
                  size="small"
                  startIcon={<AnnouncementIcon />}
                  onClick={() => setReportModalOpen(true)}
                  disabled={auditLoading || auditComplete}
                  sx={{
                    textTransform: "none",
                    borderRadius: "4px",
                    fontSize: "0.75rem",
                  }}
                >
                  แจ้งปฏิเสธรายการ
                </Button>

                {/* 🌟 ไอคอนตะกร้าพร้อมตัวเลข Badge ที่อิงจากความยาวของ rejectedBasket */}
                <Tooltip title="ดูตะกร้ารายการปฏิเสธ">
                  <IconButton
                    onClick={() => setBasketModalOpen(true)}
                    disabled={auditComplete}
                    color="error"
                    sx={{
                      bgcolor: "#ffebee",
                      borderRadius: "4px",
                      p: 0.5,
                      "&:hover": { bgcolor: "#ffcdd2" },
                    }}
                  >
                    <Badge badgeContent={rejectedBasket.length} color="error">
                      <ShoppingBasketIcon fontSize="small" />
                    </Badge>
                  </IconButton>
                </Tooltip>
              </>
            )}
          </Box>

          {!loading && data.length > 0 && (
            <TablePagination
              rowsPerPageOptions={[10, 25, 50, 100]}
              component="div"
              count={totalCount}
              rowsPerPage={rowsPerPage}
              page={page}
              onPageChange={handleChangePage}
              onRowsPerPageChange={handleChangeRowsPerPage}
              sx={{
                ".MuiTablePagination-toolbar": { minHeight: "36px", px: 1 },
                ".MuiTablePagination-selectLabel, .MuiTablePagination-input, .MuiTablePagination-displayedRows":
                  { fontSize: "0.75rem" },
              }}
            />
          )}
        </Box>
      </React.Fragment>
    );

    const openFilter = Boolean(anchorEl);

    return (
      <Paper
        sx={{
          p: 1.5,
          pt: 1,
          display: "flex",
          flexDirection: "column",
          borderRadius: 2,
          mb: 2,
          flex: "1 1 0",
          minHeight: { xs: "400px", md: 0 },
          overflow: "hidden",
        }}
      >
        <Box
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            mb: 2,
            px: 1,
          }}
        >
          <Box sx={{ bgcolor: "#043478", px: 2, py: 0.5, borderRadius: "8px" }}>
            <Typography
              variant="h6"
              sx={{ color: "white", fontWeight: "bold" }}
            >
              Actual Detail
            </Typography>
          </Box>
          <Box>
            <Box sx={{ display: "flex", gap: 1, alignItems: "center" }}>
              <IconButton
                onClick={handleFilterClick}
                size="small"
                sx={{ color: "#424242" }}
              >
                <FilterListIcon />
              </IconButton>
              <IconButton
                onClick={() => setOpenFullScreen(true)}
                size="small"
                sx={{ color: "#424242" }}
              >
                <FullscreenIcon />
              </IconButton>
              <IconButton
                size="small"
                onClick={onDownload}
                sx={{ color: "#424242" }}
                disabled={!onDownload || loading || data.length === 0}
              >
                <DownloadIcon sx={{ fontSize: 18 }} />
              </IconButton>
            </Box>
          </Box>
        </Box>

        <Popover
          open={openFilter}
          anchorEl={anchorEl}
          onClose={handleFilterClose}
          anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
          transformOrigin={{ vertical: "top", horizontal: "right" }}
        >
          <Box
            sx={{
              p: 2,
              display: "flex",
              flexDirection: "column",
              gap: 2,
              minWidth: 300,
            }}
          >
            <Typography variant="subtitle2" fontWeight="bold">
              เลือกเดือนที่ต้องการตรวจสอบ
            </Typography>

            <TextField
              label="เลือกเดือน/ปี"
              type="month" // 🌟 เปลี่ยนจาก date เป็น month
              size="small"
              InputLabelProps={{ shrink: true }}
              value={
                tempDateFilter.startDate
                  ? tempDateFilter.startDate.substring(0, 7)
                  : ""
              }
              onChange={(e) => {
                const val = e.target.value; // จะได้ค่า format "YYYY-MM" เช่น "2026-04"
                setTempDateFilter({
                  startDate: `${val}-01`, // เติม -01 ต่อท้ายเพื่อให้ระบบเดิมที่เช็ควันที่ยังทำงานได้
                  endDate: "",
                });
              }}
              fullWidth
            />

            <Box
              sx={{ display: "flex", justifyContent: "space-between", mt: 1 }}
            >
              <Button onClick={handleClearFilter} size="small" color="inherit">
                Clear
              </Button>
              <Button
                onClick={handleApplyFilter}
                size="small"
                variant="contained"
                color="primary"
              >
                Apply
              </Button>
            </Box>
          </Box>
        </Popover>
        {/* <Popover
          open={openFilter}
          anchorEl={anchorEl}
          onClose={handleFilterClose}
          anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
          transformOrigin={{ vertical: "top", horizontal: "right" }}
        >
          <Box
            sx={{
              p: 2,
              display: "flex",
              flexDirection: "column",
              gap: 2,
              minWidth: 300,
            }}
          >
            <Typography variant="subtitle2" fontWeight="bold">
              Filter Date
            </Typography>
            <TextField
              label="Start Date"
              type="date"
              size="small"
              InputLabelProps={{ shrink: true }}
              value={tempDateFilter.startDate}
              onChange={(e) =>
                setTempDateFilter({
                  ...tempDateFilter,
                  startDate: e.target.value,
                })
              }
              fullWidth
            />
            <TextField
              label="End Date"
              type="date"
              size="small"
              InputLabelProps={{ shrink: true }}
              value={tempDateFilter.endDate}
              onChange={(e) =>
                setTempDateFilter({
                  ...tempDateFilter,
                  endDate: e.target.value,
                })
              }
              fullWidth
            />
            <Box
              sx={{ display: "flex", justifyContent: "space-between", mt: 1 }}
            >
              <Button onClick={handleClearFilter} size="small" color="inherit">
                Clear
              </Button>
              <Button
                onClick={handleApplyFilter}
                size="small"
                variant="contained"
                color="primary"
              >
                Apply
              </Button>
            </Box>
          </Box>
        </Popover> */}
        {renderTable(false)}
        <Dialog
          open={openFullScreen}
          onClose={() => setOpenFullScreen(false)}
          fullScreen
        >
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              bgcolor: "#043478",
              p: 2,
            }}
          >
            <Typography
              variant="h6"
              sx={{ color: "white", fontWeight: "bold" }}
            >
              Actual Detail (Full Screen)
            </Typography>
            <Box>
              <IconButton
                onClick={handleFilterClick}
                sx={{ color: "white", mr: 2 }}
              >
                <FilterListIcon />
              </IconButton>
              <IconButton
                onClick={() => setOpenFullScreen(false)}
                sx={{ color: "white" }}
              >
                <CloseIcon />
              </IconButton>
            </Box>
          </Box>
          <DialogContent
            sx={{ p: 2, display: "flex", flexDirection: "column" }}
          >
            {renderTable(true)}
          </DialogContent>
        </Dialog>
        <AuditReportModal
          open={reportModalOpen}
          onClose={() => setReportModalOpen(false)}
          filters={filters}
          onSubmit={handleReportSubmit}
          loading={auditLoading}
          basketItems={rejectedBasket}
        />
        {/* 🌟 เรียกใช้ BasketModal และโยน state ลงไปให้ */}
        <BasketModal
          open={basketModalOpen}
          onClose={() => setBasketModalOpen(false)}
          basketItems={rejectedBasket}
          onRemove={handleRemoveFromBasket}
          onApprove={handleOpenApproveDialog}
          onUpdateNote={handleUpdateBasketNote}
          canApprove={isOwner}
          loading={auditLoading}
        />
        <ApproveConfirmModal
          open={approveDialogOpen}
          onClose={() => setApproveDialogOpen(false)}
          onConfirm={executeApprove}
          loading={auditLoading}
          filters={{
            ...filters,
            year: getApproveTarget().year,
            month: getApproveTarget().month,
          }}
          basketCount={rejectedBasket.length}
        />
      </Paper>
    );
  }
);

export default ActualTable;
