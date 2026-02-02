import React from 'react';
import { Paper, Box, Typography } from '@mui/material';

const StatCard = ({ title, value, subValues = [], icon, color = 'primary.main', bgcolor = 'background.paper', textColor = 'text.primary' }) => {
    return (
        <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column', height: '100%', position: 'relative', overflow: 'hidden', bgcolor: bgcolor }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
                <Box>
                    <Typography variant="subtitle2" sx={{ fontWeight: 'bold', mb: 0.5, color: textColor === 'text.primary' ? 'text.secondary' : textColor }}>
                        {title}
                    </Typography>
                    <Typography variant="h4" sx={{ fontWeight: 'bold', color: textColor }}>
                        {value}
                    </Typography>
                </Box>
                <Box sx={{
                    bgcolor: 'action.hover',
                    p: 1,
                    borderRadius: '50%',
                    color: color,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center'
                }}>
                    {icon}
                </Box>
            </Box>

            {/* Sub Values / Footer */}
            {subValues.length > 0 && (
                <Box sx={{ mt: 'auto', display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                    {subValues.map((sv, idx) => (
                        <Typography key={idx} variant="caption" sx={{ color: textColor, opacity: 0.8 }}>
                            {sv}
                        </Typography>
                    ))}
                </Box>
            )}
        </Paper>
    );
};

export default StatCard;
