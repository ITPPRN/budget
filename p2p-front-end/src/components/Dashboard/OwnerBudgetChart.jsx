import React from 'react';
import { Paper, Typography, Box } from '@mui/material';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const BudgetChart = ({ data }) => {
    return (
        <Box sx={{ width: '100%', height: '100%', minHeight: 300 }}>
            <ResponsiveContainer width="100%" height="100%">
                <LineChart
                    data={data}
                    margin={{
                        top: 20,
                        right: 30,
                        left: 0,
                        bottom: 0,
                    }}
                >
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e0e0e0" />
                    <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fill: '#9e9e9e', fontSize: 12 }} dy={10} />
                    <YAxis
                        axisLine={false}
                        tickLine={false}
                        tick={{ fill: '#9e9e9e', fontSize: 12 }}
                        tickFormatter={(value) => `${(value / 1000000).toFixed(1)}M`}
                    />
                    <Tooltip
                        contentStyle={{ borderRadius: 8, border: 'none', boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}
                        formatter={(value) => [`${new Intl.NumberFormat('en-US').format(value)}`, undefined]}
                    />
                    <Legend wrapperStyle={{ paddingTop: '20px' }} />
                    <Line type="monotone" dataKey="budget" stroke="#5b7cfa" strokeWidth={3} dot={{ r: 4, strokeWidth: 2 }} activeDot={{ r: 6 }} name="Budget" />
                    <Line type="monotone" dataKey="actual" stroke="#26c6da" strokeWidth={3} dot={{ r: 4, strokeWidth: 2 }} activeDot={{ r: 6 }} name="Actual" />
                </LineChart>
            </ResponsiveContainer>
        </Box>
    );
};

export default BudgetChart;
