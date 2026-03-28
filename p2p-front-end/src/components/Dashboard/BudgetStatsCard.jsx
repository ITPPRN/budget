import React from 'react';
import { Paper, Typography, Box, CircularProgress, Stack, LinearProgress } from '@mui/material';
import Inventory2OutlinedIcon from '@mui/icons-material/Inventory2Outlined';
import LinkIcon from '@mui/icons-material/Link';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';

// Colors
const COLORS = {
    // Soft Blue Gradient Base
    cardBg: '#4facfe',
    progressTrack: 'rgba(255,255,255,0.2)',
    progressValue: '#00e5ff',
    textMain: '#ffffff',
    textLabel: 'rgba(255, 255, 255, 0.9)' // High contrast
};

const formatMB = (val) => {
    const safeVal = val || 0;
    const mb = safeVal / 1000000;
    // Truncate to 2 decimal places (Strict No-Rounding)
    const truncated = Math.trunc(mb * 100) / 100;
    return `${truncated.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} MB`;
};

// Common Style for Paper
const cardStyle = {
    p: 2.5,
    borderRadius: 5,
    // Soft Royal Blue Gradient
    background: 'linear-gradient(135deg, #5b7cfa 0%, #3e64f0 100%)',
    color: 'white',
    width: '100%',
    height: 130,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    boxShadow: '0 6px 16px rgba(0,0,0,0.1)'
};

// 1. Total Organization Budget Card
export const TotalBudgetCard = ({ totalBudget = 0, totalActual = 0 }) => {
    // Percentage uses absolute values to handle negative budgets correctly
    const percentage = Math.abs(totalBudget) > 0 ? (totalActual / totalBudget) * 100 : 0;
    const safePercentage = Math.max(0, Math.min(percentage, 100)); // Still cap at 0-100 for circle

    return (
        <Paper elevation={0} sx={cardStyle}>
            <Stack spacing={0} sx={{ flex: 1 }}>
                <Stack direction="row" alignItems="center" spacing={1} sx={{ opacity: 0.9, mb: 0.5 }}>
                    <Inventory2OutlinedIcon sx={{ fontSize: 18 }} />
                    <Typography variant="subtitle2" sx={{ fontWeight: 700, letterSpacing: 0.5, fontSize: '0.85rem' }}>
                        Total Organization Budget
                    </Typography>
                </Stack>

                <Typography variant="h4" sx={{ fontWeight: 800, color: COLORS.textMain, lineHeight: 1.2 }}>
                    {formatMB(totalBudget)}
                </Typography>

                <Typography variant="body2" sx={{ fontWeight: 500, color: COLORS.textLabel, mt: 0.5 }}>
                    Actual: {formatMB(totalActual)}
                </Typography>
            </Stack>

            <Box sx={{ position: 'relative', display: 'flex', ml: 2 }}>
                <CircularProgress
                    variant="determinate"
                    value={100}
                    size={70} // Reduced size
                    thickness={5}
                    sx={{ color: COLORS.progressTrack }}
                />
                <CircularProgress
                    variant="determinate"
                    value={safePercentage}
                    size={70} // Reduced size
                    thickness={5}
                    sx={{
                        color: COLORS.progressValue,
                        position: 'absolute',
                        left: 0,
                        strokeLinecap: 'round',
                        [`& .MuiCircularProgress-circle`]: { strokeLinecap: 'round' },
                    }}
                />
                <Box sx={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                    <Typography variant="subtitle1" sx={{ fontWeight: 800, color: COLORS.textMain }}>
                        {`${Math.round(percentage)}%`}
                    </Typography>
                </Box>
            </Box>
        </Paper>
    );
};

// 2. Remaining Budget Card
export const RemainingBudgetCard = ({ totalBudget = 331.46, totalActual = 0 }) => {
    const remaining = totalBudget - totalActual;

    return (
        <Paper elevation={0} sx={cardStyle}>
            <Stack spacing={0} sx={{ flex: 1 }}>
                <Stack direction="row" alignItems="center" spacing={1} sx={{ opacity: 0.9, mb: 0.5 }}>
                    <LinkIcon sx={{ fontSize: 18, transform: 'rotate(-45deg)' }} />
                    <Typography variant="subtitle2" sx={{ fontWeight: 700, letterSpacing: 0.5, fontSize: '0.85rem' }}>
                        Remaining Budget
                    </Typography>
                </Stack>

                <Typography variant="h4" sx={{ fontWeight: 800, color: COLORS.textMain, lineHeight: 1.2 }}>
                    {formatMB(remaining)}
                </Typography>
            </Stack>

            <Box sx={{
                position: 'relative',
                width: 70, height: 70, // Reduced size
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                bgcolor: 'rgba(255,255,255,0.15)',
                borderRadius: '50%',
                ml: 2
            }}>
                <Box sx={{
                    width: 45, height: 45, // Reduced size
                    bgcolor: 'rgba(255,255,255,0.1)',
                    borderRadius: '50%',
                    display: 'flex', alignItems: 'center', justifyContent: 'center'
                }}>
                    <CheckCircleIcon sx={{ fontSize: 32, color: COLORS.progressValue }} />
                </Box>
            </Box>
        </Paper>
    );
};

// 3. Department Status Alert Card
export const DepartmentAlertCard = ({ overBudgetCount = 0, nearLimitCount = 0 }) => {
    return (
        <Paper elevation={0} sx={{
            ...cardStyle,
            // background: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)', // Lighter Blue/Cyan variant for distinction
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center',
            alignItems: 'flex-start',
            px: 3
        }}>
            <Stack direction="row" alignItems="center" spacing={1} sx={{ mb: 1.5 }}>
                <Inventory2OutlinedIcon sx={{ fontSize: 20 }} /> {/* Reusing icon or import LocalFlorist if available */}
                <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
                    Department Status Alert
                </Typography>
            </Stack>

            <Stack spacing={1} sx={{ width: '100%' }}>
                {/* Over Budget Row */}
                <Stack direction="row" alignItems="center" spacing={1.5}>
                    <Box sx={{
                        bgcolor: '#ff4d4d',
                        borderRadius: '4px',
                        width: 24, height: 24,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
                    }}>
                        <Typography sx={{ fontWeight: 'bold', fontSize: '14px', color: 'white' }}>!</Typography>
                    </Box>
                    <Typography variant="body1" sx={{ color: "white" }}>
                        {overBudgetCount} Dep - Over BG
                    </Typography>
                </Stack>

                {/* Near Limit Row */}
                <Stack direction="row" alignItems="center" spacing={1.5}>
                    <Box sx={{
                        bgcolor: '#ffc107',
                        borderRadius: '4px',
                        width: 24, height: 24,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        boxShadow: '0 2px 4px rgba(0,0,0,0.2)'
                    }}>
                        <Typography sx={{ fontWeight: 'bold', fontSize: '14px', color: 'black' }}>!</Typography>
                    </Box>
                    <Typography variant="body1" sx={{ color: "white" }}>
                        {nearLimitCount} Dep - Near Limit
                    </Typography>
                </Stack>
            </Stack>
        </Paper>
    );
};

// 4. CAPEX Card (Optional - Keeping for safety)
export const CapexCard = ({ budget = 0, actual = 0 }) => {
    return (<Box>Capex Card</Box>);
};

export default TotalBudgetCard;