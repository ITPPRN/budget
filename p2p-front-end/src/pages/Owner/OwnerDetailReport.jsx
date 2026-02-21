import React from 'react';
import { Box, Typography } from '@mui/material';

const OwnerDetailReport = () => {
    return (
        <Box sx={{ p: 3 }}>
            <Typography variant="h4" gutterBottom>
                Owner Detail Report
            </Typography>
            <Box sx={{ mt: 2, p: 4, border: '1px dashed grey', borderRadius: 2, textAlign: 'center' }}>
                <Typography color="text.secondary">
                    Content for Owner Detail Report will be displayed here.
                </Typography>
            </Box>
        </Box>
    );
};

export default OwnerDetailReport;
