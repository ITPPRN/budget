import React from 'react';
import { Paper, Typography, Box } from '@mui/material';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';

const COLORS = ['#4d6eff', '#64b5f6', '#10254a', '#1e88e5', '#90caf9'];

const DonutChart = ({ data }) => {
    const total = data.reduce((sum, entry) => sum + entry.value, 0);

    const renderCustomizedLabel = ({ cx, cy, midAngle, innerRadius, outerRadius, value, name }) => {
        const RADIAN = Math.PI / 180;
        const radius = outerRadius * 1.3;
        const x = cx + radius * Math.cos(-midAngle * RADIAN);
        const y = cy + radius * Math.sin(-midAngle * RADIAN);
        const percent = ((value / total) * 100).toFixed(1);

        return (
            <text
                x={x}
                y={y}
                fill="#333"
                textAnchor={x > cx ? 'start' : 'end'}
                dominantBaseline="central"
                fontSize="12"
                fontWeight="700"
            >
                <tspan x={x} dy="-0.6em">{name}</tspan>
                <tspan x={x} dy="1.2em" fill="#666" fontWeight="500">{`${percent}%`}</tspan>
            </text>
        );
    };

    return (
        <Box sx={{ width: '100%', height: '100%', minHeight: 250 }}>
            <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                    <Pie
                        data={data}
                        cx="50%"
                        cy="45%"
                        innerRadius={55}
                        outerRadius={75}
                        paddingAngle={2}
                        dataKey="value"
                        nameKey="name"
                        label={renderCustomizedLabel}
                        labelLine={{ stroke: '#cfd8dc', strokeWidth: 1 }}
                    >
                        {data.map((entry, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                    </Pie>
                    <Tooltip
                        formatter={(value) => `${new Intl.NumberFormat('en-US').format(value)}`}
                        contentStyle={{ borderRadius: 12, border: 'none', boxShadow: '0 8px 24px rgba(0,0,0,0.1)' }}
                    />
                </PieChart>
            </ResponsiveContainer>
        </Box>
    );
};

export default DonutChart;
