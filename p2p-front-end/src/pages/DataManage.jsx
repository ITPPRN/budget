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

const DataManagePage = () => {
  const { user } = useAuth();

  // --- State for Versions ---
  const [budgetVersions, setBudgetVersions] = useState([]);
  const [capexBgVersions, setCapexBgVersions] = useState([]);
  const [capexActualVersions, setCapexActualVersions] = useState([]);

  // --- Selected Versions (External) ---
  const [selectedBudget, setSelectedBudget] = useState('');
  const [selectedCapexBg, setSelectedCapexBg] = useState('');
  const [selectedCapexActual, setSelectedCapexActual] = useState('');

  // --- Modal State ---
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

  // --- Sync Handler ---
  const [syncing, setSyncing] = useState(false);
  const [lastUpdate, setLastUpdate] = useState(new Date());

  const handleGlobalSync = async () => {
    const tasks = [];
    if (selectedBudget) tasks.push({ endpoint: `/budgets/files-budget/${selectedBudget}/sync`, label: 'Budget' });
    if (selectedCapexBg) tasks.push({ endpoint: `/budgets/files-capex-budget/${selectedCapexBg}/sync`, label: 'CAPEX Plan' });
    if (selectedCapexActual) tasks.push({ endpoint: `/budgets/files-capex-actual/${selectedCapexActual}/sync`, label: 'CAPEX Actual' });

    if (tasks.length === 0) {
      toast.warning("กรุณาเลือก Version ที่ต้องการ Sync อย่างน้อย 1 รายการ");
      return;
    }

    if (!window.confirm(`ระบบจะทำการลบข้อมูลเก่าและแทนที่ด้วยข้อมูลจากไฟล์ที่เลือก (${tasks.length} รายการ) คุณแน่ใจหรือไม่?`)) return;

    setSyncing(true);
    let successCount = 0;

    for (const task of tasks) {
      try {
        await api.post(task.endpoint);
        successCount++;
      } catch (error) {
        console.error(`Sync ${task.label} failed:`, error);
        toast.error(`Sync ${task.label} ล้มเหลว: ${error.message}`);
      }
    }

    if (successCount > 0) {
      toast.success(`Sync สำเร็จ ${successCount} รายการ`);
      setLastUpdate(new Date());
    }
    setSyncing(false);
  };

  return (
    <Box sx={{ p: 3, bgcolor: '#f5f7fb', minHeight: '100vh' }}>

      {/* 1. Header */}
      <Box sx={{
        display: 'flex',
        flexDirection: { xs: 'column', md: 'row' },
        justifyContent: 'space-between',
        alignItems: { xs: 'flex-start', md: 'flex-start' },
        mb: 3,
        gap: 2
      }}>
        <Typography variant="h4" sx={{ fontWeight: 'bold', color: 'primary.main', mt: 1 }}>
          จัดการข้อมูล
        </Typography>

        {/* Sync Widget */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, alignSelf: { xs: 'flex-end', md: 'auto' } }}>
          <Box sx={{ textAlign: 'right' }}>
            <Typography variant="h6" sx={{ fontWeight: 'bold', lineHeight: 1.2 }}>Sync data</Typography>
            <Typography variant="caption" display="block" color="text.secondary">
              Last update
            </Typography>
            <Typography variant="caption" display="block" color="text.secondary" sx={{ fontWeight: 'bold' }}>
              {lastUpdate.toLocaleString('en-GB', { day: '2-digit', month: '2-digit', year: '2-digit', hour: '2-digit', minute: '2-digit' })}
            </Typography>
          </Box>
          <IconButton
            onClick={handleGlobalSync}
            disabled={syncing}
            sx={{
              width: 50, height: 50,
              border: '3px solid',
              borderColor: syncing ? 'grey.400' : 'black',
              color: syncing ? 'grey.400' : 'black',
              '&:hover': { bgcolor: 'rgba(0,0,0,0.05)' }
            }}
          >
            {syncing ? <CircularProgress size={24} color="inherit" /> : <SyncIcon sx={{ fontSize: 30 }} />}
          </IconButton>
        </Box>
      </Box>

      {/* 2. Main Logic Area */}
      <Paper elevation={0} sx={{ p: { xs: 2, md: 4 }, borderRadius: '15px', border: '1px solid #e3e6f0' }}>

        {/* Budget Version */}
        <Grid container alignItems="center" spacing={2} sx={{ mb: 4 }}>
          <Grid item xs={12} md={3} lg={2}>
            <Typography sx={{ fontWeight: 'bold' }}>Budget version</Typography>
          </Grid>
          <Grid item xs={12} md={6} lg={4}>
            <Select
              size="small"
              value={selectedBudget}
              onChange={(e) => setSelectedBudget(e.target.value)}
              displayEmpty
              fullWidth
              sx={{ borderRadius: '10px' }}
            >
              <MenuItem value=""><em>-- Select Version --</em></MenuItem>
              {budgetVersions.map((v) => (
                <MenuItem key={v.id} value={v.id}>
                  {v.file_name} ({new Date(v.upload_at).toLocaleDateString()})
                </MenuItem>
              ))}
            </Select>
          </Grid>
          <Grid item xs={12} md={3} lg={3}>
            <Link component="button" onClick={() => handleOpenModal('BUDGET', 'MANAGE BUDGET VERSION')} sx={{ fontSize: '0.85rem', textDecoration: 'none' }}>
              Manage Budget Version
            </Link>
          </Grid>
        </Grid>

        {/* Manage Actual CAPEX Section */}
        <Box>
          <Typography sx={{ fontWeight: 'bold', mb: 2 }}>Manage Actual CAPEX</Typography>

          {/* Capex Budget (Plan) */}
          <Grid container alignItems="center" spacing={2} sx={{ mb: 2, pl: { md: 4 } }}>
            <Grid item xs={12} md={3} lg={2}>
              <Typography>Budget Version</Typography>
            </Grid>
            <Grid item xs={12} md={6} lg={4}>
              <Select
                size="small"
                value={selectedCapexBg}
                onChange={(e) => setSelectedCapexBg(e.target.value)}
                displayEmpty
                fullWidth
                sx={{ borderRadius: '10px' }}
              >
                <MenuItem value=""><em>-- Select Capex Plan --</em></MenuItem>
                {capexBgVersions.map((v) => (
                  <MenuItem key={v.id} value={v.id}>
                    {v.file_name}
                  </MenuItem>
                ))}
              </Select>
            </Grid>
            <Grid item xs={12} md={3} lg={3}>
              <Link component="button" onClick={() => handleOpenModal('CAPEX_BG', 'MANAGE CAPEX-BG VERSION')} sx={{ fontSize: '0.85rem', textDecoration: 'none' }}>
                Manage CAPEX-BG Version
              </Link>
            </Grid>
          </Grid>

          {/* Capex Actual */}
          <Grid container alignItems="center" spacing={2} sx={{ pl: { md: 4 } }}>
            <Grid item xs={12} md={3} lg={2}>
              <Typography>Actual Version</Typography>
            </Grid>
            <Grid item xs={12} md={6} lg={4}>
              <Select
                size="small"
                value={selectedCapexActual}
                onChange={(e) => setSelectedCapexActual(e.target.value)}
                displayEmpty
                fullWidth
                sx={{ borderRadius: '10px' }}
              >
                <MenuItem value=""><em>-- Select Capex Actual --</em></MenuItem>
                {capexActualVersions.map((v) => (
                  <MenuItem key={v.id} value={v.id}>
                    {v.file_name}
                  </MenuItem>
                ))}
              </Select>
            </Grid>
            <Grid item xs={12} md={3} lg={3}>
              <Link component="button" onClick={() => handleOpenModal('CAPEX_ACTUAL', 'MANAGE CAPEX-ACTUAL VERSION')} sx={{ fontSize: '0.85rem', textDecoration: 'none' }}>
                Manage CAPEX-Actual Version
              </Link>
            </Grid>
          </Grid>
        </Box>
      </Paper>

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
                              <Divider orientation="vertical" flexItem sx={{ mx: 0.5, height: 20, alignSelf: 'center' }} />
                              <Switch size="small" defaultChecked disabled />
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