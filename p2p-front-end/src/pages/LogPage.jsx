import React, { useState, useEffect } from 'react';
import { 
    Box, 
    Typography, 
    Paper, 
    Container, 
    Table, 
    TableHead, 
    TableRow, 
    TableCell, 
    TableBody, 
    Chip, 
    IconButton, 
    Tooltip,
    TextField,
    MenuItem,
    Grid,
    Button,
    CircularProgress,
    TableContainer,
    Card,
    CardContent,
    Avatar,
    Stack,
    useMediaQuery,
    useTheme,
    Fade,
    FormControl,
    InputLabel,
    Select
} from '@mui/material';
import HistoryIcon from '@mui/icons-material/History';
import VisibilityIcon from '@mui/icons-material/Visibility';
import RefreshIcon from '@mui/icons-material/Refresh';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import BusinessIcon from '@mui/icons-material/Business';
import PersonIcon from '@mui/icons-material/Person';
import CalendarMonthIcon from '@mui/icons-material/CalendarMonth';
import FilterListIcon from '@mui/icons-material/FilterList';
import api from '../utils/api/axiosInstance';
import AuditDetailsModal from '../components/Budget/AuditDetailsModal';

const LogPage = () => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('md'));
    const [logs, setLogs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedLog, setSelectedLog] = useState(null);
    const [detailModalOpen, setDetailModalOpen] = useState(false);

    // Filter Options
    const [availableDepts, setAvailableDepts] = useState([]);
    const [availableMonths, setAvailableMonths] = useState([]);

    // Filters
    const [filters, setFilters] = useState({
        department: '',
        month: ''
    });

    const fetchLogs = async (currentFilters = filters) => {
        setLoading(true);
        try {
            const queryParams = new URLSearchParams(
                Object.entries(currentFilters).filter(([_, v]) => v !== '')
            ).toString();
            const res = await api.get(`/budgets/audit/logs?${queryParams}`);
            const data = res.data || [];
            setLogs(data);

            // Populate filter options from "real data" if not already set
            if (availableDepts.length === 0 || availableMonths.length === 0) {
                const depts = [...new Set(data.map(l => l.department_code || l.department))].sort();
                const months = [...new Set(data.map(l => l.month))].sort((a,b) => parseInt(a) - parseInt(b));
                if (depts.length > 0) setAvailableDepts(depts);
                if (months.length > 0) setAvailableMonths(months);
            }
        } catch (err) {
            console.error("Fetch Logs Error:", err);
        } finally {
            setLoading(false);
        }
    };

    // Initial fetch
    useEffect(() => {
        fetchLogs();
    }, []);

    // Auto-filter on filter change
    useEffect(() => {
        // Debounce slightly to avoid flicker if multiple states change rapidly
        const timeout = setTimeout(() => {
            fetchLogs(filters);
        }, 100);
        return () => clearTimeout(timeout);
    }, [filters.department, filters.month]);

    const handleViewDetails = (log) => {
        setSelectedLog(log);
        setDetailModalOpen(true);
    };

    const getStatusStyles = (status) => {
        switch (status?.toUpperCase()) {
            case 'APPROVED':
            case 'CONFIRMED':
                return {
                    label: 'APPROVED',
                    color: 'success',
                    gradient: 'linear-gradient(135deg, #4caf50 0%, #2e7d32 100%)',
                    icon: <CheckCircleIcon fontSize="small" />
                };
            case 'REPORTED':
            case 'REJECTED':
                return {
                    label: 'REPORTED',
                    color: 'error',
                    gradient: 'linear-gradient(135deg, #f44336 0%, #d32f2f 100%)',
                    icon: <ErrorIcon fontSize="small" />
                };
            default:
                return {
                    label: status,
                    color: 'default',
                    gradient: 'linear-gradient(135deg, #9e9e9e 0%, #757575 100%)',
                    icon: null
                };
        }
    };

    const renderStatusChip = (status) => {
        const styles = getStatusStyles(status);
        return (
            <Chip 
                icon={styles.icon} 
                label={styles.label} 
                sx={{ 
                    background: styles.gradient, 
                    color: 'white', 
                    fontWeight: 'bold',
                    boxShadow: '0 4px 10px rgba(0,0,0,0.1)',
                    '& .MuiChip-icon': { color: 'white' }
                }} 
                size="small" 
            />
        );
    };

    return (
        <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
            {/* Header Section */}
            <Box sx={{ 
                display: 'flex', 
                flexDirection: { xs: 'column', sm: 'row' }, 
                alignItems: { xs: 'flex-start', sm: 'center' }, 
                justifyContent: 'space-between', 
                mb: 4, 
                gap: 2 
            }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <Avatar sx={{ bgcolor: '#043478', width: 56, height: 56, boxShadow: '0 8px 16px rgba(4,52,120,0.2)' }}>
                        <HistoryIcon sx={{ fontSize: 32 }} />
                    </Avatar>
                    <Box>
                        <Typography variant="h4" fontWeight="900" sx={{ color: '#043478', letterSpacing: -1 }}>
                            Audit Logs
                        </Typography>
                        <Typography variant="body2" color="textSecondary">
                            ตรวจสอบประวัติการแจ้งแก้ไขและอนุมัติงบประมาณ
                        </Typography>
                    </Box>
                </Box>
                <Button 
                    startIcon={<RefreshIcon />} 
                    variant="contained" 
                    onClick={fetchLogs}
                    disabled={loading}
                    sx={{ 
                        bgcolor: '#043478', 
                        borderRadius: 3, 
                        px: 3, 
                        py: 1, 
                        textTransform: 'none', 
                        fontWeight: 'bold',
                        boxShadow: '0 8px 20px rgba(4,52,120,0.3)',
                        '&:hover': { bgcolor: '#032a61' }
                    }}
                >
                    Refresh Data
                </Button>
            </Box>

            {/* Filter Section */}
            <Paper 
                elevation={0} 
                sx={{ 
                    p: 4, 
                    mb: 3, 
                    borderRadius: 4, 
                    border: '1px solid #eef0f2',
                    boxShadow: '0 4px 20px rgba(0,0,0,0.02)'
                }}
            >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 3 }}>
                    <FilterListIcon color="primary" sx={{ fontSize: 24 }} />
                    <Typography variant="h6" fontWeight="900" color="textSecondary" sx={{ letterSpacing: -0.5 }}>FILTERS</Typography>
                </Box>
                
                <Grid container spacing={4} alignItems="center">
                    <Grid item xs={12} md={5}>
                        <FormControl fullWidth size="small">
                            <InputLabel id="dept-label" sx={{ fontWeight: 'bold' }}>Department</InputLabel>
                            <Select
                                labelId="dept-label"
                                label="Department"
                                value={filters.department}
                                onChange={(e) => setFilters({...filters, department: e.target.value})}
                                sx={{ 
                                    borderRadius: 3,
                                    bgcolor: '#f8f9fa',
                                    fontWeight: '800',
                                    color: '#043478',
                                    '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e9ecef' },
                                    minWidth: 280
                                }}
                            >
                                <MenuItem value="" sx={{ fontWeight: 'bold' }}>All Departments (แสดงข้อมูลทั้งหมด)</MenuItem>
                                {availableDepts.map(d => (
                                    <MenuItem key={d} value={d} sx={{ fontWeight: '600' }}>{d}</MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                    </Grid>

                    <Grid item xs={12} md={4}>
                        <FormControl fullWidth size="small">
                            <InputLabel id="month-label" sx={{ fontWeight: 'bold' }}>Month</InputLabel>
                            <Select
                                labelId="month-label"
                                label="Month"
                                value={filters.month}
                                onChange={(e) => setFilters({...filters, month: e.target.value})}
                                sx={{ 
                                    borderRadius: 3,
                                    bgcolor: '#f8f9fa',
                                    fontWeight: '800',
                                    color: '#043478',
                                    '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e9ecef' },
                                    minWidth: 200
                                }}
                            >
                                <MenuItem value="" sx={{ fontWeight: 'bold' }}>All Months (ทุกเดือน)</MenuItem>
                                {availableMonths.map(m => (
                                    <MenuItem key={m} value={m} sx={{ fontWeight: '600' }}>
                                        {m === "01" ? "JAN" : 
                                         m === "02" ? "FEB" :
                                         m === "03" ? "MAR" :
                                         m === "04" ? "APR" :
                                         m === "05" ? "MAY" :
                                         m === "06" ? "JUN" :
                                         m === "07" ? "JUL" :
                                         m === "08" ? "AUG" :
                                         m === "09" ? "SEP" :
                                         m === "10" ? "OCT" :
                                         m === "11" ? "NOV" : "DEC"} ({m})
                                    </MenuItem>
                                ))}
                            </Select>
                        </FormControl>
                    </Grid>

                    {logs.length > 0 && (
                        <Grid item xs={12} md={3} sx={{ textAlign: 'right' }}>
                            <Typography variant="caption" fontWeight="900" color="primary" sx={{ bgcolor: '#f0f4ff', px: 2, py: 1, borderRadius: 2 }}>
                                ผลลัพธ์: {logs.length} รายการ
                            </Typography>
                        </Grid>
                    )}
                </Grid>
            </Paper>

            {/* Logs Section */}
            {loading ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', py: 10 }}>
                    <CircularProgress color="primary" />
                </Box>
            ) : logs.length === 0 ? (
                <Paper sx={{ p: 10, textAlign: 'center', borderRadius: 4, bgcolor: '#f8f9fa', border: '2px dashed #dee2e6' }}>
                    <HistoryIcon sx={{ fontSize: 60, color: '#ced4da', mb: 2 }} />
                    <Typography variant="h6" color="textSecondary">No audit logs found</Typography>
                    <Typography variant="body2" color="textSecondary">ไม่พบข้อมูลประวัติในช่วงเวลาที่เลือก</Typography>
                </Paper>
            ) : (
                <>
                    {isMobile ? (
                        <Grid container spacing={2}>
                            {logs.map((log, idx) => (
                                <Grid item xs={12} key={log.id}>
                                    <Fade in timeout={300 + idx * 100}>
                                        <Card sx={{ borderRadius: 4, overflow: 'hidden', border: '1px solid #f1f3f5' }}>
                                            <CardContent sx={{ p: 2.5 }}>
                                                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                                                    {renderStatusChip(log.status)}
                                                    <Typography variant="caption" color="textSecondary">
                                                        {new Date(log.created_at).toLocaleDateString('th-TH')}
                                                    </Typography>
                                                </Box>
                                                <Stack spacing={1.5}>
                                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                                                        <BusinessIcon color="action" fontSize="small" />
                                                        <Box>
                                                            <Typography variant="subtitle2" fontWeight="800">
                                                                {log.department_code || log.department}
                                                            </Typography>
                                                            <Typography variant="caption" color="textSecondary">Period: {log.month}/{log.year}</Typography>
                                                        </Box>
                                                    </Box>
                                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                                                        <PersonIcon color="action" fontSize="small" />
                                                        <Typography variant="body2">{log.owner_name || log.created_by}</Typography>
                                                    </Box>
                                                </Stack>
                                                
                                                {(log.status === 'REPORTED' || log.status === 'REJECTED') && (
                                                    <Button 
                                                        fullWidth 
                                                        variant="outlined" 
                                                        color="error"
                                                        onClick={() => handleViewDetails(log)}
                                                        startIcon={<VisibilityIcon />}
                                                        sx={{ mt: 2, borderRadius: 3, textTransform: 'none', fontWeight: 'bold' }}
                                                    >
                                                        ดู {log.rejected_count || 0} รายการที่แจ้งแก้ไข
                                                    </Button>
                                                )}
                                            </CardContent>
                                        </Card>
                                    </Fade>
                                </Grid>
                            ))}
                        </Grid>
                    ) : (
                        <Paper sx={{ width: '100%', overflow: 'hidden', borderRadius: 4, boxShadow: '0 10px 30px rgba(0,0,0,0.06)' }}>
                            <TableContainer sx={{ maxHeight: '70vh' }}>
                                <Table stickyHeader>
                                    <TableHead>
                                        <TableRow>
                                            <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold', py: 2.5 }}>Department</TableCell>
                                            <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold' }}>Status</TableCell>
                                            <TableCell align="center" sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold' }}>Issues</TableCell>
                                            <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold' }}>Reporter Details</TableCell>
                                        </TableRow>
                                    </TableHead>
                                    <TableBody>
                                        {logs.map((log) => (
                                            <TableRow key={log.id} hover sx={{ '&:hover': { bgcolor: '#f8f9fa' } }}>
                                                <TableCell sx={{ py: 2 }}>
                                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                                                        <Avatar sx={{ bgcolor: '#e9ecef', color: '#043478' }}>
                                                            <BusinessIcon />
                                                        </Avatar>
                                                        <Box>
                                                            <Typography variant="subtitle2" fontWeight="900">
                                                                {log.department_code || log.department}
                                                            </Typography>
                                                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                                                                <CalendarMonthIcon sx={{ fontSize: 12, color: 'text.secondary' }} />
                                                                <Typography variant="caption" color="textSecondary">
                                                                    Period: {log.month}/{log.year}
                                                                </Typography>
                                                            </Box>
                                                        </Box>
                                                    </Box>
                                                </TableCell>
                                                <TableCell>{renderStatusChip(log.status)}</TableCell>
                                                <TableCell align="center">
                                                    {(log.status === 'REPORTED' || log.status === 'REJECTED') ? (
                                                        <Tooltip title="คลิกเพื่อดูรายละเอียด">
                                                            <Button 
                                                                size="small" 
                                                                variant="contained" 
                                                                color="error"
                                                                onClick={() => handleViewDetails(log)}
                                                                startIcon={<VisibilityIcon />}
                                                                sx={{ 
                                                                    borderRadius: 5, 
                                                                    textTransform: 'none', 
                                                                    px: 2,
                                                                    boxShadow: '0 4px 12px rgba(211, 47, 47, 0.2)'
                                                                }}
                                                            >
                                                                {log.rejected_count || 0} รายการ
                                                            </Button>
                                                        </Tooltip>
                                                    ) : (
                                                        <Typography variant="body2" color="textSecondary">-</Typography>
                                                    )}
                                                </TableCell>
                                                <TableCell>
                                                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
                                                        <PersonIcon sx={{ color: 'text.secondary', fontSize: 20 }} />
                                                        <Box>
                                                            <Typography variant="subtitle2" fontWeight="bold">{log.owner_name || log.created_by}</Typography>
                                                            <Typography variant="caption" display="block" color="textSecondary">
                                                                {new Date(log.created_at).toLocaleString('th-TH', { 
                                                                    month: 'short', 
                                                                    day: 'numeric',
                                                                    hour: '2-digit',
                                                                    minute: '2-digit'
                                                                })}
                                                            </Typography>
                                                        </Box>
                                                    </Box>
                                                </TableCell>
                                            </TableRow>
                                        ))}
                                    </TableBody>
                                </Table>
                            </TableContainer>
                        </Paper>
                    )}
                </>
            )}

            <AuditDetailsModal 
                open={detailModalOpen}
                onClose={() => setDetailModalOpen(false)}
                log={selectedLog}
            />
        </Container>
    );
};

export default LogPage;
