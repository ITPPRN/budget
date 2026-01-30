import React from 'react';
import { Box, Paper, Typography, Button, Divider, List } from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import { useBudget } from '../../contexts/BudgetContext';
import FilterNode from './FilterNode';

const FilterPane = () => {
    const { filterOptions, selectedLeaves, clearSelection } = useBudget();

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
                        <SearchIcon /> Filter Pane
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
                <Divider sx={{ mb: 2 }} />

                <Box sx={{ flexGrow: 1, overflowY: 'auto' }}>
                    <List component="nav" dense>
                        {filterOptions.map(node => (
                            <FilterNode key={node.id} node={node} />
                        ))}
                    </List>
                    {filterOptions.length === 0 && (
                        <Typography color="text.secondary" align="center" sx={{ mt: 4 }}>
                            No filters available
                        </Typography>
                    )}
                </Box>
            </Paper>
        </Box>
    );
};

export default FilterPane;
