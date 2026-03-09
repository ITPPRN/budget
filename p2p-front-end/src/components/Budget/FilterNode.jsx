import React, { useState } from 'react';
import {
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Checkbox,
    Collapse,
    List,
    IconButton,
    Typography,
    Box
} from '@mui/material';
import ExpandLess from '@mui/icons-material/ExpandLess';
import ExpandMore from '@mui/icons-material/ExpandMore';
import { useBudget } from '../../contexts/BudgetContext';

const FilterNode = React.memo(({ node }) => {
    const [open, setOpen] = useState(false);
    const { getNodeState, toggleNode } = useBudget();

    // Get calculated state from Context/Hook (O(1) lookup effectively)
    const { checked, indeterminate } = getNodeState(node.id);

    const handleExpandClick = (e) => {
        e.stopPropagation();
        setOpen(!open);
    };

    const handleCheckboxClick = (e) => {
        e.stopPropagation();
        // Toggle logic: if not fully checked, select all. If fully checked, deselect all.
        toggleNode(node.id, !checked);
    };

    const hasChildren = node.children && node.children.length > 0;

    return (
        <Box sx={{ position: 'relative' }}>
            <ListItem
                disablePadding
                secondaryAction={
                    hasChildren ? (
                        <IconButton edge="end" size="small" onClick={handleExpandClick}>
                            {open ? <ExpandLess /> : <ExpandMore />}
                        </IconButton>
                    ) : null
                }
                sx={{
                    pl: Math.max(0, ((node.level || 1) - 1) * 3),
                    backgroundColor: node.level === 1 ? 'rgba(0,0,0,0.02)' : 'transparent',
                    borderBottom: node.level === 1 ? '1px solid rgba(0,0,0,0.05)' : 'none',
                }}
            >
                {/* Vertical hierarchy line */}
                {node.level > 1 && (
                    <Box sx={{
                        position: 'absolute',
                        left: (node.level - 2) * 24 + 16,
                        top: 0,
                        bottom: 0,
                        width: '1px',
                        backgroundColor: 'rgba(0,0,0,0.1)',
                        zIndex: 0
                    }} />
                )}

                <ListItemButton
                    dense
                    onClick={hasChildren ? handleExpandClick : handleCheckboxClick}
                    sx={{
                        py: 0.25,
                        borderRadius: 1,
                        my: 0.1,
                        ml: node.level > 1 ? 0.5 : 0,
                        '&:hover': {
                            backgroundColor: 'rgba(0,0,0,0.03)'
                        }
                    }}
                >
                    <ListItemIcon sx={{ minWidth: 32, zIndex: 1 }}>
                        <Checkbox
                            edge="start"
                            checked={checked}
                            indeterminate={indeterminate}
                            tabIndex={-1}
                            disableRipple
                            size="small"
                            onClick={handleCheckboxClick}
                        />
                    </ListItemIcon>
                    <ListItemText
                        primary={
                            <Typography
                                variant="body2"
                                sx={{
                                    fontWeight: hasChildren ? 600 : 400,
                                    fontSize: node.level === 1 ? '0.875rem' : '0.8125rem',
                                    color: node.level === 4 ? 'text.secondary' : 'text.primary',
                                }}
                            >
                                {node.name || 'Unknown'}
                            </Typography>
                        }
                    />
                </ListItemButton>
            </ListItem>

            {hasChildren && (
                <Collapse in={open} timeout="auto" unmountOnExit>
                    <List component="div" disablePadding>
                        {node.children.map((child) => (
                            <FilterNode
                                key={child.id || child.name}
                                node={child}
                            />
                        ))}
                    </List>
                </Collapse>
            )}
        </Box>
    );
});

export default FilterNode;
