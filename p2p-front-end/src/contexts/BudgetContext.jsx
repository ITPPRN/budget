import React, { createContext, useContext, useState, useEffect } from 'react';
import api from '../utils/api/axiosInstance';
import { useTreeSelection } from '../hooks/useTreeSelection';

const BudgetContext = createContext();

export const BudgetProvider = ({ children }) => {
    const [filterOptions, setFilterOptions] = useState([]);
    const [isLoading, setIsLoading] = useState(true);

    // Fetch data inside the provider
    useEffect(() => {
        const fetchStructure = async () => {
            try {
                const response = await api.get('/budgets/budget-structure');
                if (response.data) {
                    setFilterOptions(response.data);
                }
            } catch (error) {
                console.error("Failed to fetch budget structure", error);
            } finally {
                setIsLoading(false);
            }
        };
        fetchStructure();
    }, []);

    // Use the custom hook for selection logic using the dynamic data
    const {
        selectedLeaves,
        toggleNode,
        getNodeState,
        clearSelection,
        getAllLeafIds
    } = useTreeSelection(filterOptions);

    const value = {
        filterOptions,
        isLoading,
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
