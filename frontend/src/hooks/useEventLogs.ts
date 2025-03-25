import { useEffect, useState, useCallback } from 'react';
import { EventLog } from '../models/eventLog.model';
import { fetchEventLogs } from '../api/axiosConfig';
import { logger } from '../api/axiosConfig';

// Get polling interval from environment variable with fallback to 3000ms
const POLL_INTERVAL = parseInt(process.env.REACT_APP_POLL_INTERVAL || '3000', 10);

/**
 * Hook to fetch and manage event logs with automatic polling
 */
const useEventLogs = () => {
  const [eventLogs, setEventLogs] = useState<EventLog[]>([]);

  // Use callback to avoid recreating this function on every render
  const fetchLogs = useCallback(async () => {
    try {
      const logs = await fetchEventLogs();
      logger.debug(`Fetched ${logs.length} event logs`);
      setEventLogs(logs);
    } catch (error) {
      logger.error("Error fetching event logs:", error);
    }
  }, []);

  useEffect(() => {
    logger.info('Setting up event logs polling');
    
    // Initial fetch
    fetchLogs();

    // Set up polling interval
    const interval = setInterval(fetchLogs, POLL_INTERVAL);

    // Clean up interval on unmount
    return () => {
      logger.info('Cleaning up event logs polling');
      clearInterval(interval);
    };
  }, [fetchLogs]);

  return eventLogs;
};

export default useEventLogs;
