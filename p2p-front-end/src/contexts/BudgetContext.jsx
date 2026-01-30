import React, { createContext, useContext, useState } from 'react';
import { STATIC_FILTER_OPTIONS } from '../constants/budgetData';
import { useTreeSelection } from '../hooks/useTreeSelection';

const BudgetContext = createContext();

export const BudgetProvider = ({ children }) => {
    // Use the custom hook for selection logic using the static data
    const {
        selectedLeaves,
        toggleNode,
        getNodeState,
        clearSelection,
        getAllLeafIds
    } = useTreeSelection(STATIC_FILTER_OPTIONS);

    // We can also manage data fetching state here if we want to share it,
    // but for now, let's keep it focused on Selection & Filter Options.

    const value = {
        filterOptions: STATIC_FILTER_OPTIONS,
        selectedLeaves,
        toggleNode,
        getNodeState,
        clearSelection,
        getAllLeafIds,
    };

    return (
        <BudgetContext.Provider value={value}>
            {children}
        </BudgetContext.Provider>
    );
};

// Custom hook to use the context
export const useBudget = () => {
    const context = useContext(BudgetContext);
    if (!context) {
        throw new Error('useBudget must be used within a BudgetProvider');
    }
    return context;
};
