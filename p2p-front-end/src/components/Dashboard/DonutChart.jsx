import React from 'react';
import { Paper, Typography, Box } from '@mui/material';
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from 'recharts';

export const COLORS = ['#6becdf', '#4accdb', '#2b82a3', '#1e88e5', '#90caf9'];

const DonutChart = ({ data, showLegend = true }) => {
    // Ensure all empty string names are converted to 'Unknown' BEFORE Recharts reads them,
    // otherwise Recharts Legend defaults to the dataKey ("value").
    const safeData = (data || []).map(item => ({
        ...item,
        name: item.name ? item.name : 'Unknown Account'
    }));

    // Graceful fallback for completely empty data (e.g., no spending at all)
    if (safeData.length === 0) {
        return (
            <Box display="flex" justifyContent="center" alignItems="center" height="100%">
                <Typography variant="body2" color="text.secondary">
                    No expenses data available
                </Typography>
            </Box>
        );
    }

    const total = safeData.reduce((sum, entry) => sum + entry.value, 0);

    // If data exists but all values are 0, Recharts Pie will silently fail to render wedges.
    if (total === 0) {
        return (
            <Box display="flex" justifyContent="center" alignItems="center" height="100%">
                <Typography variant="body2" color="text.secondary">
                    No spending recorded for this period (Amounts are 0)
                </Typography>
            </Box>
        );
    }

    return (
        <Box sx={{ width: '100%', height: '100%', minHeight: 250 }}>
            <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                    <Pie
                        data={safeData}
                        cx="50%"
                        cy="45%"
                        innerRadius="60%"
                        outerRadius="85%"
                        paddingAngle={2}
                        dataKey="value"
                        nameKey="name"
                        labelLine={false}
                        isAnimationActive={false} // Disable animation to prevent visual glitches on re-render
                    >
                        {safeData.map((entry, index) => {
                            // Ensure empty names are categorized as Unknown for the legend and tooltip
                            const sliceName = entry.name || 'Unknown Account';
                            return (
                                <Cell key={`cell-${index}`} name={sliceName} fill={COLORS[index % COLORS.length]} />
                            );
                        })}
                    </Pie>
                    <Tooltip
                        formatter={(value) => {
                            const percent = ((value / total) * 100).toFixed(1);
                            return `${percent}%`;
                        }}
                        contentStyle={{ borderRadius: 12, border: 'none', boxShadow: '0 8px 24px rgba(0,0,0,0.1)' }}
                    />
                </PieChart>
            </ResponsiveContainer>
        </Box>
    );
};

export default DonutChart;
