import api from './api/axiosInstance';
import { toast } from 'react-toastify';

/**
 * Utility function to handle Excel file downloads from the backend.
 * @param {string} endpoint - The API endpoint to call (e.g., '/budgets/export-budget-detail')
 * @param {Object} payload - The filter payload to send in the POST request
 * @param {string} defaultFilename - Fallback filename if not provided by the server
 */
export const downloadExcelFile = async (endpoint, payload, defaultFilename = 'export.xlsx') => {
    try {
        const response = await api.post(endpoint, payload, {
            responseType: 'blob', // Important: expect binary data
        });

        // 1. Extract filename from Content-Disposition header (if present)
        let filename = defaultFilename;
        const disposition = response.headers['content-disposition'];
        if (disposition && disposition.indexOf('attachment') !== -1) {
            const filenameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/;
            const matches = filenameRegex.exec(disposition);
            if (matches != null && matches[1]) {
                filename = matches[1].replace(/['"]/g, '');
            }
        }

        // 2. Create a Blob from the response data
        const blob = new Blob([response.data], { 
            type: response.headers['content-type'] || 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' 
        });

        // 3. Create a temporary URL and trigger the download
        const downloadUrl = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        
        // 4. Cleanup
        document.body.removeChild(link);
        window.URL.revokeObjectURL(downloadUrl);

        toast.success(`Download started: ${filename}`);
        
    } catch (error) {
        console.error("Download Excel Error:", error);
        
        // If the backend returned JSON error inside a blob, we need to parse it
        if (error.response && error.response.data && error.response.data.type === 'application/json') {
            const reader = new FileReader();
            reader.onload = () => {
                try {
                    const jsonError = JSON.parse(reader.result);
                    toast.error(jsonError.error || "Failed to download file");
                } catch(e) {
                    toast.error("Failed to download file");
                }
            };
            reader.readAsText(error.response.data);
        } else {
             toast.error("Failed to process download request. Please check your connection.");
        }
    }
};
