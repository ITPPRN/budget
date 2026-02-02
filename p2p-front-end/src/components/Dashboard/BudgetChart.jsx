import React from 'react';
import { Paper, Typography, Box } from '@mui/material';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const BudgetChart = ({ data }) => {
    return (
        <Paper sx={{ p: 2, height: '100%', width: '100%', minWidth: 0, display: 'flex', flexDirection: 'column' }}>
            <Typography variant="h6" sx={{ mb: 2, fontWeight: 'bold' }}>
                Budget vs Actual
            </Typography>
            <Box sx={{ flexGrow: 1, minHeight: 300, width: '100%', minWidth: 0 }}>
                <ResponsiveContainer width="100%" height="100%">
                    <LineChart
                        data={data}
                        margin={{
                            top: 5,
                            right: 30,
                            left: 20,
                            bottom: 5,
                        }}
                    >
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="name" />
                        <YAxis tickFormatter={(value) => `${(value / 1000000).toFixed(1)}`} />
                        <Tooltip />
                        <Legend />
                        <Line type="monotone" dataKey="budget" stroke="#1976d2" activeDot={{ r: 8 }} name="Budget" strokeWidth={2} />
                        <Line type="monotone" dataKey="actual" stroke="#00bcd4" name="Actual" strokeWidth={2} />
                    </LineChart>
                </ResponsiveContainer>
            </Box>
        </Paper>
    );
};

export default BudgetChart;
