import React, { useState, useEffect, useRef } from 'react';
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

const DataManagePage = () => {
  const { user } = useAuth();

  // --- State for Versions ---
  const [budgetVersions, setBudgetVersions] = useState([]);
  const [capexBgVersions, setCapexBgVersions] = useState([]);
  const [capexActualVersions, setCapexActualVersions] = useState([]);

  // --- Selected Versions (External) with Persistence (Last Synced) ---
  // Load initial state from the LAST SUCCESSFUL SYNC config, not the draft.
  const loadSavedState = (key, defaultVal) => {
    const saved = localStorage.getItem('dm_lastSyncedConfig');
    if (saved) {
      const config = JSON.parse(saved);
      return config[key] || defaultVal;
    }
    return defaultVal;
  };

  const [selectedBudget, setSelectedBudget] = useState(() => loadSavedState('selectedBudget', ''));
  const [selectedCapexBg, setSelectedCapexBg] = useState(() => loadSavedState('selectedCapexBg', ''));
  const [selectedCapexActual, setSelectedCapexActual] = useState(() => loadSavedState('selectedCapexActual', ''));

  // --- New State for Operational Actuals ---
  const [actualYear, setActualYear] = useState(() => loadSavedState('actualYear', new Date().getFullYear()));
  const [selectedMonths, setSelectedMonths] = useState(() => loadSavedState('selectedMonths', []));

  // Note: We NO LONGER save to localStorage on every change (useEffect). 
  // We only save when the user successfully SYNCS.

  const [open, setOpen] = useState(false);
  const [modalType, setModalType] = useState(''); // 'BUDGET', 'CAPEX_BG', 'CAPEX_ACTUAL'
  const [modalTitle, setModalTitle] = useState('');

  // --- Search & Upload State inside Modal ---
  const [searchTerm, setSearchTerm] = useState('');
  const [fileToUpload, setFileToUpload] = useState(null);
  const [versionName, setVersionName] = useState('');
  const [isUploading, setIsUploading] = useState(false);
  const fileInputRef = useRef(null);

  // --- Rename State ---
  const [editingId, setEditingId] = useState(null);
  const [newName, setNewName] = useState('');

  // --- Fetch Data on Load ---
  const fetchAllVersions = async () => {
    try {
      const [resBudget, resCapexBg, resCapexAct] = await Promise.all([
        api.get('/budgets/files-budget'),
        api.get('/budgets/files-capex-budget'),
        api.get('/budgets/files-capex-actual'),
      ]);

      setBudgetVersions(resBudget.data || []);
      setCapexBgVersions(resCapexBg.data || []);
      setCapexActualVersions(resCapexAct.data || []);
    } catch (error) {
      console.error("Failed to fetch versions:", error);
      toast.error("ไม่สามารถดึงข้อมูล Version ได้");
    }
  };

  useEffect(() => {
    fetchAllVersions();
  }, []);

  // --- Handlers ---
  const handleOpenModal = (type, title) => {
    setModalType(type);
    setModalTitle(title);
    setSearchTerm('');
    setFileToUpload(null);
    setVersionName('');
    setOpen(true);
  };

  const handleClose = () => setOpen(false);

  const getActiveList = () => {
    switch (modalType) {
      case 'BUDGET': return budgetVersions;
      case 'CAPEX_BG': return capexBgVersions;
      case 'CAPEX_ACTUAL': return capexActualVersions;
      default: return [];
    }
  };

  const activeList = getActiveList().filter(item =>
    item.file_name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // --- Upload Handlers ---
  const handleFileSelect = (event) => {
    const file = event.target.files[0];
    if (file) {
      setFileToUpload(file);
      setVersionName(file.name); // Default to filename
    }
  };

  const handleSaveUpload = async () => {
    if (!fileToUpload) return;
    setIsUploading(true);

    const formData = new FormData();
    formData.append('file', fileToUpload);
    formData.append('version_name', versionName); // Send custom name

    let endpoint = '';
    if (modalType === 'BUDGET') endpoint = '/budgets/import-budget';
    else if (modalType === 'CAPEX_BG') endpoint = '/budgets/import-capex-budget';
    else if (modalType === 'CAPEX_ACTUAL') endpoint = '/budgets/import-capex-actual';

    try {
      await api.post(endpoint, formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      });
      toast.success("อัปโหลดไฟล์สำเร็จ!");
      setFileToUpload(null);
      setVersionName('');
      await fetchAllVersions(); // Refresh list
    } catch (error) {
      console.error("Upload error:", error);
      toast.error("อัปโหลดล้มเหลว: " + (error.response?.data?.error || error.message));
    } finally {
      setIsUploading(false);
    }
  };

  // --- Delete Handler ---
  const handleDelete = async (id) => {
    if (!window.confirm("คุณต้องการลบไฟล์นี้ใช่หรือไม่?")) return;

    let endpoint = '';
    if (modalType === 'BUDGET') endpoint = `/budgets/files-budget/${id}`;
    else if (modalType === 'CAPEX_BG') endpoint = `/budgets/files-capex-budget/${id}`;
    else if (modalType === 'CAPEX_ACTUAL') endpoint = `/budgets/files-capex-actual/${id}`;

    try {
      await api.delete(endpoint);
      toast.success("ลบไฟล์สำเร็จ");
      fetchAllVersions();
    } catch (error) {
      console.error("Delete failed", error);
      toast.error("ลบไฟล์ไม่สำเร็จ");
    }
  };

  // --- Rename Handlers ---
  const startRename = (id, currentName) => {
    setEditingId(id);
    setNewName(currentName);
  };

  const MONTHS = ['JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN', 'JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC'];

  const toggleMonth = (month) => {
    const monthIdx = MONTHS.indexOf(month);

    // Check if the clicked month is already the end of the selection range
    if (selectedMonths.length > 0 && selectedMonths[selectedMonths.length - 1] === month) {
      // Toggle off
      setSelectedMonths([]);
    } else {
      // If selecting a new month, select all months from JAN to that month (Cumulative)
      const newSelected = MONTHS.slice(0, monthIdx + 1);
      setSelectedMonths(newSelected);
    }
  };

  const saveRename = async (id) => {
    let endpoint = '';
    if (modalType === 'BUDGET') endpoint = `/budgets/files-budget/${id}`;
    else if (modalType === 'CAPEX_BG') endpoint = `/budgets/files-capex-budget/${id}`;
    else if (modalType === 'CAPEX_ACTUAL') endpoint = `/budgets/files-capex-actual/${id}`;

    try {
      await api.patch(endpoint, { new_name: newName });
      toast.success("เปลี่ยนชื่อสำเร็จ");
      setEditingId(null);
      fetchAllVersions();
    } catch (error) {
      console.error("Rename failed", error);
      toast.error("เปลี่ยนชื่อไม่สำเร็จ");
    }
  };

  // --- Sync State with Persistence for Last Update ---
  const [syncing, setSyncing] = useState(false);
  const [lastUpdate, setLastUpdate] = useState(() => {
    const saved = localStorage.getItem('dm_lastUpdate');
    return saved ? new Date(saved) : new Date();
  });

  useEffect(() => {
    localStorage.setItem('dm_lastUpdate', lastUpdate.toISOString());
  }, [lastUpdate]);

  const handleGlobalSync = async () => {
    const tasks = [];

    // Budget: If selected -> Sync, If empty -> Clear
    if (selectedBudget) {
      tasks.push({ endpoint: `/budgets/files-budget/${selectedBudget}/sync`, label: 'Budget (Update)' });
    } else {
      tasks.push({ endpoint: `/budgets/clear-budget`, label: 'Budget (Clear)' });
    }

    // Capex Plan: If selected -> Sync, If empty -> Clear
    if (selectedCapexBg) {
      tasks.push({ endpoint: `/budgets/files-capex-budget/${selectedCapexBg}/sync`, label: 'CAPEX Plan (Update)' });
    } else {
      tasks.push({ endpoint: `/budgets/clear-capex-budget`, label: 'CAPEX Plan (Clear)' });
    }

    // Capex Actual: If selected -> Sync, If empty -> Clear
    if (selectedCapexActual) {
      tasks.push({ endpoint: `/budgets/files-capex-actual/${selectedCapexActual}/sync`, label: 'CAPEX Actual (Update)' });
    } else {
      tasks.push({ endpoint: `/budgets/clear-capex-actual`, label: 'CAPEX Actual (Clear)' });
    }

    // Add Operational Actuals task if year and months are selected
    if (actualYear && selectedMonths.length > 0) {
      tasks.push({
        endpoint: `/budgets/sync-actuals`,
        data: { year: String(actualYear), months: selectedMonths },
        label: `Actuals (${actualYear} ${selectedMonths[0]}-${selectedMonths[selectedMonths.length - 1]})`
      });
    }

    if (tasks.length === 0) {
      toast.warning("กรุณาเลือกรายการที่ต้องการ Sync อย่างน้อย 1 รายการ");
      return;
    }

    if (!window.confirm(`ระบบจะทำการลบข้อมูลเก่าและแทนที่ด้วยข้อมูลใหม่ (${tasks.length} รายการ) คุณแน่ใจหรือไม่?`)) return;

    setSyncing(true);
    let successCount = 0;

    for (const task of tasks) {
      try {
        if (task.data) {
          await api.post(task.endpoint, task.data);
        } else {
          await api.post(task.endpoint);
        }
        successCount++;
        toast.info(`Synced ${task.label} แล้ว`);
      } catch (error) {
        console.error(`Sync ${task.label} failed:`, error);
        toast.error(`Sync ${task.label} ล้มเหลว: ${error.response?.data?.error || error.message}`);
      }
    }

    if (successCount === tasks.length) {
      toast.success(`Sync สำเร็จครบทั้ง ${successCount} รายการ! 🎉`);
    } else if (successCount > 0) {
      toast.warning(`Sync สำเร็จ ${successCount} จาก ${tasks.length} รายการ`);
    }

    // Save success state to localStorage to persist "Last Synced Config"
    const successConfig = {
      selectedBudget,
      selectedCapexBg,
      selectedCapexActual,
      actualYear,
      selectedMonths
    };
    localStorage.setItem('dm_lastSyncedConfig', JSON.stringify(successConfig));

    setLastUpdate(new Date());
    setSyncing(false);
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
          <Typography variant="body2" color="text.secondary" sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <SyncIcon fontSize="small" color="action" /> Last Global Sync: {lastUpdate.toLocaleString('en-GB')}
          </Typography>
        </Box>

        <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
          <Button
            variant="contained"
            onClick={handleGlobalSync}
            disabled={syncing}
            startIcon={syncing ? <CircularProgress size={20} color="inherit" /> : <SyncIcon />}
            sx={{
              height: '60px',
              px: 4,
              borderRadius: '12px',
              fontSize: '1.1rem',
              fontWeight: 'bold',
              background: 'linear-gradient(45deg, #1a237e 30%, #283593 90%)',
              boxShadow: '0 8px 25px rgba(26, 35, 126, 0.2)',
              transition: 'all 0.3s ease',
              '&:hover': {
                background: 'linear-gradient(45deg, #283593 30%, #1a237e 90%)',
                transform: 'translateY(-2px)',
                boxShadow: '0 12px 30px rgba(26, 35, 126, 0.3)',
              },
              '&:active': {
                transform: 'translateY(0)',
              }
            }}
          >
            {syncing ? 'กำลัง SYNC...' : 'Sync Data'}
          </Button>
        </Box>
      </Box>

      {/* 2. Main Logic Area */}
      <Grid container spacing={3} sx={{ width: '100%', flexWrap: 'nowrap' }}>
        {/* Left Column: Version Selections */}
        <Grid item xs={6} sx={{ minWidth: 0, maxWidth: '50%', flexBasis: '50%' }}>
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
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1.5 }}>
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
                <Select
                  size="small"
                  value={selectedBudget}
                  onChange={(e) => setSelectedBudget(e.target.value)}
                  displayEmpty
                  fullWidth
                  sx={{
                    borderRadius: '12px',
                    bgcolor: '#fcfcfc',
                    '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e0e0e0' },
                    '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: '#1a237e' }
                  }}
                >
                  <MenuItem value=""><em>-- เลือกไฟล์งบประมาณ --</em></MenuItem>
                  {budgetVersions.map((v) => (
                    <MenuItem key={v.id} value={v.id}>{v.file_name}</MenuItem>
                  ))}
                </Select>
              </Box>

              <Divider sx={{ my: 4, borderStyle: 'dashed' }} />

              <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 3, display: 'flex', alignItems: 'center', gap: 1.5, color: '#1a237e' }}>
                <Avatar sx={{ bgcolor: '#e8eaf6', width: 32, height: 32 }}>
                  <DescriptionIcon sx={{ color: '#1a237e', fontSize: 20 }} />
                </Avatar>
                CAPEX
              </Typography>

              <Box sx={{ mb: 3, pl: 0 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
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
                <Select
                  size="small"
                  value={selectedCapexBg}
                  onChange={(e) => setSelectedCapexBg(e.target.value)}
                  displayEmpty
                  fullWidth
                  sx={{
                    borderRadius: '12px',
                    bgcolor: '#fcfcfc',
                    '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e0e0e0' }
                  }}
                >
                  <MenuItem value=""><em>-- เลือก Capex Plan --</em></MenuItem>
                  {capexBgVersions.map((v) => (
                    <MenuItem key={v.id} value={v.id}>{v.file_name}</MenuItem>
                  ))}
                </Select>
              </Box>

              <Box sx={{ pl: 0 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
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
                <Select
                  size="small"
                  value={selectedCapexActual}
                  onChange={(e) => setSelectedCapexActual(e.target.value)}
                  displayEmpty
                  fullWidth
                  sx={{
                    borderRadius: '12px',
                    bgcolor: '#fcfcfc',
                    '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e0e0e0' }
                  }}
                >
                  <MenuItem value=""><em>-- เลือก Capex Actual --</em></MenuItem>
                  {capexActualVersions.map((v) => (
                    <MenuItem key={v.id} value={v.id}>{v.file_name}</MenuItem>
                  ))}
                </Select>
              </Box>
            </Box>
          </Paper>
        </Grid>

        {/* Right Column: Operational Actuals */}
        <Grid item xs={6} sx={{ minWidth: 0, maxWidth: '50%', flexBasis: '50%' }}>
          <Paper elevation={0}
            sx={{
              p: 4,
              borderRadius: '20px',
              border: '1px solid #e3e6f0',
              width: '100%',
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              position: 'relative',
              overflow: 'hidden'
            }}>

            {/* Background Decorative Circle */}
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
                sx={{
                  borderRadius: '12px',
                  mb: 4,
                  bgcolor: 'white',
                  '& .MuiOutlinedInput-notchedOutline': { borderColor: '#e0e0e0' },
                  '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: '#1a237e' }
                }}
              >
                {/* ใช้ dynamicYears แทนการระบุปีตรงๆ */}
                {Array.from({ length: 5 }, (_, i) => (new Date().getFullYear() + 2) - i).map(year => (
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

              <Grid container spacing={1.5}>
                {MONTHS.map((month) => {
                  const isSelected = selectedMonths.includes(month);
                  return (
                    <Grid item xs={12} sm={6} md={4} key={month}>
                      <Button
                        fullWidth
                        variant={isSelected ? "contained" : "outlined"}
                        onClick={() => toggleMonth(month)}
                        sx={{
                          height: '45px',
                          borderRadius: '10px',
                          textTransform: 'none',
                          fontSize: '0.85rem',
                          fontWeight: isSelected ? 'bold' : 500,
                          bgcolor: isSelected ? '#1a237e' : 'white',
                          color: isSelected ? 'white' : '#1a237e',
                          borderColor: isSelected ? 'transparent' : '#eceff1',
                          boxShadow: isSelected ? '0 4px 12px rgba(26, 35, 126, 0.2)' : 'none',
                          '&:hover': {
                            bgcolor: isSelected ? '#283593' : '#f5f5f5',
                            borderColor: '#1a237e',
                            transform: 'translateY(-1px)',
                            boxShadow: isSelected ? '0 6px 15px rgba(26, 35, 126, 0.3)' : '0 2px 5px rgba(0,0,0,0.05)',
                          },
                          transition: 'all 0.2s cubic-bezier(0.4, 0, 0.2, 1)'
                        }}
                      >
                        {month}
                      </Button>
                    </Grid>
                  );
                })}
              </Grid>
            </Box>
          </Paper>
        </Grid>
      </Grid>

      {/* --- Modal (Dialog) --- */}
      <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
        <DialogTitle sx={{ bgcolor: '#4e73df', color: 'white', display: 'flex', justifyContent: 'space-between', alignItems: 'center', py: 1 }}>
          <Typography sx={{ fontWeight: 'bold', textTransform: 'uppercase' }}>{modalTitle}</Typography>
          <IconButton onClick={handleClose} sx={{ color: 'white' }}><CloseIcon /></IconButton>
        </DialogTitle>
        <DialogContent sx={{ p: 4, bgcolor: '#f8f9fc' }}>
          <Grid container spacing={4}>

            {/* Left: Available Versions List */}
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 1 }}>AVAILABLE VERSIONS</Typography>
              <TextField
                fullWidth size="small" placeholder="search..." sx={{ mb: 1, bgcolor: 'white' }}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                InputProps={{ endAdornment: <InputAdornment position="end"><SearchIcon color="primary" /></InputAdornment> }}
              />
              <Typography variant="caption" color="text.secondary" sx={{ fontStyle: 'italic', mb: 2, display: 'block' }}>
                Total versions: {activeList.length}
              </Typography>

              <Paper sx={{ maxHeight: 300, overflow: 'auto', bgcolor: 'white' }}>
                <List dense>
                  {activeList.map((item) => (
                    <React.Fragment key={item.id}>
                      <ListItem sx={{ py: 1.5, px: 2 }}>
                        {/* 1. Avatar */}
                        <ListItemAvatar>
                          <Avatar sx={{ bgcolor: '#e3f2fd', color: '#1976d2' }}>
                            <InsertDriveFileIcon style={{ fontSize: 18 }} />
                          </Avatar>
                        </ListItemAvatar>

                        {/* 2. Content Area (Text or Edit Input) */}
                        <Box sx={{ flexGrow: 1, minWidth: 0, mr: 2 }}>
                          {editingId === item.id ? (
                            <TextField
                              size="small"
                              value={newName}
                              onChange={(e) => setNewName(e.target.value)}
                              autoFocus
                              fullWidth
                              variant="standard"
                            />
                          ) : (
                            <>
                              <Typography variant="body2" sx={{ fontWeight: 600 }} noWrap title={item.file_name}>
                                {item.file_name}
                              </Typography>
                              <Typography variant="caption" display="block" color="text.secondary" noWrap>
                                {new Date(item.upload_at).toLocaleString()}
                              </Typography>
                            </>
                          )}
                        </Box>

                        {/* 3. Actions Area */}
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                          {editingId === item.id ? (
                            <>
                              <IconButton size="small" onClick={() => saveRename(item.id)} sx={{ color: 'success.main' }}>
                                <SaveIcon fontSize="small" />
                              </IconButton>
                              <IconButton size="small" onClick={() => setEditingId(null)} color="default">
                                <CancelIcon fontSize="small" />
                              </IconButton>
                            </>
                          ) : (
                            <>
                              <IconButton size="small" onClick={() => startRename(item.id, item.file_name)} sx={{ color: 'primary.main' }}>
                                <EditIcon fontSize="small" />
                              </IconButton>
                              <IconButton size="small" onClick={() => handleDelete(item.id)} sx={{ color: 'error.main' }}>
                                <DeleteIcon fontSize="small" />
                              </IconButton>
                            </>
                          )}
                        </Box>
                      </ListItem>
                      <Divider component="li" />
                    </React.Fragment>
                  ))}
                  {activeList.length === 0 && (
                    <Box sx={{ p: 2, textAlign: 'center', color: '#999' }}>No versions found</Box>
                  )}
                </List>
              </Paper>
            </Grid>

            {/* Right: Upload Area */}
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>UPLOAD NEW VERSION</Typography>

              {/* 1. Upload Box */}
              {!fileToUpload ? (
                <Box
                  sx={{ border: '2px dashed #ccc', borderRadius: '15px', p: 4, textAlign: 'center', bgcolor: 'white', mb: 3, cursor: 'pointer' }}
                  onClick={() => fileInputRef.current.click()}
                >
                  <input
                    type="file"
                    hidden
                    ref={fileInputRef}
                    onChange={handleFileSelect}
                    accept=".xlsx,.csv"
                  />
                  <Typography variant="body2" color="text.secondary">Drag & drop your budget file here or click to browse</Typography>
                  <Button variant="contained" startIcon={<CloudUploadIcon />} sx={{ mt: 2, borderRadius: '20px' }}>
                    SELECT FILE
                  </Button>
                  <Typography variant="caption" display="block" sx={{ mt: 1 }} color="text.secondary">Supported file: .xlsx</Typography>
                </Box>
              ) : (
                <Box sx={{ border: '1px solid #ccc', borderRadius: '15px', p: 3, bgcolor: 'white', mb: 3, textAlign: 'center' }}>
                  <Typography variant="body1" sx={{ fontWeight: 'bold', mb: 1 }}>Selected File:</Typography>
                  <Typography variant="body2" color="primary" sx={{ mb: 2 }}>{fileToUpload.name}</Typography>

                  <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 1, textAlign: 'left' }}>Version Name:</Typography>
                  <TextField
                    fullWidth
                    size="small"
                    value={versionName}
                    onChange={(e) => setVersionName(e.target.value)}
                    sx={{ mb: 2 }}
                  />

                  <Button
                    variant="contained"
                    color="success"
                    sx={{ borderRadius: '20px', px: 4 }}
                    onClick={handleSaveUpload}
                    disabled={isUploading}
                    startIcon={isUploading ? <CircularProgress size={20} color="inherit" /> : <SaveIcon />}
                  >
                    {isUploading ? "UPLOADING..." : "SAVE & UPLOAD"}
                  </Button>

                  <Button
                    sx={{ ml: 2, color: 'text.secondary' }}
                    onClick={() => setFileToUpload(null)}
                  >
                    CANCEL
                  </Button>
                </Box>
              )}
            </Grid>

          </Grid>
        </DialogContent>
      </Dialog>
    </Box>
  );
};

export default DataManagePage;