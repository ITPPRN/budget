import React from 'react';
import { Box, Paper, Typography, Button, Divider, List, TextField, InputAdornment } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import { useBudget } from '../../contexts/BudgetContext';
import FilterNode from './FilterNode';

const FilterPane = () => {
    const { filterOptions, selectedLeaves, clearSelection } = useBudget();
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
                    // If matches or has matching children, keep it.
                    // If name matches, we technically could show all children, but 
                    // usually showing just the matching trail is cleaner. 
                    // Let's decide: If name matches, show all children? 
                    // For now: Show minimal path. matches OR children match.
                    // Actually, if I search "Admin", I want to see content of Admin.
                    // So if nameMatch, maybe keep original children? 
                    // Let's stick to strict filter for now: only show things that match or contain matches.

                    // Re-construct node with filtered children
                    acc.push({
                        ...node,
                        children: nameMatch ? node.children : filteredChildren
                        // Logic above: If I match, keep my original children (showing everything inside). 
                        // If I don't match, keep me only if I have matching descendants.
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
            <Paper sx={{ p: 2, height: '100%', display: 'flex', flexDirection: 'column', borderRadius: 2, overflow: 'hidden' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2, flexShrink: 0 }}>
                    <Typography variant="h6" sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    Filter Pane
                    </Typography>
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

                {/* Search Input */}
                <Box sx={{ mb: 2 }}>
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

                <Divider sx={{ mb: 2 }} />

                <Box sx={{ flexGrow: 1, overflowY: 'auto' }}>
                    <List component="nav" dense>
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
            </Paper>
        </Box>
    );
};

export default FilterPane;
