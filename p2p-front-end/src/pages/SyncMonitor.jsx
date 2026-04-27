import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box, Typography, Paper, Grid, Chip, Button, TextField, MenuItem,
  Table, TableBody, TableCell, TableContainer, TableHead, TableRow,
  CircularProgress, Alert, Divider, IconButton, Tooltip,
  FormControl, InputLabel, Select,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import HourglassEmptyIcon from '@mui/icons-material/HourglassEmpty';
import WarningIcon from '@mui/icons-material/Warning';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import CloseIcon from '@mui/icons-material/Close';
import { useAuth } from '../hooks/useAuth';
import api from '../utils/api/axiosInstance';
import { toast } from 'react-toastify';

const STATUS_COLORS = {
  SUCCESS: 'success',
  FAILED: 'error',
  RUNNING: 'info',
  PARTIAL: 'warning',
};

const STATUS_ICONS = {
  SUCCESS: <CheckCircleIcon fontSize="small" />,
  FAILED: <ErrorIcon fontSize="small" />,
  RUNNING: <HourglassEmptyIcon fontSize="small" />,
  PARTIAL: <WarningIcon fontSize="small" />,
};

const JOB_TYPES = [
  { value: 'DW_SYNC', label: 'DW Sync (raw data ← Data Warehouse)' },
  { value: 'TIER1_FAST', label: 'Tier 1 Fast (เดือนปัจจุบันเท่านั้น)' },
  { value: 'TIER2_FULL', label: 'Tier 2 Full (ทั้งปี)' },
  { value: 'ACTUAL_FACT', label: 'Actual Fact Sync (เลือกปี/เดือนเอง)' },
  { value: 'MANUAL', label: 'Manual' },
];

const ALL_MONTHS = ['JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN', 'JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC'];

const fmtDuration = (ms) => {
  if (!ms || ms <= 0) return '-';
  const s = ms / 1000;
  if (s < 60) return `${s.toFixed(1)}s`;
  const m = Math.floor(s / 60);
  const rs = (s % 60).toFixed(0);
  if (m < 60) return `${m}m ${rs}s`;
  const h = Math.floor(m / 60);
  const rm = m % 60;
  return `${h}h ${rm}m`;
};

const fmtDateTime = (iso) => {
  if (!iso) return '-';
  try {
    return new Date(iso).toLocaleString('th-TH', { hour12: false });
  } catch {
    return iso;
  }
};

