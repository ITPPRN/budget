import React from 'react';
import { Box, Paper, Typography, Button, Divider, List, TextField, InputAdornment } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import { useBudget } from '../../contexts/BudgetContext';
import FilterNode from './FilterNode';

const FilterPane = ({ compact = false }) => {
    const { filterOptions, selectedLeaves, clearSelection, isLoading } = useBudget();
    const [searchTerm, setSearchTerm] = React.useState('');

    // Recursive Filter Logic
    const filteredTree = React.useMemo(() => {
        if (!searchTerm) return filterOptions;

        const lowerTerm = searchTerm.toLowerCase();

        const filterNodes = (nodes) => {
            return nodes.reduce((acc, node) => {
                const nameMatch = node.name.toLowerCase().includes(lowerTerm);

                // If it has children, filter them first
                const filteredChildren = node.children ? filterNodes(node.children) : [];

                if (nameMatch || filteredChildren.length > 0) {
                    acc.push({
                        ...node,
                        children: nameMatch ? node.children : filteredChildren
                    });
                }

                return acc;
            }, []);
        };

        return filterNodes(filterOptions);
    }, [filterOptions, searchTerm]);

    return (
        <Box sx={{
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
            height: '100%'
        }}>
            <Paper sx={{
                p: compact ? 1 : 2,
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                borderRadius: compact ? 0 : 2,
                overflow: 'hidden',
                boxShadow: compact ? 'none' : 1
            }}>
                {/* Header: In compact mode, hide "Filter Pane" text, but keep Clear button */}
                <Box sx={{ display: 'flex', justifyContent: compact ? 'flex-end' : 'space-between', alignItems: 'center', mb: compact ? 1 : 2, flexShrink: 0 }}>
                    {!compact && (
                        <Typography variant="h6" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            Filter Pane
                        </Typography>
                    )}
                    <Button
                        variant="outlined"
                        color="error"
                        size="small"
                        onClick={clearSelection}
                        disabled={selectedLeaves.size === 0}
                    >
                        Clear
                    </Button>
                </Box>

                {/* Search Input (Always visible) */}
                <Box sx={{ mb: compact ? 1 : 2 }}>
                    <TextField
                        fullWidth
                        size="small"
                        placeholder="Search filters..."
                        variant="outlined"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        InputProps={{
                            startAdornment: (
                                <InputAdornment position="start">
                                    <SearchIcon color="action" fontSize="small" />
                                </InputAdornment>
                            ),
                        }}
                    />
                </Box>

                <Divider sx={{ mb: compact ? 1 : 2 }} />

                {isLoading ? (
                    <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                        <Typography color="text.secondary">Loading filters...</Typography>
                    </Box>
                ) : (
                    <Box sx={{ flexGrow: 1, overflowY: 'auto' }}>
                        <List component="nav" dense disablePadding={compact}>
                            {filteredTree.map(node => (
                                <FilterNode key={node.id} node={node} />
                            ))}
                        </List>
                        {filteredTree.length === 0 && (
                            <Typography color="text.secondary" align="center" sx={{ mt: 4 }}>
                                No results found
                            </Typography>
                        )}
                    </Box>
                )}
            </Paper>
        </Box>
    );
};

export default FilterPane;
