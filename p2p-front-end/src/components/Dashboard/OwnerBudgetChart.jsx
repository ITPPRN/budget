import React from 'react';
import { Paper, Typography, Box } from '@mui/material';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const BudgetChart = ({ data }) => {
    return (
        <Box sx={{ width: '100%', height: '100%', minHeight: 400 }}>
            <ResponsiveContainer width="100%" height="100%">
                <LineChart
                    data={data}
                    margin={{
                        top: 25,
                        right: 30,
                        left: 20,
                        bottom: 35,
                    }}
                >
                    <CartesianGrid
                        strokeDasharray="4 4"
                        stroke="#e2e8f0"
                        vertical={true}
                        horizontal={true}
                    />
                    <XAxis
                        dataKey="name"
                        axisLine={{ stroke: '#cbd5e1', strokeWidth: 1 }}
                        tickLine={{ stroke: '#cbd5e1' }}
                        tick={{ fill: '#64748b', fontSize: 11, fontWeight: 700 }}
                        dy={10}
                        padding={{ left: 20, right: 20 }}
                    />
                    <YAxis
                        width={60}
                        axisLine={{ stroke: '#cbd5e1', strokeWidth: 1 }}
                        tickLine={{ stroke: '#cbd5e1' }}
                        tick={{ fill: '#64748b', fontSize: 11, fontWeight: 600 }}
                        tickFormatter={(value) => {
                            if (value === 0) return '0';
                            return `${(value / 1000000).toFixed(1)}M`;
                        }}
                    />
                    <Tooltip
                        contentStyle={{ borderRadius: 12, border: 'none', boxShadow: '0 8px 24px rgba(0,0,0,0.1)', padding: '12px' }}
                        itemStyle={{ fontWeight: 700 }}
                    />
                    <Line
                        type="monotone"
                        dataKey="budget"
                        stroke="#4d6eff"
                        strokeWidth={3}
                        dot={{ r: 5, fill: '#fff', stroke: '#4d6eff', strokeWidth: 2 }}
                        activeDot={{ r: 7, fill: '#4d6eff', stroke: '#fff', strokeWidth: 2 }}
                        name="Budget"
                        connectNulls
                    />
                    <Line
                        type="monotone"
                        dataKey="actual"
                        stroke="#64b5f6"
                        strokeWidth={3}
                        dot={{ r: 5, fill: '#fff', stroke: '#64b5f6', strokeWidth: 2 }}
                        activeDot={{ r: 7, fill: '#64b5f6', stroke: '#fff', strokeWidth: 2 }}
                        name="Actual"
                        connectNulls
                    />
                </LineChart>
            </ResponsiveContainer>
        </Box>
    );
};

export default BudgetChart;
