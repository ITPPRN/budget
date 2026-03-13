import React from 'react';
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
    IconButton,
    Dialog,
    DialogContent,
    TablePagination
} from '@mui/material';
import FullscreenIcon from '@mui/icons-material/Fullscreen';
import CloseIcon from '@mui/icons-material/Close';
import DownloadIcon from '@mui/icons-material/Download';

const months = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];

const BudgetTable = React.memo(({ loading, data, selectedCount }) => {
    const [openFullScreen, setOpenFullScreen] = React.useState(false);
    const [page, setPage] = React.useState(0);
    const [rowsPerPage, setRowsPerPage] = React.useState(10);

    const handleChangePage = (event, newPage) => {
        setPage(newPage);
    };

    const handleChangeRowsPerPage = (event) => {
        setRowsPerPage(parseInt(event.target.value, 10));
        setPage(0);
    };

    // Slice data for pagination
    const paginatedData = data.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage);

    const renderTable = (isMaximized = false) => (
        <React.Fragment>
            <TableContainer sx={{ flexGrow: 1, width: '100%', overflow: 'auto', border: '1px solid #e0e0e0', borderRadius: '4px' }}>
                <Table stickyHeader size={isMaximized ? "medium" : "small"} sx={{ minWidth: '100%' }}>
                    <TableHead>
                        <TableRow>
                            <TableCell sx={{ bgcolor: '#5271ff', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)', whiteSpace: 'nowrap', position: 'sticky', left: 0, zIndex: 3, minWidth: 100 }}>GL Code</TableCell>
                            <TableCell sx={{ bgcolor: '#5271ff', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)', minWidth: 200, whiteSpace: 'nowrap', position: 'sticky', left: 100, zIndex: 3 }}>GL Name</TableCell>
                            {months.map(m => (
                                <TableCell key={m} align="right" sx={{ bgcolor: '#5271ff', color: 'white', fontWeight: 'bold', borderRight: '1px solid rgba(255,255,255,0.3)', whiteSpace: 'nowrap' }}>{m}</TableCell>
                            ))}
                            <TableCell align="right" sx={{ bgcolor: '#5271ff', color: 'white', fontWeight: 'bold', whiteSpace: 'nowrap' }}>Total</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {loading ? (
                            Array.from(new Array(10)).map((_, index) => (
                                <TableRow key={index}>
                                    <TableCell sx={{ position: 'sticky', left: 0, bgcolor: 'background.paper', zIndex: 1, minWidth: 100 }}><Skeleton variant="text" width={80} /></TableCell>
                                    <TableCell sx={{ position: 'sticky', left: 100, bgcolor: 'background.paper', zIndex: 1 }}><Skeleton variant="text" width={150} /></TableCell>
                                    {months.map(m => (
                                        <TableCell key={m}><Skeleton variant="text" /></TableCell>
                                    ))}
                                    <TableCell><Skeleton variant="text" width={60} /></TableCell>
                                </TableRow>
                            ))
                        ) : paginatedData.length > 0 ? (
                            paginatedData.map((row) => {
                                const amountMap = {};
                                row.budget_amounts?.forEach(a => {
                                    amountMap[a.month] = a.amount;
                                });

                                return (
                                    <TableRow key={row.conso_gl} hover>
                                        <TableCell sx={{ whiteSpace: 'nowrap', borderRight: '1px solid #e0e0e0', position: 'sticky', left: 0, bgcolor: 'background.paper', zIndex: 1, minWidth: 100 }}>{row.conso_gl || "-"}</TableCell>
                                        <TableCell sx={{ whiteSpace: 'nowrap', borderRight: '1px solid #e0e0e0', position: 'sticky', left: 100, bgcolor: 'background.paper', zIndex: 1 }}>{row.gl_name}</TableCell>
                                        {months.map(m => (
                                            <TableCell key={m} align="right" sx={{ whiteSpace: 'nowrap', borderRight: '1px solid #e0e0e0' }}>
                                                {parseFloat(amountMap[m] || 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                                            </TableCell>
                                        ))}
                                        <TableCell align="right" sx={{ fontWeight: 'bold', bgcolor: '#f5f5f5', whiteSpace: 'nowrap' }}>
                                            {parseFloat(row.year_total || 0).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                                        </TableCell>
                                    </TableRow>
                                );
                            })
                        ) : (
                            <TableRow>
                                <TableCell colSpan={15} align="center" sx={{ py: 5, color: 'text.secondary' }}>
                                    {selectedCount === 0 ? "Select items from the Filter Pane to view details" : "No data found for selected filters"}
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
                        count={data.length}
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

    return (
        <Paper sx={{
            p: 1.5,
            pt: 1,
            display: 'flex',
            flexDirection: 'column',
            borderRadius: 2,
            mb: 2,
            flex: '1 1 0', // Equal height split
            minHeight: { xs: '400px', md: 0 }, // Mobile gets fixed minHeight, Desktop fits available space
            overflow: 'hidden'
        }}>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2, px: 1 }}>
                <Box sx={{ bgcolor: '#5271ff', px: 2, py: 0.5, borderRadius: '8px' }}>
                    <Typography variant="h6" sx={{ color: 'white', fontWeight: 'bold' }}>
                        Budget Detail
                    </Typography>
                </Box>
                <Box sx={{ display: 'flex', gap: 1 }}>
                <IconButton onClick={() => setOpenFullScreen(true)} size="small" sx={{ color: '#424242' }}>
                    <FullscreenIcon />
                </IconButton>
                <IconButton size="small" sx={{ color: '#424242' }}>
                    <DownloadIcon sx={{ fontSize: 18 }} />
                </IconButton>
                </Box>
            </Box>

            <Box sx={{ flexGrow: 1, overflow: 'hidden', width: '100%', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
                {renderTable(false)}
            </Box>

            {/* Fullscreen Dialog */}
            <Dialog
                open={openFullScreen}
                onClose={() => setOpenFullScreen(false)}
                fullScreen
            >
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', bgcolor: '#5271ff', p: 2 }}>
                    <Typography variant="h6" sx={{ color: 'white', fontWeight: 'bold' }}>
                        Budget Detail (Full Screen)
                    </Typography>
                    <IconButton onClick={() => setOpenFullScreen(false)} sx={{ color: 'white' }}>
                        <CloseIcon />
                    </IconButton>
                </Box>
                <DialogContent sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
                    {renderTable(true)}
                </DialogContent>
            </Dialog>
        </Paper>
    );
});

export default BudgetTable;