const SyncMonitor = () => {
  const { user } = useAuth();
  const navigate = useNavigate();

  // Strict check: only username "admin" can access
  const isAuthorized = user?.username === 'admin';

  useEffect(() => {
    if (user && !isAuthorized) {
      toast.error('คุณไม่มีสิทธิ์เข้าถึงหน้านี้');
      navigate('/home', { replace: true });
    }
  }, [user, isAuthorized, navigate]);

  // ──────────────────── Status ────────────────────
  const [status, setStatus] = useState({});
  const [statusLoading, setStatusLoading] = useState(false);

  const fetchStatus = useCallback(async () => {
    setStatusLoading(true);
    try {
      const res = await api.get('/admin/sync/status');
      setStatus(res.data?.latest_by_type || {});
    } catch (err) {
      console.error(err);
      toast.error('ดึงสถานะไม่สำเร็จ');
    } finally {
      setStatusLoading(false);
    }
  }, []);

  // ──────────────────── History ────────────────────
  const [history, setHistory] = useState([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [filterJobType, setFilterJobType] = useState('');
  const [filterStatus, setFilterStatus] = useState('');

  const fetchHistory = useCallback(async () => {
    setHistoryLoading(true);
    try {
      const params = new URLSearchParams({ limit: '50' });
      if (filterJobType) params.set('job_type', filterJobType);
      if (filterStatus) params.set('status', filterStatus);
      const res = await api.get(`/admin/sync/history?${params.toString()}`);
      setHistory(res.data?.runs || []);
    } catch (err) {
      console.error(err);
      toast.error('ดึงประวัติไม่สำเร็จ');
    } finally {
      setHistoryLoading(false);
    }
  }, [filterJobType, filterStatus]);

  // ──────────────────── Reconciliation ────────────────────
  const [reconYear, setReconYear] = useState(String(new Date().getFullYear()));
  const [recon, setRecon] = useState(null);
  const [reconLoading, setReconLoading] = useState(false);

  const fetchRecon = async () => {
    setReconLoading(true);
    try {
      const res = await api.get(`/admin/sync/reconciliation?year=${reconYear}`);
      setRecon(res.data);
    } catch (err) {
      console.error(err);
      toast.error('ดึง reconciliation ไม่สำเร็จ');
    } finally {
      setReconLoading(false);
    }
  };

  // ──────────────────── Trigger ────────────────────
  const [trigJobType, setTrigJobType] = useState('ACTUAL_FACT');
  const [trigYear, setTrigYear] = useState(String(new Date().getFullYear()));
  const [trigMonths, setTrigMonths] = useState([]);
  const [trigLoading, setTrigLoading] = useState(false);

  const handleTrigger = async () => {
    if (!trigYear) {
      toast.warning('กรุณาระบุปี');
      return;
    }
    setTrigLoading(true);
    try {
      const res = await api.post('/admin/sync/trigger', {
        job_type: trigJobType,
        year: trigYear,
        months: trigMonths,
      });
      if (res.status === 202) {
        toast.success('ส่งคำสั่ง sync แล้ว — ทำงาน background; เช็คผลที่ Status/History');
        // Refresh status after a moment
        setTimeout(fetchStatus, 1500);
      }
    } catch (err) {
      console.error(err);
      toast.error('Trigger ไม่สำเร็จ: ' + (err.response?.data?.error || err.message));
    } finally {
      setTrigLoading(false);
    }
  };

  // ──────────────────── Queue ────────────────────
  const [queue, setQueue] = useState({ current: null, pending: [] });
  const [queueLoading, setQueueLoading] = useState(false);

  const fetchQueue = useCallback(async () => {
    setQueueLoading(true);
    try {
      const res = await api.get('/admin/sync/queue');
      setQueue({
        current: res.data?.current || null,
        pending: res.data?.pending || [],
      });
    } catch (err) {
      console.error(err);
    } finally {
      setQueueLoading(false);
    }
  }, []);

  const handleCancel = async (id) => {
    if (!window.confirm('ยกเลิก job นี้ในคิว?')) return;
    try {
      await api.post(`/admin/sync/queue/cancel/${id}`);
      toast.success('ยกเลิกสำเร็จ');
      fetchQueue();
    } catch (err) {
      toast.error('ยกเลิกไม่สำเร็จ: ' + (err.response?.data?.error || err.message));
    }
  };

  const handlePromote = async (id) => {
    try {
      await api.post(`/admin/sync/queue/promote/${id}`);
      toast.success('ย้ายขึ้นหัวคิวสำเร็จ');
      fetchQueue();
    } catch (err) {
      toast.error('Promote ไม่สำเร็จ: ' + (err.response?.data?.error || err.message));
    }
  };

  // ──────────────────── Initial load ────────────────────
  useEffect(() => {
    if (!isAuthorized) return;
    fetchStatus();
    fetchHistory();
    fetchQueue();
  }, [isAuthorized, fetchStatus, fetchHistory, fetchQueue]);

  // Auto-refresh status every 30s, queue every 5s
  useEffect(() => {
    if (!isAuthorized) return;
    const statusI = setInterval(fetchStatus, 30000);
    const queueI = setInterval(fetchQueue, 5000);
    return () => { clearInterval(statusI); clearInterval(queueI); };
  }, [isAuthorized, fetchStatus, fetchQueue]);

  if (!isAuthorized) {
    return (
      <Box sx={{ p: 4 }}>
        <Alert severity="error">คุณไม่มีสิทธิ์เข้าถึงหน้านี้</Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3, maxWidth: 1400, mx: 'auto' }}>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" sx={{ fontWeight: 'bold', color: '#043478', flexGrow: 1 }}>
          Sync Monitor
        </Typography>
        <Tooltip title="Refresh ทั้งหมด">
          <IconButton onClick={() => { fetchStatus(); fetchHistory(); }} color="primary">
            <RefreshIcon />
          </IconButton>
        </Tooltip>
      </Box>

      {/* ──────── Section 1: Latest Status ──────── */}
      <Paper elevation={2} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6" sx={{ fontWeight: 'bold', flexGrow: 1 }}>
            สถานะ Sync ล่าสุด (รีเฟรชอัตโนมัติทุก 30 วิ)
          </Typography>
          {statusLoading && <CircularProgress size={20} />}
        </Box>
        <Grid container spacing={2}>
          {JOB_TYPES.filter((j) => j.value !== 'MANUAL' && j.value !== 'ACTUAL_FACT').map((j) => {
            const run = status[j.value];
            return (
              <Grid item xs={12} md={4} key={j.value}>
                <Paper variant="outlined" sx={{ p: 2, borderRadius: 2, height: '100%' }}>
                  <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1 }}>
                    {j.label}
                  </Typography>
                  {run ? (
                    <>
                      <Box sx={{ display: 'flex', alignItems: 'center', mb: 1, gap: 1 }}>
                        <Chip
                          icon={STATUS_ICONS[run.status] || null}
                          label={run.status}
                          color={STATUS_COLORS[run.status] || 'default'}
                          size="small"
                        />
                        {run.retry_count > 0 && (
                          <Chip label={`retry ${run.retry_count}`} size="small" variant="outlined" color="warning" />
                        )}
                      </Box>
                      <Typography variant="caption" display="block" color="textSecondary">
                        เริ่ม: {fmtDateTime(run.started_at)}
                      </Typography>
                      {run.finished_at && (
                        <Typography variant="caption" display="block" color="textSecondary">
                          เสร็จ: {fmtDateTime(run.finished_at)} ({fmtDuration(run.duration_ms)})
                        </Typography>
                      )}
                      {run.year && (
                        <Typography variant="caption" display="block" color="textSecondary">
                          {run.year} {run.month && `/ ${run.month}`}
                        </Typography>
                      )}
                      {run.rows_fetched > 0 && (
                        <Typography variant="caption" display="block" color="textSecondary">
                          rows: {run.rows_fetched.toLocaleString()}
                        </Typography>
                      )}
                      {run.error_message && (
                        <Alert severity="error" sx={{ mt: 1, fontSize: '0.7rem' }}>
                          {run.error_message.slice(0, 200)}
                        </Alert>
                      )}
                    </>
                  ) : (
                    <Typography variant="caption" color="textSecondary">ยังไม่มีประวัติ</Typography>
                  )}
                </Paper>
              </Grid>
            );
          })}
        </Grid>
      </Paper>

      {/* ──────── Section 2: Manual Trigger ──────── */}
      <Paper elevation={2} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
        <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 2 }}>
          Manual Trigger Sync
        </Typography>
        <Grid container spacing={2} alignItems="flex-end">
          <Grid item xs={12} md={4}>
            <FormControl fullWidth size="small">
              <InputLabel>Job Type</InputLabel>
              <Select value={trigJobType} label="Job Type" onChange={(e) => setTrigJobType(e.target.value)}>
                {JOB_TYPES.map((j) => (
                  <MenuItem key={j.value} value={j.value}>{j.label}</MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={6} md={2}>
            <TextField
              label="Year"
              size="small"
              fullWidth
              value={trigYear}
              onChange={(e) => setTrigYear(e.target.value)}
              placeholder="2026"
            />
          </Grid>
          <Grid item xs={6} md={4}>
            <FormControl fullWidth size="small">
              <InputLabel>Months</InputLabel>
              <Select
                multiple
                value={trigMonths}
                label="Months"
                onChange={(e) => setTrigMonths(e.target.value)}
                renderValue={(sel) => sel.length === 0 ? '(ทั้งปี)' : sel.join(', ')}
              >
                {ALL_MONTHS.map((m) => (
                  <MenuItem key={m} value={m}>{m}</MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={2}>
            <Button
              fullWidth
              variant="contained"
              color="primary"
              startIcon={trigLoading ? <CircularProgress size={18} color="inherit" /> : <PlayArrowIcon />}
              onClick={handleTrigger}
              disabled={trigLoading}
              sx={{ height: '40px', textTransform: 'none', fontWeight: 'bold' }}
            >
              {trigLoading ? 'กำลังส่ง...' : 'Trigger Sync'}
            </Button>
          </Grid>
        </Grid>
        <Typography variant="caption" color="textSecondary" sx={{ mt: 1, display: 'block' }}>
          ⓘ Sync จะรัน background — ดูผลที่ Status (ด้านบน) หรือ History (ด้านล่าง). ถ้า months ว่าง = ทั้งปี
        </Typography>
      </Paper>

      {/* ──────── Queue ──────── */}
      <Paper elevation={2} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6" sx={{ fontWeight: 'bold', flexGrow: 1 }}>
            คิวงาน (Queue) — รีเฟรชอัตโนมัติทุก 5 วิ
          </Typography>
          {queueLoading && <CircularProgress size={20} />}
          <Tooltip title="Refresh">
            <IconButton size="small" onClick={fetchQueue}><RefreshIcon /></IconButton>
          </Tooltip>
        </Box>

        {/* Currently running */}
        {queue.current ? (
          <Paper variant="outlined" sx={{ p: 2, mb: 2, borderColor: 'primary.main', borderWidth: 2 }}>
            <Typography variant="subtitle2" sx={{ fontWeight: 600, color: 'primary.main', mb: 1 }}>
              ▶ กำลังทำงาน
            </Typography>
            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <Typography variant="body2"><strong>Job:</strong> {queue.current.job_type}</Typography>
                <Typography variant="body2">
                  <strong>Year/Months:</strong> {queue.current.year || '-'}
                  {queue.current.months?.length > 0 ? ` / ${queue.current.months.join(', ')}` : ' / (ทั้งปี)'}
                </Typography>
                <Typography variant="body2"><strong>By:</strong> {queue.current.triggered_by}</Typography>
                <Typography variant="body2"><strong>เริ่ม:</strong> {fmtDateTime(queue.current.started_at)}</Typography>
              </Grid>
              <Grid item xs={12} md={6}>
                <Typography variant="body2">
                  <strong>ทำงานมาแล้ว:</strong> {fmtDuration(queue.current.elapsed_ms)}
                </Typography>
                {queue.current.avg_total_ms > 0 ? (
                  <>
                    <Typography variant="body2">
                      <strong>เฉลี่ย:</strong> {fmtDuration(queue.current.avg_total_ms)} (จาก 7 runs ล่าสุด)
                    </Typography>
                    <Typography variant="body2" sx={{ color: 'success.main', fontWeight: 600 }}>
                      <strong>คาดว่าจะเสร็จในอีก:</strong> {fmtDuration(queue.current.remaining_ms)}
                    </Typography>
                  </>
                ) : (
                  <Typography variant="caption" color="textSecondary">
                    (ยังไม่มีประวัติเฉลี่ย — ETA ไม่สามารถคำนวณ)
                  </Typography>
                )}
              </Grid>
            </Grid>
          </Paper>
        ) : (
          <Alert severity="info" sx={{ mb: 2 }}>ไม่มี job ที่กำลังทำงาน</Alert>
        )}

        {/* Pending */}
        <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 1 }}>
          รออยู่ในคิว ({queue.pending.length})
        </Typography>
        {queue.pending.length === 0 ? (
          <Typography variant="caption" color="textSecondary">— ไม่มี job รออยู่ —</Typography>
        ) : (
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell sx={{ fontWeight: 'bold', width: 60 }}>ลำดับ</TableCell>
                  <TableCell sx={{ fontWeight: 'bold', width: 80 }}>Priority</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>Job Type</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>Year/Months</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>By</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>เข้าคิว</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>เริ่มในอีก ~</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>คาดใช้เวลา</TableCell>
                  <TableCell sx={{ fontWeight: 'bold', width: 130 }} align="center">การจัดการ</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {queue.pending.map((p) => {
                  const prioColor = p.priority === 1 ? 'error'
                    : p.priority === 2 ? 'warning'
                    : p.priority === 3 ? 'info'
                    : 'default';
                  return (
                  <TableRow key={p.id} hover>
                    <TableCell>{p.position}</TableCell>
                    <TableCell>
                      <Chip label={`P${p.priority}`} size="small" color={prioColor} />
                    </TableCell>
                    <TableCell><code>{p.job_type}</code></TableCell>
                    <TableCell>
                      {p.year || '-'}
                      {p.months?.length > 0 ? ` / ${p.months.join(', ')}` : ' / (ทั้งปี)'}
                    </TableCell>
                    <TableCell sx={{ fontSize: '0.7rem' }}>{p.triggered_by}</TableCell>
                    <TableCell>{fmtDateTime(p.enqueued_at)}</TableCell>
                    <TableCell>{fmtDuration(p.eta_start_ms)}</TableCell>
                    <TableCell>
                      {p.avg_total_ms > 0 ? fmtDuration(p.avg_total_ms) : '-'}
                    </TableCell>
                    <TableCell align="center">
                      <Tooltip title="ย้ายขึ้นหัวคิว">
                        <IconButton size="small" color="primary"
                          onClick={() => handlePromote(p.id)}
                          disabled={p.position === 1}>
                          <ArrowUpwardIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="ยกเลิก">
                        <IconButton size="small" color="error" onClick={() => handleCancel(p.id)}>
                          <CloseIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </Paper>

      {/* ──────── Section 3: Reconciliation ──────── */}
      <Paper elevation={2} sx={{ p: 3, mb: 3, borderRadius: 2 }}>
        <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 2 }}>
          Reconciliation (Health Check)
        </Typography>
        <Box sx={{ display: 'flex', gap: 2, mb: 2, alignItems: 'center' }}>
          <TextField
            label="Year"
            size="small"
            value={reconYear}
            onChange={(e) => setReconYear(e.target.value)}
            sx={{ width: 120 }}
          />
          <Button
            variant="outlined"
            onClick={fetchRecon}
            disabled={reconLoading}
            startIcon={reconLoading ? <CircularProgress size={16} /> : <RefreshIcon />}
          >
            ตรวจสอบ
          </Button>
        </Box>
        {recon && (
          <>
            <Box sx={{ mb: 2 }}>
              <Chip
                label={recon.is_healthy ? '✓ HEALTHY' : '⚠ ISSUES DETECTED'}
                color={recon.is_healthy ? 'success' : 'error'}
                sx={{ fontWeight: 'bold' }}
              />
            </Box>
            <Grid container spacing={2}>
              <Grid item xs={6} md={3}>
                <Paper variant="outlined" sx={{ p: 2, textAlign: 'center' }}>
                  <Typography variant="caption" color="textSecondary">Raw HMW</Typography>
                  <Typography variant="h5">{recon.raw_hmw_count?.toLocaleString() || 0}</Typography>
                </Paper>
              </Grid>
              <Grid item xs={6} md={3}>
                <Paper variant="outlined" sx={{ p: 2, textAlign: 'center' }}>
                  <Typography variant="caption" color="textSecondary">Raw CLIK</Typography>
                  <Typography variant="h5">{recon.raw_clik_count?.toLocaleString() || 0}</Typography>
                </Paper>
              </Grid>
              <Grid item xs={6} md={3}>
                <Paper variant="outlined" sx={{ p: 2, textAlign: 'center' }}>
                  <Typography variant="caption" color="textSecondary">Fact Transactions</Typography>
                  <Typography variant="h5" color="primary">
                    {recon.fact_transaction_count?.toLocaleString() || 0}
                  </Typography>
                </Paper>
              </Grid>
              <Grid item xs={6} md={3}>
                <Paper variant="outlined" sx={{ p: 2, textAlign: 'center' }}>
                  <Typography variant="caption" color="textSecondary">Fact Amounts</Typography>
                  <Typography variant="h5" color="primary">
                    {recon.fact_amount_count?.toLocaleString() || 0}
                  </Typography>
                </Paper>
              </Grid>
            </Grid>
            {recon.warnings?.length > 0 && (
              <Box sx={{ mt: 2 }}>
                {recon.warnings.map((w, i) => (
                  <Alert key={i} severity="warning" sx={{ mb: 1 }}>{w}</Alert>
                ))}
              </Box>
            )}
            {recon.monthly_fact_counts && Object.keys(recon.monthly_fact_counts).length > 0 && (
              <Box sx={{ mt: 2 }}>
                <Typography variant="subtitle2" sx={{ mb: 1 }}>Fact rows per month:</Typography>
                <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                  {ALL_MONTHS.map((m) => {
                    const count = recon.monthly_fact_counts[m] || 0;
                    return (
                      <Chip
                        key={m}
                        label={`${m}: ${count.toLocaleString()}`}
                        size="small"
                        variant={count > 0 ? 'filled' : 'outlined'}
                        color={count > 0 ? 'primary' : 'default'}
                      />
                    );
                  })}
                </Box>
              </Box>
            )}
          </>
        )}
      </Paper>

      {/* ──────── Section 4: History ──────── */}
      <Paper elevation={2} sx={{ p: 3, borderRadius: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, gap: 2, flexWrap: 'wrap' }}>
          <Typography variant="h6" sx={{ fontWeight: 'bold', flexGrow: 1 }}>
            ประวัติ Sync (50 รอบล่าสุด)
          </Typography>
          <FormControl size="small" sx={{ minWidth: 160 }}>
            <InputLabel>Job Type</InputLabel>
            <Select value={filterJobType} label="Job Type" onChange={(e) => setFilterJobType(e.target.value)}>
              <MenuItem value="">ทั้งหมด</MenuItem>
              {JOB_TYPES.map((j) => (
                <MenuItem key={j.value} value={j.value}>{j.value}</MenuItem>
              ))}
            </Select>
          </FormControl>
          <FormControl size="small" sx={{ minWidth: 140 }}>
            <InputLabel>Status</InputLabel>
            <Select value={filterStatus} label="Status" onChange={(e) => setFilterStatus(e.target.value)}>
              <MenuItem value="">ทั้งหมด</MenuItem>
              <MenuItem value="SUCCESS">SUCCESS</MenuItem>
              <MenuItem value="FAILED">FAILED</MenuItem>
              <MenuItem value="RUNNING">RUNNING</MenuItem>
              <MenuItem value="PARTIAL">PARTIAL</MenuItem>
            </Select>
          </FormControl>
          <Button
            variant="outlined"
            size="small"
            onClick={fetchHistory}
            disabled={historyLoading}
            startIcon={historyLoading ? <CircularProgress size={14} /> : <RefreshIcon />}
          >
            กรอง
          </Button>
        </Box>
        <Divider sx={{ mb: 2 }} />
        <TableContainer sx={{ maxHeight: 600 }}>
          <Table stickyHeader size="small">
            <TableHead>
              <TableRow>
                <TableCell sx={{ fontWeight: 'bold' }}>เริ่ม</TableCell>
                <TableCell sx={{ fontWeight: 'bold' }}>Job Type</TableCell>
                <TableCell sx={{ fontWeight: 'bold' }}>Year/Month</TableCell>
                <TableCell sx={{ fontWeight: 'bold' }}>Status</TableCell>
                <TableCell sx={{ fontWeight: 'bold' }} align="right">Duration</TableCell>
                <TableCell sx={{ fontWeight: 'bold' }} align="right">Rows</TableCell>
                <TableCell sx={{ fontWeight: 'bold' }} align="center">Retry</TableCell>
                <TableCell sx={{ fontWeight: 'bold' }}>Triggered By</TableCell>
                <TableCell sx={{ fontWeight: 'bold' }}>Error</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {history.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={9} align="center" sx={{ py: 3, color: 'text.secondary' }}>
                    ไม่มีประวัติ
                  </TableCell>
                </TableRow>
              ) : (
                history.map((r) => (
                  <TableRow key={r.id} hover>
                    <TableCell>{fmtDateTime(r.started_at)}</TableCell>
                    <TableCell><code>{r.job_type}</code></TableCell>
                    <TableCell>{r.year}{r.month && ` / ${r.month}`}</TableCell>
                    <TableCell>
                      <Chip
                        label={r.status}
                        color={STATUS_COLORS[r.status] || 'default'}
                        size="small"
                      />
                    </TableCell>
                    <TableCell align="right">{fmtDuration(r.duration_ms)}</TableCell>
                    <TableCell align="right">{(r.rows_fetched || 0).toLocaleString()}</TableCell>
                    <TableCell align="center">{r.retry_count || 0}</TableCell>
                    <TableCell sx={{ fontSize: '0.7rem' }}>{r.triggered_by}</TableCell>
                    <TableCell sx={{ fontSize: '0.7rem', maxWidth: 250, color: 'error.main' }}>
                      {r.error_message ? r.error_message.slice(0, 100) : '-'}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>
    </Box>
  );
};

export default SyncMonitor;
