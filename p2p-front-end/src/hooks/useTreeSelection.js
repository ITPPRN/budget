import { useState, useMemo, useCallback } from 'react';

/**
 * Custom Hook for managing tree selection state with performance optimization.
 * 
 * @param {Array} data - The recursive tree data structure.
 * @returns {Object} - { selectedLeaves, toggleNode, getNodeState, clearSelection, selectAll }
 */
export const useTreeSelection = (data) => {
    const [selectedLeaves, setSelectedLeaves] = useState(new Set());

    // 1. Indexing: Pre-calculate a map of NodeID -> All Descendant Leaf IDs
    // This runs once on mount (or when data changes), making subsequent lookups O(1).
    const nodeLeafMap = useMemo(() => {
        const map = new Map();

        const processNode = (node) => {
            let leaves = [];
            if (!node.children || node.children.length === 0) {
                // It's a leaf
                leaves.push(node.id);
            } else {
                // It's a parent
                node.children.forEach(child => {
                    leaves = leaves.concat(processNode(child));
                });
            }
            map.set(node.id, leaves);
            return leaves;
        };

        data.forEach(rootNode => processNode(rootNode));
        return map;
    }, [data]);

    // 2. Toggle Logic
    const toggleNode = useCallback((nodeId, shouldSelect) => {
        const targetLeaves = nodeLeafMap.get(nodeId) || [];

        setSelectedLeaves(prev => {
            const next = new Set(prev);
            if (shouldSelect) {
                targetLeaves.forEach(id => next.add(id));
            } else {
                targetLeaves.forEach(id => next.delete(id));
            }
            return next;
        });
    }, [nodeLeafMap]);

    // 3. Get Node State (checked, indeterminate)
    const getNodeState = useCallback((nodeId) => {
        const targetLeaves = nodeLeafMap.get(nodeId) || [];
        if (targetLeaves.length === 0) return { checked: false, indeterminate: false };

        // Count how many of this node's leaves are selected
        let count = 0;
        // optimization: loop through targetLeaves and check existence in Set is O(N) where N is leaves count.
        // Set lookup is O(1).
        for (const id of targetLeaves) {
            if (selectedLeaves.has(id)) count++;
        }

        const checked = count === targetLeaves.length;
        const indeterminate = count > 0 && count < targetLeaves.length;

        return { checked, indeterminate };
    }, [nodeLeafMap, selectedLeaves]);

    // 4. Utilities
    const clearSelection = useCallback(() => setSelectedLeaves(new Set()), []);

    // Optional: Select All
    const selectAll = useCallback(() => {
        // Collect ALL leaves from the map values? Or just top level nodes?
        // Easiest is to iterate all values in map that are leaves?
        // Or just iterate root nodes from data.
        // But creating a Set of *all* leaves is straightforward if we collected them.
        // For now, clear is user requirement, select all is extra.
    }, []);

    return {
        selectedLeaves,
        toggleNode,
        getNodeState,
        clearSelection,
        nodeLeafMap // exposed for debugging if needed
    };
};
