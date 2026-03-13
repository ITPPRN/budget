import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import {
  Box, Grid, Paper, Typography, Button, Select,
  MenuItem, Switch, Link, IconButton, Dialog,
  DialogTitle, DialogContent, TextField, InputAdornment,
  List, ListItem, ListItemText, ListItemAvatar, Avatar, Divider,
  CircularProgress
} from '@mui/material';
import SyncIcon from '@mui/icons-material/Sync';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import SearchIcon from '@mui/icons-material/Search';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import SaveIcon from '@mui/icons-material/Save';
import CancelIcon from '@mui/icons-material/Cancel';
import CloseIcon from '@mui/icons-material/Close';
import DescriptionIcon from '@mui/icons-material/Description';
import InsertDriveFileIcon from '@mui/icons-material/InsertDriveFile';
import { useAuth } from '../hooks/useAuth';
import api from '../utils/api/axiosInstance';
import { toast } from 'react-toastify';
import StorageIcon from '@mui/icons-material/Storage';
import GLMappingModal from '../components/Budget/GLMappingModal';
import BudgetStructureModal from '../components/Budget/BudgetStructureModal';
import { BudgetVersionModal, CapexPlanModal, CapexActualModal } from '../components/Budget/VersionModals';


const DataManagePage = () => {
  const { user } = useAuth();

  // --- State for Versions ---
  const [budgetVersions, setBudgetVersions] = useState([]);
  const [capexBgVersions, setCapexBgVersions] = useState([]);
  const [capexActualVersions, setCapexActualVersions] = useState([]);

  const [selectedBudget, setSelectedBudget] = useState('');
  const [selectedCapexBg, setSelectedCapexBg] = useState('');
  const [selectedCapexActual, setSelectedCapexActual] = useState('');
  const [actualYear, setActualYear] = useState('');
  const [selectedMonths, setSelectedMonths] = useState([]);
  const [actualYears, setActualYears] = useState([]);

  const [open, setOpen] = useState(false);
  const [modalType, setModalType] = useState(''); // 'BUDGET', 'CAPEX_BG', 'CAPEX_ACTUAL'
  const [modalTitle, setModalTitle] = useState('');

  // --- Fetch Data on Load ---
  const fetchAllVersions = async () => {
    try {
      const [resBudget, resCapexBg, resCapexAct, resYears, resConfigs] = await Promise.all([
        api.get('/budgets/files-budget'),
        api.get('/budgets/files-capex-budget'),
        api.get('/budgets/files-capex-actual'),
        api.get('/budgets/actual-years'),
        api.get('/budgets/configs'),
      ]);

      setBudgetVersions(resBudget.data || []);
      setCapexBgVersions(resCapexBg.data || []);
      setCapexActualVersions(resCapexAct.data || []);
      setActualYears(resYears.data || []);

      // Apply System Configs
      const configs = resConfigs.data || {};
      if (configs.selectedBudget) setSelectedBudget(configs.selectedBudget);
      if (configs.selectedCapexBg) setSelectedCapexBg(configs.selectedCapexBg);
      if (configs.selectedCapexActual) setSelectedCapexActual(configs.selectedCapexActual);
      if (configs.actualYear) setActualYear(configs.actualYear);
      if (configs.selectedMonths) {
        try {
          setSelectedMonths(JSON.parse(configs.selectedMonths));
        } catch (e) {
          setSelectedMonths([]);
        }
      }
    } catch (error) {
      console.error("Failed to fetch versions or configs:", error);
      toast.error("ไม่สามารถดึงข้อมูล Version หรือการตั้งค่าได้");
    }
  };

  useEffect(() => {
    fetchAllVersions();
  }, []);

  // --- Handlers ---
  const handleOpenModal = (type, title) => {
    setModalType(type);
    setModalTitle(title);
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
  };


  const saveRemoteConfig = async (key, value) => {
    try {
      // If value is array or object, stringify it
      const valStr = typeof value === 'string' ? value : JSON.stringify(value);
      await api.post(`/budgets/configs/${key}`, { value: valStr });
    } catch (error) {
      console.error(`Failed to save config ${key}:`, error);
    }
  };

  const [loadingUpdate, setLoadingUpdate] = useState({ budget: false, capexBg: false, capexActual: false, dbActuals: false });

  const handleUpdateBudget = async () => {
    if (!selectedBudget) {
      if (window.confirm("คุณต้องการยกเลิกการเลือกไฟล์ Budget หรือไม่? (ข้อมูลจะถูกล้างออกจากระบบด้วย)")) {
        try {
          await api.post('/budgets/clear-budget');
          await saveRemoteConfig('selectedBudget', '');
          setSelectedBudget('');
          toast.success("ยกเลิกการเลือกและล้างข้อมูล Budget สำเร็จ");
        } catch (error) {
          toast.error("ไม่สามารถล้างข้อมูลได้: " + (error.response?.data?.error || error.message));
        }
      }
      return;
    }
    setLoadingUpdate(prev => ({ ...prev, budget: true }));
    try {
      await api.post(`/budgets/files-budget/${selectedBudget}/sync`);
      await saveRemoteConfig('selectedBudget', selectedBudget);
      toast.success("อัปเดต Budget สำเร็จ");
    } catch (error) {
      toast.error("อัปเดต Budget ล้มเหลว: " + (error.response?.data?.error || error.message));
    } finally {
      setLoadingUpdate(prev => ({ ...prev, budget: false }));
    }
  };

  const handleUpdateCapexBg = async () => {
    if (!selectedCapexBg) {
      if (window.confirm("คุณต้องการยกเลิกการเลือกไฟล์ CAPEX Plan หรือไม่? (ข้อมูลจะถูกล้างออกจากระบบด้วย)")) {
        try {
          await api.post('/budgets/clear-capex-budget');
          await saveRemoteConfig('selectedCapexBg', '');
          setSelectedCapexBg('');
          toast.success("ยกเลิกการเลือกและล้างข้อมูล CAPEX Plan สำเร็จ");
        } catch (error) {
          toast.error("ไม่สามารถล้างข้อมูลได้: " + (error.response?.data?.error || error.message));
        }
      }
      return;
    }
    setLoadingUpdate(prev => ({ ...prev, capexBg: true }));
    try {
      await api.post(`/budgets/files-capex-budget/${selectedCapexBg}/sync`);
      await saveRemoteConfig('selectedCapexBg', selectedCapexBg);
      toast.success("อัปเดต CAPEX Plan สำเร็จ");
    } catch (error) {
      toast.error("อัปเดต CAPEX Plan ล้มเหลว: " + (error.response?.data?.error || error.message));
    } finally {
      setLoadingUpdate(prev => ({ ...prev, capexBg: false }));
    }
  };

  const handleUpdateCapexActual = async () => {
    if (!selectedCapexActual) {
      if (window.confirm("คุณต้องการยกเลิกการเลือกไฟล์ CAPEX Actual หรือไม่? (ข้อมูลจะถูกล้างออกจากระบบด้วย)")) {
        try {
          await api.post('/budgets/clear-capex-actual');
          await saveRemoteConfig('selectedCapexActual', '');
          setSelectedCapexActual('');
          toast.success("ยกเลิกการเลือกและล้างข้อมูล CAPEX Actual สำเร็จ");
        } catch (error) {
          toast.error("ไม่สามารถล้างข้อมูลได้: " + (error.response?.data?.error || error.message));
        }
      }
      return;
    }
    setLoadingUpdate(prev => ({ ...prev, capexActual: true }));
    try {
      await api.post(`/budgets/files-capex-actual/${selectedCapexActual}/sync`);
      await saveRemoteConfig('selectedCapexActual', selectedCapexActual);
      toast.success("อัปเดต CAPEX Actual สำเร็จ");
    } catch (error) {
      toast.error("อัปเดต CAPEX Actual ล้มเหลว: " + (error.response?.data?.error || error.message));
    } finally {
      setLoadingUpdate(prev => ({ ...prev, capexActual: false }));
    }
  };

  const MONTHS = ['JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN', 'JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC'];

  const toggleMonth = (month) => {
    const monthIdx = MONTHS.indexOf(month);
    if (selectedMonths.length > 0 && selectedMonths[selectedMonths.length - 1] === month) {
      setSelectedMonths([]);
    } else {
      const newSelected = MONTHS.slice(0, monthIdx + 1);
      setSelectedMonths(newSelected);
    }
  };

  const handleUpdateDatabaseActuals = async () => {
    if (!actualYear) {
      toast.warning("กรุณาเลือกปีงบประมาณ");
      return;
    }
    setLoadingUpdate(prev => ({ ...prev, dbActuals: true }));
    try {
      // Phase 31 Update: 
      // We no longer trigger a heavy Manual Sync on every click.
      // Instead, we just Save the Filter Config. 
      // The Dashboard and Reports will use these filters to query the "Central Table" (Table กลาง)
      // which is kept up-to-date by the Backend Cron Job (every 5 mins).

      await saveRemoteConfig('actualYear', actualYear);
      await saveRemoteConfig('selectedMonths', selectedMonths);

      // --- Sync to LocalStorage for immediate cross-page reactivity ---
      const syncConfig = {
        actualYear: actualYear,
        selectedMonths: selectedMonths,
        selectedBudget: selectedBudget,
        selectedCapexBg: selectedCapexBg,
        selectedCapexActual: selectedCapexActual,
        timestamp: new Date().toISOString()
      };
      localStorage.setItem('dm_lastSyncedConfig', JSON.stringify(syncConfig));

      toast.success("บันทึกการตั้งค่าตัวกรอง Actuals สำเร็จ");
    } catch (error) {
      console.error("Save Config Error:", error);
      toast.error("บันทึกการตั้งค่าล้มเหลว");
    } finally {
      // Small artificial delay to show state change
      setTimeout(() => {
        setLoadingUpdate(prev => ({ ...prev, dbActuals: false }));
      }, 300);
    }
  };

  return (
    <Box sx={{ p: 3, bgcolor: '#f5f7fb', minHeight: '100vh' }}>

      {/* 1. Header */}
      <Box sx={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        mb: 4,
        p: 3,
        bgcolor: 'white',
        borderRadius: '15px',
        boxShadow: '0 4px 20px rgba(0,0,0,0.05)'
      }}>
        <Box>
          <Typography variant="h4" sx={{ fontWeight: 'bold', color: '#1a237e' }}>
            จัดการข้อมูล
          </Typography>
        </Box>
      </Box>

      {/* 2. Main Logic Area - Using CSS Grid for absolute 2x2 stability */}
      <Box sx={{ 
        display: 'grid', 
        gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr' }, 
        gap: 3,
        alignItems: 'stretch'
      }}>
        
        {/* Card 1: Budget & CAPEX Versions */}
        <Box>
          <Paper
            elevation={0}
            sx={{
              p: 4,
              borderRadius: '20px',
              border: '1px solid #e3e6f0',
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              bgcolor: 'white',
              boxShadow: '0 10px 30px rgba(0,0,0,0.03)',
              position: 'relative',
              overflow: 'hidden'
            }}
          >
            {/* Background Decorative Element */}
            <Box sx={{
              position: 'absolute',
              bottom: -30,
              left: -30,
              width: 120,
              height: 120,
              borderRadius: '50%',
              background: 'rgba(26, 35, 126, 0.02)',
              zIndex: 0
            }} />

            <Box sx={{ position: 'relative', zIndex: 1 }}>
              <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 4, display: 'flex', alignItems: 'center', gap: 1.5, color: '#1a237e' }}>
                <Avatar sx={{ bgcolor: '#e8eaf6', width: 32, height: 32 }}>
                  <DescriptionIcon sx={{ color: '#1a237e', fontSize: 20 }} />
                </Avatar>
                Budget
              </Typography>

              {/* Budget Version */}
              <Box sx={{ mb: 5 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1.5, flexWrap: 'wrap', gap: 1 }}>
                  <Typography variant="subtitle2" sx={{ fontWeight: 'bold', color: 'text.secondary', fontSize: '0.75rem', letterSpacing: '0.05rem' }}>
                    BUDGET VERSION
                  </Typography>
                  <Button
                    variant="text"
                    size="small"
                    onClick={() => handleOpenModal('BUDGET', 'MANAGE BUDGET')}
                    sx={{ textTransform: 'none', fontSize: '0.75rem', fontWeight: 'bold' }}
                    startIcon={<EditIcon sx={{ fontSize: 14 }} />}
                  >
                    จัดการ Version
                  </Button>
                </Box>
                <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                  <Select
                    size="small"
                    value={selectedBudget}
                    onChange={(e) => setSelectedBudget(e.target.value)}
                    displayEmpty
                    fullWidth
                    sx={{
                      borderRadius: '12px',
                      bgcolor: '#fcfcfc',
                      flex: 1,
                      minWidth: { xs: '100%', md: '200px' },
                      '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e0e0e0' },
                      '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: '#1a237e' }
                    }}
                  >
                    <MenuItem value=""><em>-- เลือกไฟล์งบประมาณ --</em></MenuItem>
                    {budgetVersions.map((v) => (
                      <MenuItem key={v.id} value={v.id}>{v.file_name}</MenuItem>
                    ))}
                  </Select>
                  <Button
                    variant="contained"
                    onClick={handleUpdateBudget}
                    disabled={loadingUpdate.budget}
                    sx={{ borderRadius: '12px', minWidth: '100px', fontWeight: 'bold' }}
                  >
                    {loadingUpdate.budget ? <CircularProgress size={20} color="inherit" /> : 'อัปเดต'}
                  </Button>
                </Box>
              </Box>

              <Divider sx={{ my: 4, borderStyle: 'dashed' }} />

              <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 3, display: 'flex', alignItems: 'center', gap: 1.5, color: '#1a237e' }}>
                <Avatar sx={{ bgcolor: '#e8eaf6', width: 32, height: 32 }}>
                  <DescriptionIcon sx={{ color: '#1a237e', fontSize: 20 }} />
                </Avatar>
                CAPEX
              </Typography>

              <Box sx={{ mb: 3, pl: 0 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1, flexWrap: 'wrap', gap: 1 }}>
                  <Typography variant="caption" sx={{ fontWeight: 'bold', color: 'text.secondary' }}>CAPEX PLAN (แผนงบประมาณ)</Typography>
                  <Button
                    variant="text"
                    size="small"
                    onClick={() => handleOpenModal('CAPEX_BG', 'MANAGE CAPEX PLAN')}
                    sx={{ textTransform: 'none', fontSize: '0.75rem', fontWeight: 'bold' }}
                    startIcon={<EditIcon sx={{ fontSize: 14 }} />}
                  >
                    จัดการ Version
                  </Button>
                </Box>
                <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                  <Select
                    size="small"
                    value={selectedCapexBg}
                    onChange={(e) => setSelectedCapexBg(e.target.value)}
                    displayEmpty
                    fullWidth
                    sx={{
                      borderRadius: '12px',
                      bgcolor: '#fcfcfc',
                      flex: 1,
                      minWidth: { xs: '100%', md: '200px' },
                      '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e0e0e0' }
                    }}
                  >
                    <MenuItem value=""><em>-- เลือก Capex Plan --</em></MenuItem>
                    {capexBgVersions.map((v) => (
                      <MenuItem key={v.id} value={v.id}>{v.file_name}</MenuItem>
                    ))}
                  </Select>
                  <Button
                    variant="contained"
                    onClick={handleUpdateCapexBg}
                    disabled={loadingUpdate.capexBg}
                    sx={{ borderRadius: '12px', minWidth: '100px', fontWeight: 'bold' }}
                  >
                    {loadingUpdate.capexBg ? <CircularProgress size={20} color="inherit" /> : 'อัปเดต'}
                  </Button>
                </Box>
              </Box>

              <Box sx={{ pl: 0, mb: 2 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1, flexWrap: 'wrap', gap: 1 }}>
                  <Typography variant="caption" sx={{ fontWeight: 'bold', color: 'text.secondary' }}>CAPEX ACTUAL (ข้อมูลจริง)</Typography>
                  <Button
                    variant="text"
                    size="small"
                    onClick={() => handleOpenModal('CAPEX_ACTUAL', 'MANAGE CAPEX ACTUAL')}
                    sx={{ textTransform: 'none', fontSize: '0.75rem', fontWeight: 'bold' }}
                    startIcon={<EditIcon sx={{ fontSize: 14 }} />}
                  >
                    จัดการ Version
                  </Button>
                </Box>
                <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                  <Select
                    size="small"
                    value={selectedCapexActual}
                    onChange={(e) => setSelectedCapexActual(e.target.value)}
                    displayEmpty
                    fullWidth
                    sx={{
                      borderRadius: '12px',
                      bgcolor: '#fcfcfc',
                      flex: 1,
                      minWidth: { xs: '100%', md: '200px' },
                      '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e0e0e0' }
                    }}
                  >
                    <MenuItem value=""><em>-- เลือก Capex Actual --</em></MenuItem>
                    {capexActualVersions.map((v) => (
                      <MenuItem key={v.id} value={v.id}>{v.file_name}</MenuItem>
                    ))}
                  </Select>
                  <Button
                    variant="contained"
                    onClick={handleUpdateCapexActual}
                    disabled={loadingUpdate.capexActual}
                    sx={{ borderRadius: '12px', minWidth: '100px', fontWeight: 'bold' }}
                  >
                    {loadingUpdate.capexActual ? <CircularProgress size={20} color="inherit" /> : 'อัปเดต'}
                  </Button>
                </Box>
              </Box>
            </Box>
          </Paper>
        </Box>

        {/* Card 2: Database Actuals Selection */}
        <Box>
          <Paper elevation={0}
            sx={{
              p: 4,
              borderRadius: '20px',
              border: '1px solid #e3e6f0',
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              position: 'relative',
              overflow: 'hidden',
              bgcolor: 'white',
              boxShadow: '0 10px 30px rgba(0,0,0,0.03)'
            }}>
            <Box sx={{
              position: 'absolute',
              top: -50,
              right: -50,
              width: 150,
              height: 150,
              borderRadius: '50%',
              background: 'rgba(26, 35, 126, 0.03)',
              zIndex: 0
            }} />

            <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 4, display: 'flex', alignItems: 'center', gap: 1.5, color: '#1a237e', position: 'relative', zIndex: 1 }}>
              <Avatar sx={{ bgcolor: '#e8eaf6', width: 32, height: 32 }}>
                <StorageIcon sx={{ color: '#1a237e', fontSize: 20 }} />
              </Avatar>
              Database Actuals
            </Typography>

            <Box sx={{ position: 'relative', zIndex: 1 }}>
              <Typography variant="subtitle2" sx={{ fontWeight: 'bold', color: 'text.secondary', mb: 1.5, fontSize: '0.75rem', letterSpacing: '0.05rem' }}>
                SELECT FISCAL YEAR
              </Typography>
              <Select
                size="small"
                value={actualYear}
                onChange={(e) => setActualYear(e.target.value)}
                fullWidth
                displayEmpty
                sx={{
                  borderRadius: '12px',
                  mb: 4,
                  bgcolor: 'white',
                  '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e0e0e0' },
                  '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: '#1a237e' }
                }}
              >
                <MenuItem value=""><em>-- เลือกปี --</em></MenuItem>
                {actualYears.map(year => (
                  <MenuItem key={year} value={year}>{year}</MenuItem>
                ))}
              </Select>

              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'end', mb: 2 }}>
                <Typography variant="subtitle2" sx={{ fontWeight: 'bold', color: 'text.secondary', fontSize: '0.75rem', letterSpacing: '0.05rem' }}>
                  SELECT FISCAL MONTH
                </Typography>
                {selectedMonths.length > 0 && (
                  <Button
                    size="small"
                    onClick={() => setSelectedMonths([])}
                    color="error"
                    sx={{ textTransform: 'none', fontSize: '0.75rem', p: 0, minWidth: 'auto' }}
                  >
                    Clear Selection
                  </Button>
                )}
              </Box>

              {/* Robust Grid for Months */}
              <Box sx={{ 
                display: 'grid', 
                gridTemplateColumns: { xs: 'repeat(2, 1fr)', sm: 'repeat(3, 1fr)', md: 'repeat(4, 1fr)' }, 
                gap: 1 
              }}>
                {MONTHS.map((month) => {
                  const isSelected = selectedMonths.includes(month);
                  return (
                    <Button
                      key={month}
                      fullWidth
                      variant={isSelected ? "contained" : "outlined"}
                      onClick={() => toggleMonth(month)}
                      sx={{
                        height: '36px',
                        minWidth: 'auto',
                        borderRadius: '8px',
                        textTransform: 'none',
                        fontSize: '0.75rem',
                        fontWeight: isSelected ? 'bold' : 500,
                        bgcolor: isSelected ? '#043478' : 'white',
                        color: isSelected ? 'white' : '#043478',
                        borderColor: isSelected ? 'transparent' : '#eceff1',
                        '&:hover': {
                          bgcolor: isSelected ? '#043478' : '#f5f5f5',
                          borderColor: '#043478'
                        }
                      }}
                    >
                      {month}
                    </Button>
                  );
                })}
              </Box>

              <Button
                variant="contained"
                onClick={handleUpdateDatabaseActuals}
                disabled={loadingUpdate.dbActuals}
                fullWidth
                sx={{ mt: 4, height: '45px', borderRadius: '12px', fontWeight: 'bold' }}
              >
                {loadingUpdate.dbActuals ? <CircularProgress size={20} color="inherit" /> : 'อัปเดตการตั้งค่า'}
              </Button>
            </Box>
          </Paper>
        </Box>

        {/* Card 3: Filter Pane Config */}
        <Box>
          <Paper elevation={0}
            sx={{
              p: 4,
              borderRadius: '20px',
              border: '1px solid #e3e6f0',
              height: '100%',
              bgcolor: 'white',
              boxShadow: '0 10px 30px rgba(0,0,0,0.03)'
            }}>
            <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 3, display: 'flex', alignItems: 'center', gap: 1.5, color: '#043478' }}>
              <Avatar sx={{ bgcolor: '#e8eaf6', width: 32, height: 32 }}>
                <InsertDriveFileIcon sx={{ color: '#043478', fontSize: 20 }} />
              </Avatar>
              Filter Pane Config
            </Typography>

            <Button
              variant="outlined"
              fullWidth
              onClick={() => handleOpenModal('BUDGET_STRUCTURE', 'MANAGE BUDGET STRUCTURE')}
              sx={{
                height: '45px',
                borderRadius: '12px',
                textTransform: 'none',
                fontWeight: 'bold',
                borderColor: '#043478',
                color: '#043478',
                '&:hover': {
                  bgcolor: 'rgba(4, 52, 120, 0.04)',
                  borderColor: '#043478'
                }
              }}
              startIcon={<EditIcon />}
            >
              จัดการ Filter Pane
            </Button>
          </Paper>
        </Box>

        {/* Card 4: GL Mapping Config */}
        <Box>
          <Paper elevation={0}
            sx={{
              p: 4,
              borderRadius: '20px',
              border: '1px solid #e3e6f0',
              height: '100%',
              bgcolor: 'white',
              boxShadow: '0 10px 30px rgba(0,0,0,0.03)'
            }}>
            <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 3, display: 'flex', alignItems: 'center', gap: 1.5, color: '#043478' }}>
              <Avatar sx={{ bgcolor: '#e8eaf6', width: 32, height: 32 }}>
                <InsertDriveFileIcon sx={{ color: '#043478', fontSize: 20 }} />
              </Avatar>
              GL Mapping Config
            </Typography>

            <Button
              variant="outlined"
              fullWidth
              onClick={() => handleOpenModal('MAP_GL', 'MANAGE GL MAPPING')}
              sx={{
                height: '45px',
                borderRadius: '12px',
                textTransform: 'none',
                fontWeight: 'bold',
                borderColor: '#043478',
                color: '#043478',
                '&:hover': {
                  bgcolor: 'rgba(4, 52, 120, 0.04)',
                  borderColor: '#043478'
                }
              }}
              startIcon={<EditIcon />}
            >
              จัดการ GL Mapping
            </Button>
          </Paper>
        </Box>
      </Box>

      {/* --- Specialized Management Modals --- */}
      <BudgetVersionModal
        open={open && modalType === 'BUDGET'}
        onClose={handleClose}
        items={budgetVersions}
        onRefresh={fetchAllVersions}
        activeId={selectedBudget}
      />
      <CapexPlanModal
        open={open && modalType === 'CAPEX_BG'}
        onClose={handleClose}
        items={capexBgVersions}
        onRefresh={fetchAllVersions}
        activeId={selectedCapexBg}
      />
      <CapexActualModal
        open={open && modalType === 'CAPEX_ACTUAL'}
        onClose={handleClose}
        items={capexActualVersions}
        onRefresh={fetchAllVersions}
        activeId={selectedCapexActual}
      />

      {/* Legacy/Common Dialog removed and replaced by specialized modals above */}

      {/* --- GL Mapping Modal --- */}
      <GLMappingModal open={modalType === 'MAP_GL' && open} onClose={handleClose} />

      {/* --- Budget Structure Modal --- */}
      <BudgetStructureModal open={modalType === 'BUDGET_STRUCTURE' && open} onClose={handleClose} />

    </Box>
  );
};

export default DataManagePage;