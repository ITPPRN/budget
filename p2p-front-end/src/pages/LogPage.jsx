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
    TableContainer
} from '@mui/material';
import HistoryIcon from '@mui/icons-material/History';
import VisibilityIcon from '@mui/icons-material/Visibility';
import RefreshIcon from '@mui/icons-material/Refresh';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import api from '../utils/api/axiosInstance';
import AuditDetailsModal from '../components/Budget/AuditDetailsModal';

const LogPage = () => {
    const [logs, setLogs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedLog, setSelectedLog] = useState(null);
    const [detailModalOpen, setDetailModalOpen] = useState(false);

    // Filters
    const [filters, setFilters] = useState({
        department: '',
        year: '',
        month: ''
    });

    const fetchLogs = async () => {
        setLoading(true);
        try {
            const queryParams = new URLSearchParams(
                Object.entries(filters).filter(([_, v]) => v !== '')
            ).toString();
            const res = await api.get(`/budgets/audit/logs?${queryParams}`);
            setLogs(res.data || []);
        } catch (err) {
            console.error("Fetch Logs Error:", err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchLogs();
    }, []);

    const handleViewDetails = (log) => {
        setSelectedLog(log);
        setDetailModalOpen(true);
    };

    const getStatusChip = (status) => {
        switch (status?.toUpperCase()) {
            case 'APPROVED':
            case 'CONFIRMED':
                return <Chip icon={<CheckCircleIcon />} label="APPROVED" color="success" size="small" variant="outlined" />;
            case 'REPORTED':
            case 'REJECTED':
                return <Chip icon={<ErrorIcon />} label="REPORTED" color="error" size="small" variant="outlined" />;
            default:
                return <Chip label={status} size="small" />;
        }
    };

    return (
        <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
            <Box sx={{ display: 'flex', flexDirection: { xs: 'column', sm: 'row' }, alignItems: { xs: 'flex-start', sm: 'center' }, justifyContent: 'space-between', mb: 3, gap: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                    <HistoryIcon sx={{ fontSize: 40, color: '#043478' }} />
                    <Typography variant="h4" fontWeight="bold" sx={{ color: '#043478' }}>
                        Audit Logs
                    </Typography>
                </Box>
                <Button 
                    startIcon={<RefreshIcon />} 
                    variant="contained" 
                    onClick={fetchLogs}
                    disabled={loading}
                    sx={{ bgcolor: '#043478', borderRadius: 2 }}
                >
                    Refresh Data
                </Button>
            </Box>

            {/* Filter Pane */}
            <Paper sx={{ p: 3, mb: 3, borderRadius: 2, boxShadow: '0 2px 10px rgba(0,0,0,0.05)' }}>
                <Grid container spacing={2} alignItems="center">
                    <Grid item xs={12} sm={4}>
                        <TextField
                            label="Department"
                            fullWidth
                            size="small"
                            placeholder="ค้นหาแผนก..."
                            value={filters.department}
                            onChange={(e) => setFilters({...filters, department: e.target.value})}
                        />
                    </Grid>
                    <Grid item xs={12} sm={3}>
                        <TextField
                            label="Year"
                            fullWidth
                            size="small"
                            placeholder="YYYY"
                            value={filters.year}
                            onChange={(e) => setFilters({...filters, year: e.target.value})}
                        />
                    </Grid>
                    <Grid item xs={12} sm={3}>
                        <TextField
                            select
                            label="Month"
                            fullWidth
                            size="small"
                            value={filters.month}
                            onChange={(e) => setFilters({...filters, month: e.target.value})}
                        >
                            <MenuItem value="">All Months</MenuItem>
                            {["01","02","03","04","05","06","07","08","09","10","11","12"].map(m => (
                                <MenuItem key={m} value={m}>{m}</MenuItem>
                            ))}
                        </TextField>
                    </Grid>
                    <Grid item xs={12} sm={2}>
                        <Button variant="contained" fullWidth onClick={fetchLogs} sx={{ height: 40, borderRadius: 2 }}>Filter</Button>
                    </Grid>
                </Grid>
            </Paper>

            <Paper sx={{ width: '100%', overflow: 'hidden', borderRadius: 2, boxShadow: '0 4px 20px rgba(0,0,0,0.08)' }}>
                <TableContainer sx={{ maxHeight: '70vh' }}>
                    <Table stickyHeader size="medium">
                        <TableHead>
                            <TableRow>
                                <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold' }}>Department</TableCell>
                                <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold' }}>Status</TableCell>
                                <TableCell align="center" sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold' }}>Rejected Count</TableCell>
                                <TableCell sx={{ bgcolor: '#043478', color: 'white', fontWeight: 'bold' }}>Created Details</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={4} align="center" sx={{ py: 10 }}>
                                        <CircularProgress />
                                    </TableCell>
                                </TableRow>
                            ) : logs.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={4} align="center" sx={{ py: 10 }}>
                                        <Typography color="textSecondary">No audit logs found</Typography>
                                    </TableCell>
                                </TableRow>
                            ) : (
                                logs.map((log) => (
                                    <TableRow key={log.id} hover>
                                        <TableCell>
                                            <Typography variant="body2" fontWeight="bold">
                                                {log.department_code || log.department}
                                            </Typography>
                                            <Typography variant="caption" color="textSecondary">
                                                Period: {log.month}/{log.year}
                                            </Typography>
                                        </TableCell>
                                        <TableCell>{getStatusChip(log.status)}</TableCell>
                                        <TableCell align="center">
                                            {log.status === 'REPORTED' || log.status === 'REJECTED' ? (
                                                <Button 
                                                    size="small" 
                                                    variant="outlined" 
                                                    color="error"
                                                    onClick={() => handleViewDetails(log)}
                                                    startIcon={<VisibilityIcon />}
                                                    sx={{ borderRadius: 5, textTransform: 'none' }}
                                                >
                                                    {log.rejected_count || 0} รายการ
                                                </Button>
                                            ) : (
                                                "-"
                                            )}
                                        </TableCell>
                                        <TableCell>
                                            <Typography variant="body2" fontWeight="medium">{log.owner_name || log.created_by}</Typography>
                                            <Typography variant="caption" display="block" color="textSecondary">
                                                {new Date(log.created_at).toLocaleString('th-TH', { 
                                                    year: 'numeric', 
                                                    month: 'long', 
                                                    day: 'numeric',
                                                    hour: '2-digit',
                                                    minute: '2-digit'
                                                })}
                                            </Typography>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </TableContainer>
            </Paper>

            <AuditDetailsModal 
                open={detailModalOpen}
                onClose={() => setDetailModalOpen(false)}
                log={selectedLog}
            />
        </Container>
    );
};

export default LogPage;
