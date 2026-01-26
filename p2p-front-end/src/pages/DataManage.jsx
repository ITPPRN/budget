import React, { useState } from 'react';
import { 
  Box, Grid, Paper, Typography, Button, Select, 
  MenuItem, Switch, Link, IconButton, Dialog, 
  DialogTitle, DialogContent, TextField, InputAdornment 
} from '@mui/material';
import SyncIcon from '@mui/icons-material/Sync';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import SearchIcon from '@mui/icons-material/Search';
import DeleteIcon from '@mui/icons-material/Delete';
import EditNoteIcon from '@mui/icons-material/EditNote';
import CloseIcon from '@mui/icons-material/Close';
import DescriptionIcon from '@mui/icons-material/Description';
import { useAuth } from '../hooks/useAuth';

const DataManagePage = () => {
  const { user } = useAuth();
  
  // State สำหรับควบคุมการเปิด/ปิด และหัวข้อของ Modal
  const [open, setOpen] = useState(false);
  const [modalTitle, setModalTitle] = useState('');

  const handleOpenModal = (title) => {
    setModalTitle(title);
    setOpen(true);
  };

  const handleClose = () => setOpen(false);

  return (
    <Box sx={{ p: 3, bgcolor: '#f5f7fb', minHeight: '100vh' }}>
      
      {/* 1. ส่วนหัวข้อ (Header) และปุ่ม Sync */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 3 }}>
        <Typography variant="h4" sx={{ fontWeight: 'bold', color: 'primary.main' }}>
          จัดการข้อมูล
        </Typography>
      </Box>

      {/* 2. พื้นที่จัดการข้อมูลหลัก */}
      <Paper elevation={0} sx={{ p: 4, borderRadius: '15px', border: '1px solid #e3e6f0' }}>
        
        {/* Budget Version */}
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 4 }}>
          <Typography sx={{ fontWeight: 'bold', width: '180px' }}>Budget version</Typography>
          <Select size="small" defaultValue="V1" sx={{ width: '350px', mr: 2, borderRadius: '10px' }}>
            <MenuItem value="V1">-</MenuItem>
          </Select>
          <Link component="button" onClick={() => handleOpenModal('MANAGE BUDGET VERSION')} sx={{ fontSize: '0.85rem', textDecoration: 'none' }}>
            Manage Budget Version
          </Link>
        </Box>

        {/* Manage Actual Expense */}
        <Box sx={{ mb: 4 }}>
        
        </Box>

        {/* Manage Actual CAPEX */}
        <Box>
          <Typography sx={{ fontWeight: 'bold', mb: 2 }}>Manage Actual CAPEX</Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', ml: 4, mb: 2 }}>
            <Typography sx={{ width: '140px' }}>Budget Version</Typography>
            <Select size="small" defaultValue="CAPEX" sx={{ width: '350px', mr: 2, borderRadius: '10px' }}>
              <MenuItem value="CAPEX">-</MenuItem>
            </Select>
            <Link component="button" onClick={() => handleOpenModal('MANAGE CAPEX-BG VERSION')} sx={{ fontSize: '0.85rem', textDecoration: 'none' }}>
              Manage CAPEX-BG Version
            </Link>
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', ml: 4 }}>
            <Typography sx={{ width: '140px' }}>Actual Version</Typography>
            <Select size="small" defaultValue="V1" sx={{ width: '350px', mr: 2, borderRadius: '10px' }}>
              <MenuItem value="V1">-</MenuItem>
            </Select>
            <Link component="button" onClick={() => handleOpenModal('MANAGE CAPEX-ACTUAL VERSION')} sx={{ fontSize: '0.85rem', textDecoration: 'none' }}>
              Manage CAPEX-Actual Version
            </Link>
          </Box>
        </Box>
      </Paper>

      {/* --- ส่วนของ Modal (Dialog) --- */}
      <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
        <DialogTitle sx={{ bgcolor: '#4e73df', color: 'white', display: 'flex', justifyContent: 'space-between', alignItems: 'center', py: 1 }}>
          <Typography sx={{ fontWeight: 'bold', textTransform: 'uppercase' }}>{modalTitle}</Typography>
          <IconButton onClick={handleClose} sx={{ color: 'white' }}><CloseIcon /></IconButton>
        </DialogTitle>
        <DialogContent sx={{ p: 4, bgcolor: '#f8f9fc' }}>
          <Grid container spacing={4}>
            {/* ฝั่งซ้าย: Available Versions */}
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 1 }}>AVAILABLE VERSIONS</Typography>
              <TextField 
                fullWidth size="small" placeholder="search" sx={{ mb: 1, bgcolor: 'white' }}
                InputProps={{ endAdornment: <InputAdornment position="end"><SearchIcon color="primary" /></InputAdornment> }}
              />
              <Typography variant="caption" color="text.secondary" sx={{ fontStyle: 'italic', mb: 2, display: 'block' }}>Total versions: 0</Typography>
              {/* รายการ Version จะมาแสดงตรงนี้ */}
            </Grid>
            {/* ฝั่งขวา: Upload New Version */}
            <Grid item xs={12} md={6}>
              <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 2 }}>UPLOAD NEW VERSION</Typography>
              <Box sx={{ border: '2px dashed #ccc', borderRadius: '15px', p: 4, textAlign: 'center', bgcolor: 'white', mb: 3 }}>
                <Typography variant="body2" color="text.secondary">Drag & drop your budget file here or click to browse</Typography>
                <Button variant="contained" startIcon={<CloudUploadIcon />} sx={{ mt: 2, borderRadius: '20px' }}>UPLOAD FILE</Button>
                <Typography variant="caption" display="block" sx={{ mt: 1 }} color="text.secondary">Supported file: xlsx, csv</Typography>
              </Box>
              <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 1 }}>Version Name</Typography>
              <Box sx={{ display: 'flex', gap: 1 }}>
                <TextField fullWidth size="small" placeholder="Enter version name" sx={{ bgcolor: 'white' }} />
                <Button variant="contained" sx={{ px: 4, borderRadius: '10px' }}>SAVE</Button>
              </Box>
            </Grid>
          </Grid>
        </DialogContent>
      </Dialog>
    </Box>
  );
};

export default DataManagePage;