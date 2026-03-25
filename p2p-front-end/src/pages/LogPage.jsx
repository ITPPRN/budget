import React from 'react';
import { Box, Typography, Paper, Container } from '@mui/material';
import HistoryIcon from '@mui/icons-material/History';

const LogPage = () => {
    return (
        <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
            <Paper 
                elevation={0}
                sx={{ 
                    p: 4, 
                    borderRadius: 3, 
                    textAlign: 'center',
                    border: '1px dashed',
                    borderColor: 'divider',
                    backgroundColor: 'rgba(0,0,0,0.02)'
                }}
            >
                <HistoryIcon sx={{ fontSize: 60, color: 'text.secondary', mb: 2 }} />
                <Typography variant="h4" fontWeight="bold" gutterBottom>
                    System Logs
                </Typography>
                <Typography variant="body1" color="text.secondary">
                    Wait for Development - This page will display system activity and synchronization logs.
                </Typography>
            </Paper>
        </Container>
    );
};

export default LogPage;
