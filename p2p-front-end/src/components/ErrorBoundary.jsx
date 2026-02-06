import React from 'react';
import { Box, Typography } from '@mui/material';

class ErrorBoundary extends React.Component {
    constructor(props) {
        super(props);
        this.state = { hasError: false, error: null, errorInfo: null };
    }

    static getDerivedStateFromError(error) {
        return { hasError: true };
    }

    componentDidCatch(error, errorInfo) {
        this.setState({ error, errorInfo });
        console.error("Uncaught error:", error, errorInfo);
    }

    render() {
        if (this.state.hasError) {
            return (
                <Box sx={{ p: 4, textAlign: 'center' }}>
                    <Typography variant="h4" color="error" gutterBottom>
                        Something went wrong.
                    </Typography>
                    <Typography variant="body1" color="text.secondary">
                        {this.state.error && this.state.error.toString()}
                    </Typography>
                    <details style={{ whiteSpace: 'pre-wrap', marginTop: 20, textAlign: 'left' }}>
                        {this.state.errorInfo && this.state.errorInfo.componentStack}
                    </details>
                </Box>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
