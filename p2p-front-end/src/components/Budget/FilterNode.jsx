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
    Typography
} from '@mui/material';
import ExpandLess from '@mui/icons-material/ExpandLess';
import ExpandMore from '@mui/icons-material/ExpandMore';
import { useBudget } from '../../contexts/BudgetContext';

const FilterNode = ({ node }) => {
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
        <>
            <ListItem
                disablePadding
                secondaryAction={
                    hasChildren ? (
                        <IconButton edge="end" size="small" onClick={handleExpandClick}>
                            {open ? <ExpandLess /> : <ExpandMore />}
                        </IconButton>
                    ) : null
                }
                sx={{ pl: (node.level - 1) * 2 }} // Indent based on level
            >
                <ListItemButton dense onClick={hasChildren ? handleExpandClick : handleCheckboxClick}>
                    <ListItemIcon sx={{ minWidth: 32 }}>
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
                            <Typography variant="body2" sx={{ fontWeight: open ? 'bold' : 'normal' }}>
                                {node.name}
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
        </>
    );
};

export default FilterNode;
