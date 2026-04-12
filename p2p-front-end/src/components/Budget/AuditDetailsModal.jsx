import React, { useState, useEffect } from 'react';
import {
    Dialog,
    DialogTitle,
    DialogContent,
    DialogActions,
    Button,
    Table,
    TableHead,
    TableRow,
    TableCell,
    TableBody,
    Typography,
    Box,
    TableContainer,
    CircularProgress
} from '@mui/material';
import VisibilityIcon from '@mui/icons-material/Visibility';
import FileDownloadIcon from '@mui/icons-material/FileDownload';
import api from '../../utils/api/axiosInstance';

const AuditDetailsModal = ({ open, onClose, log }) => {
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (open && log?.id) {
            const fetchItems = async () => {
                setLoading(true);
                try {
                    const res = await api.get(`/budgets/audit/logs/${log.id}/items`);
                    setItems(res.data || []);
                } catch (err) {
                    console.error("Fetch Rejected Items Error:", err);
                    setItems([]);
                } finally {
                    setLoading(false);
                }
            };
            fetchItems();
        } else {
            setItems([]);
        }
    }, [open, log]);

    const handleExportExcel = () => {
        if (!items || items.length === 0) return;

        const escape = (v) => {
            const s = v === null || v === undefined ? '' : String(v);
            return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        };

        const dept = log.department_code || log.department || '';
        const owner = log.owner_name || log.created_by || '';
        const period = `${log.month}/${log.year}`;
        const generatedAt = new Date().toLocaleString('th-TH');

        const headers = [
            { label: 'GL Code', width: 110 },
            { label: 'Account Name', width: 240 },
            { label: 'Doc No.', width: 130 },
            { label: 'Amount', width: 140 },
            { label: 'Vendor', width: 200 },
            { label: 'Description', width: 320 },
            { label: 'Date', width: 110 },
        ];

        const fmtMoney = (n) =>
            Number(n).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

        const colgroupHtml = headers
            .map((h) => `<col style="width:${h.width}px">`)
            .join('');

        const headerHtml = headers
            .map(
                (h) =>
                    `<th style="background:#043478;color:#ffffff;font-weight:bold;border:1px solid #1a3c70;padding:8px 10px;text-align:${
                        h.label === 'Amount' ? 'right' : 'left'
                    };font-size:12px;">${escape(h.label)}</th>`
            )
            .join('');

        const rowsHtml = items
            .map((row, idx) => {
                const amount = parseFloat(row.amount) || 0;
                const zebra = idx % 2 === 0 ? '#ffffff' : '#f6f8fb';
                const amountColor = amount < 0 ? '#c62828' : '#2e7d32';
                const cellBase = `border:1px solid #d9dee5;padding:6px 10px;font-size:11px;background:${zebra};vertical-align:top;`;
                return `<tr>
                    <td style="${cellBase}">${escape(row.conso_gl)}</td>
                    <td style="${cellBase}">${escape(row.gl_account_name)}</td>
                    <td style="${cellBase}">${escape(row.doc_no)}</td>
                    <td style="${cellBase}text-align:right;font-weight:bold;color:${amountColor};mso-number-format:'#,##0.00';">${fmtMoney(amount)}</td>
                    <td style="${cellBase}">${escape(row.vendor || '-')}</td>
                    <td style="${cellBase}">${escape(row.description)}</td>
                    <td style="${cellBase}mso-number-format:'@';">${escape(row.posting_date)}</td>
                </tr>`;
            })
            .join('');

        const titleBar = `<tr>
            <td colspan="7" style="background:#043478;color:#ffffff;padding:14px 12px;font-size:16px;font-weight:bold;border:1px solid #1a3c70;">
                Rejected Items Report
            </td>
        </tr>`;

        const metaRow = (label, value) =>
            `<tr>
                <td style="background:#f6f8fb;border:1px solid #d9dee5;padding:6px 10px;font-weight:bold;color:#6b7280;font-size:11px;width:110px;">${escape(label)}</td>
                <td colspan="6" style="background:#ffffff;border:1px solid #d9dee5;padding:6px 10px;color:#111827;font-size:12px;font-weight:bold;">${escape(value)}</td>
            </tr>`;

        const html = `<html xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:x="urn:schemas-microsoft-com:office:excel" xmlns="http://www.w3.org/TR/REC-html40">
            <head>
                <meta charset="UTF-8">
                <!--[if gte mso 9]>
                <xml>
                    <x:ExcelWorkbook>
                        <x:ExcelWorksheets>
                            <x:ExcelWorksheet>
                                <x:Name>Rejected Items</x:Name>
                                <x:WorksheetOptions>
                                    <x:DisplayGridlines/>
                                    <x:FreezePanes/>
                                    <x:FrozenNoSplit/>
                                    <x:SplitHorizontal>5</x:SplitHorizontal>
                                    <x:TopRowBottomPane>5</x:TopRowBottomPane>
                                    <x:ActivePane>2</x:ActivePane>
                                </x:WorksheetOptions>
                            </x:ExcelWorksheet>
                        </x:ExcelWorksheets>
                    </x:ExcelWorkbook>
                </xml>
                <![endif]-->
                <style>
                    table { border-collapse: collapse; font-family: 'Segoe UI', Tahoma, Arial, sans-serif; }
                    td, th { mso-number-format:'\\@'; }
                </style>
            </head>
            <body>
                <table border="0" cellspacing="0" cellpadding="0">
                    <colgroup>${colgroupHtml}</colgroup>
                    <thead>
                        ${titleBar}
                        ${metaRow('Department', dept)}
                        ${metaRow('Period', period)}
                        ${metaRow('Owner', owner)}
                        ${metaRow('Generated', generatedAt)}
                        <tr><td colspan="7" style="height:6px;border:none;"></td></tr>
                        <tr>${headerHtml}</tr>
                    </thead>
                    <tbody>
                        ${rowsHtml}
                    </tbody>
                </table>
            </body>
        </html>`;

        const blob = new Blob(['\ufeff' + html], { type: 'application/vnd.ms-excel;charset=utf-8' });
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        const filename = `rejected_items_${dept}_${log.year}-${log.month}.xls`;
        link.href = url;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);
    };

    if (!log) return null;

    return (
        <Dialog open={open} onClose={onClose} maxWidth="lg" fullWidth>
            <DialogTitle sx={{ display: 'flex', alignItems: 'center', gap: 1, bgcolor: '#043478', color: 'white' }}>
                <VisibilityIcon />
                <Typography variant="h6">Rejected Items Details</Typography>
            </DialogTitle>
            <DialogContent sx={{ p: 0 }}>
                <Box sx={{ p: 2, display: 'flex', justifyContent: 'space-between', bgcolor: '#f5f5f5' }}>
                    <Box>
                        <Typography variant="subtitle2" color="textSecondary">Department</Typography>
                        <Typography variant="body1" fontWeight="bold">{log.department_code || log.department}</Typography>
                    </Box>
                    <Box>
                        <Typography variant="subtitle2" color="textSecondary">Period</Typography>
                        <Typography variant="body1" fontWeight="bold">{log.month}/{log.year}</Typography>
                    </Box>
                    <Box>
                        <Typography variant="subtitle2" color="textSecondary">Owner</Typography>
                        <Typography variant="body1" fontWeight="bold">{log.owner_name}</Typography>
                    </Box>
                </Box>

                {loading ? (
                    <Box sx={{ p: 5, textAlign: 'center' }}>
                        <CircularProgress />
                    </Box>
                ) : (
                    <TableContainer sx={{ maxHeight: '60vh' }}>
                        <Table stickyHeader size="small">
                            <TableHead>
                                <TableRow>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>GL Code</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Account Name</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Doc No.</TableCell>
                                    <TableCell align="right" sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Amount</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Vendor</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Description</TableCell>
                                    <TableCell sx={{ bgcolor: '#f5f5f5', fontWeight: 'bold' }}>Date</TableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                                {items.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={7} align="center" sx={{ py: 5 }}>
                                            No rejected items found or snapshots not available.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    items.map((row, index) => (
                                        <TableRow key={index} hover>
                                            <TableCell>{row.conso_gl}</TableCell>
                                            <TableCell>{row.gl_account_name}</TableCell>
                                            <TableCell>{row.doc_no}</TableCell>
                                            <TableCell align="right" sx={{ fontWeight: 'bold', color: parseFloat(row.amount) < 0 ? 'red' : 'green' }}>
                                                {parseFloat(row.amount || 0).toLocaleString(undefined, { minimumFractionDigits: 2 })}
                                            </TableCell>
                                            <TableCell>{row.vendor || "-"}</TableCell>
                                            <TableCell>{row.description}</TableCell>
                                            <TableCell>{row.posting_date}</TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </TableContainer>
                )}
            </DialogContent>
            <DialogActions sx={{ p: 2 }}>
                <Button
                    onClick={handleExportExcel}
                    variant="outlined"
                    color="success"
                    startIcon={<FileDownloadIcon />}
                    disabled={loading || items.length === 0}
                >
                    Export to Excel
                </Button>
                <Button onClick={onClose} variant="contained" color="primary">ปิดหน้าต่าง (Close)</Button>
            </DialogActions>
        </Dialog>
    );
};

export default AuditDetailsModal;
