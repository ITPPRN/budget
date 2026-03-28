import React, { useState, useEffect } from 'react';
import {
    Box, Grid, Paper, Typography, Button, TextField,
    MenuItem, Select, InputAdornment, Table, TableBody,
    TableCell, TableContainer, TableHead, TableRow,
    IconButton, Chip, Switch, Pagination, CircularProgress,
    Dialog, DialogTitle, DialogContent, DialogActions,
    Divider
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import CloseIcon from '@mui/icons-material/Close';
import ManageAccountsIcon from '@mui/icons-material/ManageAccounts';
import api from '../../utils/api/axiosInstance';
import { toast } from 'react-toastify';
import { useAuth } from '../../hooks/useAuth';

const OwnerUserManagePage = () => {
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(false);
    const [totalUsers, setTotalUsers] = useState(0);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [search, setSearch] = useState('');
    const [statusFilter, setStatusFilter] = useState('ALL');
    const [deptFilter, setDeptFilter] = useState('ALL');
    const [roleFilter, setRoleFilter] = useState('ALL');

    // Modal State
    const [openModal, setOpenModal] = useState(false);
    const [selectedUser, setSelectedUser] = useState(null);
    const [departments, setDepartments] = useState([]);
    const [permissions, setPermissions] = useState([]); // List of perms for selected user
    const [isSaving, setIsSaving] = useState(false);

    // 1. Get Current User Role
    const { user: currentUser } = useAuth();

    const getCurrentUserRole = () => {
        const roles = currentUser?.roles || currentUser?.role || [];
        if (roles.some(r => r.toUpperCase().includes('ADMIN'))) return 'ADMIN';
        if (roles.some(r => r.toUpperCase().includes('OWNER'))) return 'OWNER';
        if (roles.some(r => r.toUpperCase().includes('DELEGATE'))) return 'DELEGATE';
        return 'USER';
    };

    const currentUserRole = getCurrentUserRole();
    const isDelegate = currentUserRole === 'DELEGATE';

    const getAllowedRoles = () => {
        // Owner can assign Delegate
        if (currentUserRole === 'OWNER') return ['DELEGATE'];
        // Delegate cannot assign roles (usually)
        return [];
    };
    const allowedRoles = getAllowedRoles();

    // Fetch users
    const fetchUsers = async () => {
        setLoading(true);
        try {
            // Owner logic: might be same endpoint but backend filters, or different endpoint?
            // For now using same endpoint, assuming backend handles security based on token
            const response = await api.get('/auth/manage/users', {
                params: { 
                    page, 
                    size: pageSize, 
                    search: search,
                    status: statusFilter,
                    department_code: deptFilter,
                    role: roleFilter
                }
            });
            const { users, total } = response.data.data || response.data;
            setUsers(users || []);
            setTotalUsers(total || 0);
        } catch (error) {
            console.error('Failed to fetch users:', error);
            toast.error('ไม่สามารถโหลดข้อมูลผู้ใช้ได้');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUsers();
    }, [page, pageSize, search, statusFilter, deptFilter, roleFilter]);

    const handleSearchKeyPress = (e) => {
        if (e.key === 'Enter') {
            setPage(1);
            fetchUsers();
        }
    };

    // Fetch departments (for modal)
    const fetchDepartments = async () => {
        try {
            const response = await api.get('/auth/manage/departments');
            setDepartments(Array.isArray(response.data.data) ? response.data.data : []);
        } catch (error) {
            console.error('Failed to fetch departments:', error);
        }
    };

    useEffect(() => {
        fetchDepartments();
    }, []);

    const handleOpenModal = async (user) => {
        setSelectedUser(user);
        setPermissions([]); // Clear stale data first
        setOpenModal(true);
        try {
            const response = await api.get(`/auth/manage/users/${user.id}/permissions`);
            setPermissions(Array.isArray(response.data.data) ? response.data.data : []);
        } catch (error) {
            console.error('Failed to fetch user permissions:', error);
            toast.error('ไม่สามารถโหลดข้อมูลสิทธิ์ของผู้ใช้ได้');
        }
    };

    const handleCloseModal = () => {
        setOpenModal(false);
        setSelectedUser(null);
        setPermissions([]);
    };

    const updatePermission = (deptCode, fieldUpdates) => {
        setPermissions(prev => {
            const existing = prev.find(p => p.department_code === deptCode);
            if (existing) {
                return prev.map(p => p.department_code === deptCode ? { ...p, ...fieldUpdates } : p);
            }
            return [...prev, { department_code: deptCode, role: 'DELEGATE', is_active: false, ...fieldUpdates }];
        });
    };

    const handleSavePermissions = async () => {
        if (!selectedUser) return;
        setIsSaving(true);
        try {
            const finalPermissions = permissions.filter(p => p.role && p.role !== '');
            await api.put(`/auth/manage/users/${selectedUser.id}/permissions`, finalPermissions);
            toast.success('บันทึกสิทธิ์เรียบร้อยแล้ว');
            handleCloseModal();
            fetchUsers();
        } catch (error) {
            console.error('Failed to save permissions:', error);
            toast.error('บันทึกผิดพลาด: ' + (error.response?.data?.message || error.message));
        } finally {
            setIsSaving(false);
        }
    };

    const getStatusChip = (user) => {
        const hasActivePerm = user.permissions?.some(p => p.is_active) || false;
        if (hasActivePerm) {
            return <Chip label="มีสิทธิ์เข้าถึง" size="small" sx={{ bgcolor: '#27ae60', color: 'white', fontWeight: 'bold', borderRadius: '4px', fontSize: '11px' }} />;
        }
        return <Chip label="ไม่มีสิทธิ์เข้าถึง" size="small" sx={{ bgcolor: '#e74c3c', color: 'white', fontWeight: 'bold', borderRadius: '4px', fontSize: '11px' }} />;
    };

    // Modal Hierarchy Restrictions
    const isTargetAdmin = selectedUser?.roles?.some(r => r.toUpperCase() === 'ADMIN') ||
        selectedUser?.permissions?.some(p => p.is_active && p.role?.toUpperCase() === 'ADMIN');
    const isTargetOwner = selectedUser?.roles?.some(r => r.toUpperCase() === 'OWNER') ||
        selectedUser?.permissions?.some(p => p.is_active && p.role?.toUpperCase() === 'OWNER');
    const isTargetDelegate = selectedUser?.roles?.some(r => r.toUpperCase() === 'DELEGATE') ||
        selectedUser?.permissions?.some(p => p.is_active && p.role?.toUpperCase() === 'DELEGATE');

    let canModifyModal = false;
    if (!isDelegate && selectedUser) {
        if (currentUserRole === 'ADMIN') {
            canModifyModal = isTargetOwner || (!isTargetAdmin && !isTargetDelegate);
        } else if (currentUserRole === 'OWNER') {
            canModifyModal = isTargetDelegate || (!isTargetAdmin && !isTargetOwner);
        }
    }

    return (
        <Box sx={{ p: 4, bgcolor: '#f8fafc', minHeight: '100vh' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 4 }}>
                <Box sx={{ width: 4, height: 24, bgcolor: '#0288d1', mr: 2, borderRadius: 2 }} />
                <Typography variant="h5" sx={{ fontWeight: 800, color: '#1e293b' }}>
                    จัดการสิทธิ์เข้าถึง
                </Typography>
            </Box>

            {/* Filters */}
            <Paper sx={{ p: 2, mb: 3, borderRadius: '12px', boxShadow: '0 1px 3px rgba(0,0,0,0.1)' }}>
                <Grid container spacing={2} alignItems="flex-end">
                    <Grid item xs={12} md={6}>
                        <Box sx={{ display: 'flex', gap: 1, alignItems: 'flex-end' }}>
                            <Box sx={{ flexGrow: 1 }}>
                                <Typography variant="caption" sx={{ fontWeight: 700, mb: 1, display: 'block', color: '#64748b' }}>ค้นหาผู้ใช้</Typography>
                                <TextField
                                    fullWidth size="small" placeholder="ค้นหาด้วยชื่อ หรือ รหัสพนักงาน..."
                                    value={search} onChange={(e) => setSearch(e.target.value)}
                                    onKeyPress = {handleSearchKeyPress}
                                    InputProps={{
                                        startAdornment: <InputAdornment position="start"><SearchIcon fontSize="small" sx={{ color: '#94a3b8' }} /></InputAdornment>,
                                        sx: { borderRadius: '8px', bgcolor: '#f1f5f9', '& fieldset': { border: 'none' }, fontSize: '14px' }
                                    }}
                                />
                            </Box>
                            <Button
                                variant="contained"
                                onClick={() => { setPage(1); fetchUsers(); }}
                                sx={{ bgcolor: '#0288d1', height: '40px', borderRadius: '8px', fontWeight: 'bold', px: 3, '&:hover': { bgcolor: '#0277bd' } }}
                            >
                                ค้นหา
                            </Button>
                        </Box>
                    </Grid>

                    <Grid item xs={12} md={1.4}>
                        <Typography variant="caption" sx={{ fontWeight: 700, mb: 1, display: 'block', color: '#64748b' }}>สถานะ</Typography>
                        <Select
                            fullWidth size="small"
                            value={statusFilter}
                            onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
                            sx={{ borderRadius: '8px', bgcolor: '#f1f5f9', '& .MuiOutlinedInput-notchedOutline': { border: 'none' }, fontSize: '13px' }}
                        >
                            <MenuItem value="ALL">ทั้งหมด</MenuItem>
                            <MenuItem value="ACTIVE">มีสิทธิ์เข้าถึง</MenuItem>
                            <MenuItem value="INACTIVE">ไม่มีสิทธิ์เข้าถึง</MenuItem>
                        </Select>
                    </Grid>

                    <Grid item xs={12} md={1.4}>
                        <Typography variant="caption" sx={{ fontWeight: 700, mb: 1, display: 'block', color: '#64748b' }}>แผนก</Typography>
                        <Select
                            fullWidth size="small"
                            value={deptFilter}
                            onChange={(e) => { setDeptFilter(e.target.value); setPage(1); }}
                            sx={{ borderRadius: '8px', bgcolor: '#f1f5f9', '& .MuiOutlinedInput-notchedOutline': { border: 'none' }, fontSize: '13px' }}
                        >
                            <MenuItem value="ALL">ทั้งหมด</MenuItem>
                            {departments.map(d => (
                                <MenuItem key={d.id} value={d.code}>{d.code}</MenuItem>
                            ))}
                        </Select>
                    </Grid>

                    <Grid item xs={12} md={1}>
                        {/* Spacer to shift Role filter to the right */}
                    </Grid>

                    <Grid item xs={12} md={1.2}>
                        <Typography variant="caption" sx={{ fontWeight: 700, mb: 1, display: 'block', color: '#64748b' }}>สิทธิ์</Typography>
                        <Select
                            fullWidth size="small"
                            value={roleFilter}
                            onChange={(e) => { setRoleFilter(e.target.value); setPage(1); }}
                            sx={{ borderRadius: '8px', bgcolor: '#f1f5f9', '& .MuiOutlinedInput-notchedOutline': { border: 'none' }, fontSize: '13px' }}
                        >
                            <MenuItem value="ALL">ทั้งหมด</MenuItem>
                            <MenuItem value="ADMIN">ADMIN</MenuItem>
                            <MenuItem value="OWNER">OWNER</MenuItem>
                            <MenuItem value="DELEGATE">DELEGATE</MenuItem>
                        </Select>
                    </Grid>

                    <Grid item xs={12} md={1}>
                        <Button
                            variant="outlined"
                            fullWidth
                            onClick={() => { 
                                setSearch('');
                                setStatusFilter('ALL');
                                setDeptFilter('ALL');
                                setRoleFilter('ALL');
                                setPage(1);
                            }}
                            sx={{ height: '40px', borderRadius: '8px', fontWeight: 'bold', borderColor: '#cbd5e1', color: '#64748b' }}
                        >
                            ล้าง
                        </Button>
                    </Grid>
                </Grid>
            </Paper>

            {/* Table */}
            <TableContainer component={Paper} sx={{ borderRadius: '12px', overflow: 'hidden', boxShadow: '0 4px 20px rgba(0,0,0,0.05)' }}>
                <Table sx={{ minWidth: 650 }}>
                    <TableHead sx={{ bgcolor: '#0288d1' }}>
                        <TableRow>
                            <TableCell sx={{ color: 'white', fontWeight: 600, py: 2 }}>รหัสพนักงาน</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 600 }}>ชื่อผู้ใช้งาน</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 600 }}>การจัดการ</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 600 }}>สถานะ</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 600 }}>แผนกหลัก</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 600 }}>สิทธิ์แผนก</TableCell>
                            <TableCell sx={{ color: 'white', fontWeight: 600 }}>สิทธิ์ระบบ</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {loading ? (
                            <TableRow><TableCell colSpan={6} align="center" sx={{ py: 6 }}><CircularProgress size={28} sx={{ color: '#0288d1' }} /></TableCell></TableRow>
                        ) : users.length === 0 ? (
                            <TableRow><TableCell colSpan={6} align="center" sx={{ py: 6 }}><Typography color="text.secondary">ไม่พบข้อมูลสมาชิก</Typography></TableCell></TableRow>
                        ) : users.map((user) => (
                            <TableRow key={user.id} hover sx={{ '&:last-child td, &:last-child th': { border: 0 } }}>
                                <TableCell sx={{ fontWeight: 700, color: '#334155' }}>{user.username}</TableCell>
                                <TableCell sx={{ color: '#475569' }}>
                                    {user.name_th}
                                    {user.id === currentUser?.id && (
                                        <Chip label="คุณ" size="small" sx={{ ml: 1, height: 20, fontSize: '10px', bgcolor: '#f1f5f9', color: '#64748b' }} />
                                    )}
                                </TableCell>
                                <TableCell>
                                    {(() => {
                                        const isTargetAdmin = user.roles?.some(r => r.toUpperCase() === 'ADMIN') ||
                                            user.permissions?.some(p => p.is_active && p.role?.toUpperCase() === 'ADMIN');
                                        const isTargetOwner = user.roles?.some(r => r.toUpperCase() === 'OWNER') ||
                                            user.permissions?.some(p => p.is_active && p.role?.toUpperCase() === 'OWNER');
                                        const isTargetDelegate = (user.roles?.some(r => r.toUpperCase() === 'DELEGATE')) ||
                                            (user.permissions?.some(p => p.is_active && p.role?.toUpperCase() === 'DELEGATE'));
                                        const isSelf = user.id === currentUser?.id;

                                        // Hierarchical Rules (Synchronized with UserManage.jsx):
                                        // 1. Delegates can NEVER manage anyone.
                                        // 2. Self cannot be managed.
                                        // 3. Admin: Can manage Owners and Regular Users. Cannot manage Delegates or other Admins.
                                        // 4. Owner: Can manage Delegates and Regular Users. Cannot manage Admins or other Owners.

                                        let canManage = false;
                                        if (!isSelf && !isDelegate) {
                                            if (currentUserRole === 'ADMIN') {
                                                canManage = isTargetOwner || (!isTargetAdmin && !isTargetDelegate);
                                            } else if (currentUserRole === 'OWNER') {
                                                canManage = isTargetDelegate || (!isTargetAdmin && !isTargetOwner);
                                            }
                                        }

                                        const shouldDisable = !canManage;

                                        return (
                                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                                <Typography variant="caption" sx={{ color: shouldDisable ? '#cbd5e1' : '#94a3b8' }}>จัดการ</Typography>
                                                <Switch
                                                    checked={user.permissions?.some(p => p.is_active) || false}
                                                    color="primary"
                                                    onClick={() => !shouldDisable && handleOpenModal(user)}
                                                    disabled={shouldDisable}
                                                    sx={{
                                                        opacity: shouldDisable ? 0.5 : 1,
                                                        '& .MuiSwitch-switchBase.Mui-checked': { color: '#0288d1' },
                                                        '& .MuiSwitch-switchBase.Mui-checked + .MuiSwitch-track': { bgcolor: '#0288d1' }
                                                    }}
                                                />
                                            </Box>
                                        );
                                    })()}
                                </TableCell>
                                <TableCell>{getStatusChip(user)}</TableCell>
                                <TableCell sx={{ color: '#64748b' }}>
                                    {user.mapped_department || '-'}
                                </TableCell>
                                <TableCell sx={{ color: '#64748b' }}>
                                    {(() => {
                                        const activePerms = user.permissions?.filter(p => p.is_active) || [];
                                        if (activePerms.length === 0) return '-';
                                        return (
                                            <Select
                                                value=""
                                                displayEmpty
                                                size="small"
                                                variant="standard"
                                                sx={{
                                                    fontSize: '11px',
                                                    minWidth: 100,
                                                    '&:before, &:after': { display: 'none' },
                                                    '& .MuiSelect-select': { py: 0.5, px: 1, bgcolor: '#f1f5f9', borderRadius: '4px' }
                                                }}
                                                renderValue={() => `${activePerms.length} แผนก`}
                                            >
                                                {activePerms.map(p => (
                                                    <MenuItem key={p.department_code} disabled sx={{ fontSize: '12px', opacity: '1 !important', color: '#1e293b' }}>
                                                        <Box sx={{ display: 'flex', justifyContent: 'space-between', width: '100%' }}>
                                                            <Typography variant="caption" sx={{ fontWeight: 700 }}>{p.department_code}</Typography>
                                                            <Typography variant="caption" sx={{ color: '#64748b', ml: 2 }}>{p.role}</Typography>
                                                        </Box>
                                                    </MenuItem>
                                                ))}
                                            </Select>
                                        );
                                    })()}
                                </TableCell>
                                <TableCell>
                                    {/* Show Highest Role */}
                                    {(() => {
                                        const roles = user.roles || [];
                                        let displayRole = 'USER';
                                        if (roles.some(r => r.toUpperCase() === 'ADMIN')) displayRole = 'ADMIN';
                                        else if (roles.some(r => r.toUpperCase() === 'OWNER')) displayRole = 'OWNER';
                                        else if (roles.some(r => r.toUpperCase() === 'DELEGATE')) displayRole = 'DELEGATE';
                                        else if (roles.length > 0 && roles[0] !== 'USER') displayRole = roles[0];

                                        if (displayRole === 'USER' && roles.includes('USER')) {
                                            return <Typography sx={{ color: '#cbd5e1', fontWeight: 700 }}>-</Typography>;
                                        }

                                        return (
                                            <Chip
                                                label={displayRole}
                                                variant="outlined"
                                                size="small"
                                                sx={{
                                                    color: displayRole === 'ADMIN' ? '#d32f2f' : '#0288d1',
                                                    borderColor: displayRole === 'ADMIN' ? '#ef5350' : '#b3e5fc',
                                                    bgcolor: displayRole === 'ADMIN' ? '#ffebee' : '#e1f5fe',
                                                    fontWeight: 700,
                                                    fontSize: '11px'
                                                }}
                                            />
                                        );
                                    })()}
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>

                <Box sx={{ p: 2, display: 'flex', justifyContent: 'flex-end', alignItems: 'center', gap: 3, bgcolor: '#f1f5f9' }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600 }}>Rows per page:</Typography>
                        <Select size="small" variant="standard" value={pageSize} onChange={(e) => setPageSize(e.target.value)} sx={{ fontSize: '12px', fontWeight: 600 }}>
                            <MenuItem value={10}>10</MenuItem>
                            <MenuItem value={20}>20</MenuItem>
                            <MenuItem value={50}>50</MenuItem>
                        </Select>
                    </Box>
                    <Typography variant="caption" sx={{ color: '#64748b', fontWeight: 600 }}>
                        {((page - 1) * pageSize) + 1}-{Math.min(page * pageSize, totalUsers)} of {totalUsers}
                    </Typography>
                    <Pagination count={Math.ceil(totalUsers / pageSize)} page={page} onChange={(e, v) => setPage(v)} size="small" shape="rounded" color="primary" />
                </Box>
            </TableContainer>

            {/* Permissions Modal */}
            <Dialog open={openModal} onClose={handleCloseModal} maxWidth="md" fullWidth PaperProps={{ sx: { borderRadius: '16px' } }}>
                <DialogTitle sx={{ bgcolor: '#0288d1', color: 'white', p: 2, display: 'flex', alignItems: 'center' }}>
                    <Box sx={{ bgcolor: 'rgba(255,255,255,0.2)', px: 2, py: 0.5, borderRadius: '8px', mr: 2 }}>
                        <Typography sx={{ fontWeight: 800, fontSize: '0.85rem', letterSpacing: 0.5 }}>ACCESS CONTROL (OWNER)</Typography>
                    </Box>
                    <Typography sx={{ flexGrow: 1, fontWeight: 700 }}>
                        {selectedUser?.username} - {selectedUser?.name_th}
                    </Typography>
                    <IconButton onClick={handleCloseModal} sx={{ color: 'white' }}><CloseIcon /></IconButton>
                </DialogTitle>
                <DialogContent sx={{ p: 0 }}>
                    <Box sx={{ p: 2, bgcolor: '#f8fafc', borderBottom: '1px solid #e2e8f0' }}>
                        <Typography variant="body2" sx={{ color: '#64748b', fontWeight: 500 }}>กําหนดสิทธิ์ผู้ช่วย (Delegate) ในแผนกของคุณ</Typography>
                    </Box>
                    <TableContainer sx={{ maxHeight: 400 }}>
                        <Table stickyHeader size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell sx={{ bgcolor: '#f1f5f9', fontWeight: 700, color: '#475569' }}>แผนก (Department)</TableCell>
                                    <TableCell sx={{ bgcolor: '#f1f5f9', fontWeight: 700, color: '#475569' }}>บทบาท (Role)</TableCell>
                                    <TableCell sx={{ bgcolor: '#f1f5f9', fontWeight: 700, color: '#475569' }}>สถานะการใช้งาน</TableCell>
                                    <TableCell sx={{ bgcolor: '#f1f5f9', fontWeight: 700, color: '#475569' }}>แสดงผล</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {departments.map((dept) => {
                                    const perm = permissions.find(p => p.department_code === dept.code);
                                    const isActive = perm?.is_active || false;

                                    return (
                                        <TableRow key={dept.id} hover>
                                            <TableCell sx={{ fontWeight: 700, color: '#1e293b' }}>{dept.code}</TableCell>
                                            <TableCell>
                                                <Select
                                                    fullWidth size="small" value={perm?.role || ''}
                                                    disabled={!canModifyModal}
                                                    onChange={(e) => {
                                                        const newRole = e.target.value;
                                                        const updates = { role: newRole };
                                                        if (newRole && !perm?.role) updates.is_active = true;
                                                        if (!newRole) updates.is_active = false;
                                                        updatePermission(dept.code, updates);
                                                    }}
                                                    displayEmpty sx={{ borderRadius: '8px', bgcolor: 'white', opacity: !canModifyModal ? 0.7 : 1 }}
                                                >
                                                    <MenuItem value="">เลือกบทบาท...</MenuItem>
                                                    {allowedRoles.includes('OWNER') && <MenuItem value="OWNER">Owner</MenuItem>}
                                                    {allowedRoles.includes('DELEGATE') && <MenuItem value="DELEGATE">Delegate</MenuItem>}
                                                </Select>
                                            </TableCell>
                                            <TableCell>
                                                <Switch
                                                    checked={isActive}
                                                    disabled={!canModifyModal}
                                                    onChange={(e) => {
                                                        const active = e.target.checked;
                                                        const updates = { is_active: active };
                                                        if (!active) {
                                                            updates.role = ''; // Reset role when turned OFF
                                                        } else if (!perm?.role && allowedRoles.length > 0) {
                                                            updates.role = allowedRoles[0]; // Auto-select first allowed role when turned ON
                                                        }
                                                        updatePermission(dept.code, updates);
                                                    }}
                                                />
                                            </TableCell>
                                            <TableCell>
                                                {isActive ? (
                                                    <Chip label={`เปิดสิทธิ์ ${perm?.role || ''}`} size="small" sx={{ bgcolor: '#27ae60', color: 'white', fontWeight: 'bold' }} />
                                                ) : (
                                                    <Chip
                                                        label="ไม่มีสิทธิ์"
                                                        size="small" variant="outlined"
                                                        sx={{ color: '#94a3b8', borderColor: '#e2e8f0' }}
                                                    />
                                                )}
                                            </TableCell>
                                        </TableRow>
                                    );
                                })}
                            </TableBody>
                        </Table>
                    </TableContainer>
                </DialogContent>
                <DialogActions sx={{ p: 3, bgcolor: '#f8fafc' }}>
                    <Button variant="outlined" onClick={handleCloseModal} sx={{ borderRadius: '8px', color: '#64748b', borderColor: '#e2e8f0' }}>ยกเลิก</Button>
                    <Button
                        variant="contained"
                        onClick={handleSavePermissions}
                        disabled={isSaving || !canModifyModal}
                        sx={{ bgcolor: '#0288d1', borderRadius: '8px', px: 4, fontWeight: 'bold' }}
                    >
                        {isSaving ? <CircularProgress size={20} color="inherit" /> : 'บันทึกข้อมูลสิทธิ์'}
                    </Button>
                </DialogActions>
            </Dialog>
        </Box>
    );
};

export default OwnerUserManagePage;
