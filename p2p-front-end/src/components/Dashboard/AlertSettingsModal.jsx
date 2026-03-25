import React, { useState, useEffect } from 'react';
import { Dialog, DialogTitle, DialogContent, DialogActions, TextField, Button, Box, Typography, InputAdornment, Paper } from '@mui/material';

const AlertSettingsModal = ({ open, onClose, onSave, initialValues }) => {
    const [values, setValues] = useState({
        red: 100,
        yellow: 80
    });

    useEffect(() => {
        if (initialValues) {
            setValues({
                red: initialValues.red || 100,
                yellow: initialValues.yellow || 80
            });
        }
    }, [initialValues, open]);

    const handleChange = (field) => (e) => {
        const val = parseInt(e.target.value);
        if (!isNaN(val)) {
            setValues(prev => ({ ...prev, [field]: val }));
        }
    };

    const handleSave = () => {
        onSave(values);
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle sx={{ fontWeight: 'bold', color: 'primary.main', pb: 1 }}>
                Dashboard Alert Thresholds
            </DialogTitle>
            <DialogContent>
                <Box sx={{ mt: 1, display: 'flex', flexDirection: 'column', gap: 3 }}>
                    <Typography variant="body2" color="text.secondary">
                        Configure the percentage thresholds that determine the status flags and summary counts.
                    </Typography>

                    {/* Rule 1: Red */}
                    <Paper variant="outlined" sx={{ p: 2, borderLeft: '6px solid #e74a3b', bgcolor: 'rgba(231, 74, 59, 0.02)' }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1, gap: 1 }}>
                            <Box sx={{ width: 12, height: 12, borderRadius: '50%', bgcolor: '#e74a3b' }} />
                            <Typography variant="subtitle2" fontWeight="bold" color="#e74a3b">
                                🟡 → 🔴 Over Budget (CRITICAL)
                            </Typography>
                        </Box>
                        <TextField
                            label="Red Alert Threshold"
                            type="number"
                            fullWidth
                            size="small"
                            value={values.red}
                            onChange={handleChange('red')}
                            InputProps={{
                                endAdornment: <InputAdornment position="end">%</InputAdornment>,
                            }}
                            helperText="If %spend ≥ this value, show RED flag."
                        />
                    </Paper>

                    {/* Rule 2: Yellow */}
                    <Paper variant="outlined" sx={{ p: 2, borderLeft: '6px solid #f6c23e', bgcolor: 'rgba(246, 194, 62, 0.02)' }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1, gap: 1 }}>
                            <Box sx={{ width: 12, height: 12, borderRadius: '50%', bgcolor: '#f6c23e' }} />
                            <Typography variant="subtitle2" fontWeight="bold" color="#f6c23e">
                                🟢 → 🟡 Near Limit (WARNING)
                            </Typography>
                        </Box>
                        <TextField
                            label="Yellow Alert Threshold"
                            type="number"
                            fullWidth
                            size="small"
                            value={values.yellow}
                            onChange={handleChange('yellow')}
                            InputProps={{
                                endAdornment: <InputAdornment position="end">%</InputAdornment>,
                            }}
                            helperText="If %spend ≥ this value (but < Red), show YELLOW flag."
                        />
                    </Paper>

                    {/* Rule 3: Green (Implicit) */}
                    <Paper variant="outlined" sx={{ p: 2, borderLeft: '6px solid #1cc88a', bgcolor: 'rgba(28, 200, 138, 0.02)', opacity: 0.8 }}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Box sx={{ width: 12, height: 12, borderRadius: '50%', bgcolor: '#1cc88a' }} />
                            <Typography variant="subtitle2" fontWeight="bold" color="#1cc88a">
                                🟢 In Budget (SAFE)
                            </Typography>
                        </Box>
                        <Typography variant="caption" display="block" sx={{ mt: 1 }}>
                            Automatic for everything below {values.yellow}%
                        </Typography>
                    </Paper>
                </Box>
            </DialogContent>
            <DialogActions sx={{ p: 2, bgcolor: '#f8f9fc' }}>
                <Button onClick={onClose} color="inherit">Cancel</Button>
                <Button onClick={handleSave} variant="contained" color="primary" sx={{ px: 4 }}>
                    Apply & Sync
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default AlertSettingsModal;
