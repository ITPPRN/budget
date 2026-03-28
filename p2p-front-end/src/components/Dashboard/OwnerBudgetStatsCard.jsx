import React from 'react';
import { Paper, Typography, Box, Stack, LinearProgress, Chip } from '@mui/material';
import Inventory2OutlinedIcon from '@mui/icons-material/Inventory2Outlined';
import LinkIcon from '@mui/icons-material/Link';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import LocalOfferIcon from '@mui/icons-material/LocalOffer';

// --- Styles ---
const blueCardStyle = {
    p: 2.5,
    borderRadius: 4,
    background: '#5b7cfa', // Matches the Blue in image
    color: 'white',
    width: '100%', // Ensure card fills the grid item
    height: 140,
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-between',
    boxShadow: '0 4px 12px rgba(91, 124, 250, 0.3)',
    position: 'relative',
    overflow: 'hidden'
};

const formatMB = (val) => {
    if (!val) return "";
    const safeVal = Math.abs(val);
    const mb = safeVal / 1000000;
    // Truncate to 2 decimal places (no rounding)
    const truncated = Math.floor(mb * 100) / 100;
    return `${truncated.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} MB`;
};

// Generic Blue Card Info
const InfoCard = ({ title, value, icon: Icon }) => (
    <Paper elevation={0} sx={blueCardStyle}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 1 }}>
            <Box sx={{ p: 0.5, bgcolor: 'rgba(255,255,255,0.2)', borderRadius: 1, display: 'flex' }}>
                <Icon sx={{ fontSize: 20 }} />
            </Box>
            <Typography variant="body1" sx={{ fontWeight: 500, lineHeight: 1.2 }}>
                {title}
            </Typography>
        </Box>
        <Typography variant="h4" sx={{ fontWeight: 'bold', mb: 1 }}>
            {formatMB(value)}
        </Typography>
    </Paper>
);

// 1. Approved Expense Budget
export const TotalBudgetCard = ({ totalBudget = 0 }) => (
    <InfoCard title="Approved Expense Budget" value={totalBudget} icon={Inventory2OutlinedIcon} />
);

// 2. Actual Spending
export const ActualCard = ({ totalActual = 0 }) => (
    <InfoCard title="Actual Spending" value={totalActual} icon={LocalOfferIcon} />
);

// 3. Remaining Expense Budget
export const RemainingBudgetCard = ({ remaining = 0 }) => (
    <InfoCard title="Remaining Expense Budget" value={remaining} icon={LinkIcon} />
);

// 4. Usage & Status Widget (Right Side)
export const UsageStatusWidget = ({ usagePercent = 0, status = "In Budget" }) => {
    // Usage Color
    const getUsageColor = (p) => {
        if (p > 100) return '#ff4d4d'; // Red
        if (p > 80) return '#ffca28';  // Yellow
        return '#5b7cfa'; // Blue
    };

    // Status Color (Badge)
    const getStatusColor = (s) => {
        if (s === 'Over Budget') return '#d32f2f';
        if (s === 'Near Limit') return '#fbc02d';
        return '#388e3c'; // Green
    };

    return (
        <Paper elevation={0} sx={{ height: 140, p: 2, borderRadius: 4, display: 'flex', flexDirection: 'column', justifyContent: 'center', gap: 2 }}>
            {/* Usage Bar Row */}
            <Stack direction="row" alignItems="center" spacing={2} sx={{ width: '100%' }}>
                <Typography variant="body2" fontWeight="bold" color="textSecondary" sx={{ minWidth: 50 }}>Usage</Typography>
                <Box sx={{ flexGrow: 1 }}>
                    <LinearProgress
                        variant="determinate"
                        value={Math.min(usagePercent, 100)}
                        sx={{
                            height: 10,
                            borderRadius: 5,
                            bgcolor: '#f0f2f5',
                            '& .MuiLinearProgress-bar': {
                                bgcolor: getUsageColor(usagePercent),
                                borderRadius: 5
                            }
                        }}
                    />
                </Box>
                <Typography variant="body2" fontWeight="bold" sx={{ minWidth: 40, textAlign: 'right' }}>{Math.round(usagePercent)}%</Typography>
            </Stack>

            {/* Status Badge Row */}
            <Stack direction="row" alignItems="center" spacing={2} sx={{ width: '100%' }}>
                <Typography variant="body2" fontWeight="bold" color="textSecondary" sx={{ minWidth: 50 }}>Status</Typography>
                <Chip
                    label={status}
                    sx={{
                        bgcolor: getStatusColor(status),
                        color: 'white',
                        fontWeight: 'bold',
                        borderRadius: 2,
                        height: 28,
                        px: 1
                    }}
                />
            </Stack>
        </Paper>
    );
};

// 5. CAPEX Card (Teal/Blue Bottom Right)
export const CapexCard = ({ budget = 0, actual = 0 }) => {
    return (
        <Paper elevation={0} sx={{
            p: 2.5,
            borderRadius: 4,
            background: '#2499ef', // Lighter Blue/Teal
            color: 'white',
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center',
            boxShadow: '0 4px 12px rgba(36, 153, 239, 0.3)',
            height: '100%',
            width: '100%',
            minHeight: 120
        }}>
            <Stack direction="row" justifyContent="space-between" alignItems="flex-start">
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                    <Inventory2OutlinedIcon sx={{ fontSize: 20 }} />
                    <Typography variant="body1" fontWeight={500}>CAPEX Budget</Typography>
                </Box>
                <LinkIcon sx={{ fontSize: 20, transform: 'rotate(-45deg)', cursor: 'pointer', opacity: 0.8 }} /> {/* Mock Download */}
            </Stack>

            <Typography variant="h4" fontWeight="bold">
                {budget > 0 ? formatMB(budget) : "-"}
            </Typography>
            <Typography variant="body2" sx={{ opacity: 0.9, textAlign: 'right', mt: 1 }}>
                {actual ? `of ${formatMB(actual)} (Total)` : ""}
            </Typography>
        </Paper>
    );
};

export default TotalBudgetCard;