import React, { useState } from 'react';
import {
    Paper,
    Typography,
    Box,
    TableContainer,
    Table,
    TableHead,
    TableRow,
    TableCell,
    TableBody,
    Skeleton,
    TablePagination,
    IconButton,
    Dialog,
    DialogContent,
    Popover,
    TextField,
    Button
} from '@mui/material';
import FullscreenIcon from '@mui/icons-material/Fullscreen';
import CloseIcon from '@mui/icons-material/Close';
import FilterListIcon from '@mui/icons-material/FilterList';
import DownloadIcon from '@mui/icons-material/Download';

const ActualTable = React.memo(({
    loading,
    data = [],
    dateFilter,
    onDateFilterChange,
    page,
    rowsPerPage,
    totalCount,
    onPageChange,
    onRowsPerPageChange
}) => {
    const [openFullScreen, setOpenFullScreen] = useState(false);

    // Filter Popover State
    const [anchorEl, setAnchorEl] = useState(null);
    const [tempDateFilter, setTempDateFilter] = useState({ startDate: '', endDate: '' });

    const handleFilterClick = (event) => {
        setTempDateFilter(dateFilter || { startDate: '', endDate: '' });
        setAnchorEl(event.currentTarget);
    };

    const handleFilterClose = () => {
        setAnchorEl(null);
    };

    const handleApplyFilter = () => {
        onDateFilterChange(tempDateFilter);
        setAnchorEl(null);
    };

    const handleClearFilter = () => {
        setTempDateFilter({ startDate: '', endDate: '' });
        onDateFilterChange({ startDate: '', endDate: '' });
        setAnchorEl(null);
    };

    const handleChangePage = (event, newPage) => {
        onPageChange(newPage);
    };

    const handleChangeRowsPerPage = (event) => {
        onRowsPerPageChange(parseInt(event.target.value, 10));
    };

    // Data is now paginated from server
    const paginatedData = data;

    const renderTable = (isMaximized = false) => (
        <React.Fragment>
            <TableContainer sx={{ flexGrow: 1, width: '100%', overflow: 'auto', border: '1px solid #e0e0e0', borderRadius: '4px' }}>
                <Table stickyHeader size={isMaximized ? "medium" : "small"} sx={{ minWidth: '100%' }}>
                    <TableHead>
                        <TableRow>
                            <TableCell sx={{ bgcolor: '#0388d1', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>GL Code</TableCell>
                            <TableCell sx={{ bgcolor: '#0388d1', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>GL Name</TableCell>
                            <TableCell sx={{ bgcolor: '#0388d1', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Doc No.</TableCell>
                            <TableCell align="right" sx={{ bgcolor: '#0388d1', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Amount</TableCell>
                            <TableCell sx={{ bgcolor: '#0388d1', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Vendor</TableCell>
                            <TableCell sx={{ bgcolor: '#0388d1', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)' }}>Description</TableCell>
                            <TableCell sx={{ bgcolor: '#0388d1', color: 'white', fontWeight: 'bold' }}>Date</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {loading ? (
                            Array.from(new Array(10)).map((_, index) => (
                                <TableRow key={index}>
                                    <TableCell><Skeleton variant="text" /></TableCell>
                                    <TableCell><Skeleton variant="text" /></TableCell>
                                    <TableCell><Skeleton variant="text" /></TableCell>
                                    <TableCell><Skeleton variant="text" /></TableCell>
                                    <TableCell><Skeleton variant="text" /></TableCell>
                                    <TableCell><Skeleton variant="text" /></TableCell>
                                    <TableCell><Skeleton variant="text" /></TableCell>
                                </TableRow>
                            ))
                        ) : paginatedData.length > 0 ? (
                            paginatedData.map((row, index) => (
                                <TableRow key={index} hover>
                                    <TableCell sx={{ borderRight: '1px solid #e0e0e0' }}>{row.gl_account_no}</TableCell>
                                    <TableCell sx={{ borderRight: '1px solid #e0e0e0' }}>{row.gl_account_name}</TableCell>
                                    <TableCell sx={{ borderRight: '1px solid #e0e0e0' }}>{row.document_no}</TableCell>
                                    <TableCell align="right" sx={{ fontWeight: 'bold', color: parseFloat(row.amount) < 0 ? 'red' : 'green', borderRight: '1px solid #e0e0e0' }}>
                                        {parseFloat(row.amount || 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                                    </TableCell>
                                    <TableCell sx={{ borderRight: '1px solid #e0e0e0' }}>{row.vendor || "-"}</TableCell>
                                    <TableCell sx={{ borderRight: '1px solid #e0e0e0' }}>{row.description}</TableCell>
                                    <TableCell>{row.posting_date}</TableCell>
                                </TableRow>
                            ))
                        ) : (
                            <TableRow>
                                <TableCell colSpan={7} align="center" sx={{ py: 5, color: 'text.secondary' }}>
                                    No actual data found for selected filters
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </TableContainer>
            {/* Pagination Control */}
            {!loading && data.length > 0 && (
                <Box sx={{ flexShrink: 0, borderTop: '1px solid #e0e0e0' }}>
                    <TablePagination
                        rowsPerPageOptions={[10, 25, 50, 100]}
                        component="div"
                        count={totalCount}
                        rowsPerPage={rowsPerPage}
                        page={page}
                        onPageChange={handleChangePage}
                        onRowsPerPageChange={handleChangeRowsPerPage}
                        sx={{
                            '.MuiTablePagination-toolbar': { minHeight: '36px', px: 1 },
                            '.MuiTablePagination-selectLabel, .MuiTablePagination-input, .MuiTablePagination-displayedRows': { fontSize: '0.75rem' }
                        }}
                    />
                </Box>
            )}
        </React.Fragment>
    );

    const openFilter = Boolean(anchorEl);

    return (
        <Paper sx={{
            p: 1.5,
            pt: 1,
            display: 'flex',
            flexDirection: 'column',
            borderRadius: 2,
            mb: 2,
            flex: '1 1 0', // Equal height split
            minHeight: { xs: '400px', md: 0 },
            overflow: 'hidden'
        }}>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2, px: 1 }}>
                <Box sx={{ bgcolor: '#0388d1', px: 2, py: 0.5, borderRadius: '8px' }}>
                    <Typography variant="h6" sx={{ color: 'white', fontWeight: 'bold' }}>
                        Actual Detail
                    </Typography>
                </Box>
                <Box>
                    <Box sx={{ display: 'flex', gap: 1 }}>
                    <IconButton onClick={handleFilterClick} size="small" sx={{ color: '#424242'}}>
                        <FilterListIcon />
                    </IconButton>
                    <IconButton onClick={() => setOpenFullScreen(true)} size="small" sx={{ color: '#424242' }}>
                        <FullscreenIcon />
                    </IconButton>
                    <IconButton size="small" sx={{ color: '#424242' }}>
                    <DownloadIcon sx={{ fontSize: 18 }} />
                </IconButton>
                </Box>
                </Box>
            </Box>

            {/* Filter Popover */}
            <Popover
                open={openFilter}
                anchorEl={anchorEl}
                onClose={handleFilterClose}
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'right',
                }}
                transformOrigin={{
                    vertical: 'top',
                    horizontal: 'right',
                }}
            >
                <Box sx={{ p: 2, display: 'flex', flexDirection: 'column', gap: 2, minWidth: 300 }}>
                    <Typography variant="subtitle2" fontWeight="bold">Filter Date</Typography>
                    <TextField
                        label="Start Date"
                        type="date"
                        size="small"
                        InputLabelProps={{ shrink: true }}
                        value={tempDateFilter.startDate}
                        onChange={(e) => setTempDateFilter({ ...tempDateFilter, startDate: e.target.value })}
                        fullWidth
                    />
                    <TextField
                        label="End Date"
                        type="date"
                        size="small"
                        InputLabelProps={{ shrink: true }}
                        value={tempDateFilter.endDate}
                        onChange={(e) => setTempDateFilter({ ...tempDateFilter, endDate: e.target.value })}
                        fullWidth
                    />
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 1 }}>
                        <Button onClick={handleClearFilter} size="small" color="inherit">
                            Clear
                        </Button>
                        <Button onClick={handleApplyFilter} size="small" variant="contained" color="primary">
                            Apply
                        </Button>
                    </Box>
                </Box>
            </Popover>

            {renderTable(false)}

            {/* Fullscreen Dialog */}
            <Dialog
                open={openFullScreen}
                onClose={() => setOpenFullScreen(false)}
                fullScreen
            >
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', bgcolor: '#0388d1', p: 2 }}>
                    <Typography variant="h6" sx={{ color: 'white', fontWeight: 'bold' }}>
                        Actual Detail (Full Screen)
                    </Typography>
                    <Box>
                        {/* We can show filter here too if needed, but per request it's next to expand button on main view? 
                             Let's assume users might want to filter inside full screen too. I'll clone the button here. */}
                        <IconButton onClick={handleFilterClick} sx={{ color: 'white', mr: 2 }}>
                            <FilterListIcon />
                        </IconButton>
                        <IconButton onClick={() => setOpenFullScreen(false)} sx={{ color: 'white' }}>
                            <CloseIcon />
                        </IconButton>
                    </Box>
                </Box>
                <DialogContent sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
                    {renderTable(true)}
                </DialogContent>
            </Dialog>
        </Paper>
    );
});

export default ActualTable;
